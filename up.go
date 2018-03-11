// up is the ultimate pipe composer/editor. It helps building Linux pipelines
// in a terminal-based UI interactively, with live preview of command results.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-isatty"
)

func main() {
	// TODO: Without below block, we'd hang with no piped input (see github.com/peco/peco, mattn/gof, fzf, etc.)
	if isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintln(os.Stderr, "error: up requires some data piped on standard input, e.g.: `echo hello world | up`")
		os.Exit(1)
	}

	// Init TUI code
	// TODO: maybe try gocui or tcell?
	tui, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	err = tui.Init()
	if err != nil {
		panic(err)
	}
	defer tui.Fini()

	var (
		editor      = NewEditor("| ")
		lastCommand = ""
		subprocess  *Subprocess
		inputBuf    = NewBuf()
		buf         = inputBuf
		bufStyle    = BufDrawing{}
		bufY        = 1
	)

	// In background, start collecting input from stdin to internal buffer of size 40 MB, then pause it
	go inputBuf.Collect(os.Stdin, func() {
		tui.PostEvent(tcell.NewEventInterrupt(nil))
	})

	// Main loop
main_loop:
	for {
		// Run command automatically in background if user edited it (and kill previous command)
		// TODO: allow stopping/restarting this behavior via Ctrl-Enter
		// TODO: allow stopping this behavior via Ctrl-C (and killing current command), but invent some nice way to quit then
		command := editor.String()
		if command != lastCommand {
			lastCommand = command
			subprocess.Kill()
			if command != "" {
				subprocess = StartSubprocess(inputBuf, command, func() {
					tui.PostEvent(tcell.NewEventInterrupt(nil))
				})
				buf = subprocess.Buf
			} else {
				// If command is empty, show original input data again (~ equivalent of typing `cat`)
				subprocess = nil
				buf = inputBuf
			}
		}

		// Draw command input line
		editor.Draw(tui, 0, 0, true)
		buf.Draw(tui, bufY, bufStyle)
		tui.Show()

		// Handle events
		// TODO: how to interject with timer events triggering refresh?
		switch ev := tui.PollEvent().(type) {
		case *tcell.EventKey:
			// handle command-line editing keys
			if editor.HandleKey(ev) {
				continue main_loop
			}
			// handle other keys
			switch ev.Key() {
			case tcell.KeyCtrlC:
				// quit
				return
			// TODO: move buf scroll handlers to Buf or BufDrawing struct
			case tcell.KeyUp:
				bufStyle.Y--
				bufStyle.NormalizeY(buf.Lines())
			case tcell.KeyDown:
				bufStyle.Y++
				bufStyle.NormalizeY(buf.Lines())
			case tcell.KeyPgDn:
				// TODO: in top-right corner of Buf area, draw current line number & total # of lines
				_, h := tui.Size()
				bufStyle.Y += h - bufY - 1
				bufStyle.NormalizeY(buf.Lines())
			case tcell.KeyPgUp:
				_, h := tui.Size()
				bufStyle.Y -= h - bufY - 1
				bufStyle.NormalizeY(buf.Lines())
			}
		}
	}

	// TODO: [LATER] Ctrl-O shows input via `less` or $PAGER
	// TODO: ^X - save into executable file upN.sh (with #!/bin/bash) and quit
	// TODO: properly show all licenses of dependencies on --version
	// TODO: [LATER] allow increasing size of input buffer with some key
	// TODO: [LATER] on ^X, leave TUI and run the command through buffered input, then unpause rest of input
	// TODO: [LATER] allow adding more elements of pipeline (initially, just writing `foo | bar` should work)
	// TODO: [LATER] allow invocation with partial command, like: `up grep -i`
	// TODO: [LATER][MAYBE] allow reading upN.sh scripts
	// TODO: [LATER] auto-save and/or save on Ctrl-S or something
	// TODO: [MUCH LATER] readline-like rich editing support? and completion?
	// TODO: [MUCH LATER] integration with fzf? and pindexis/marker?
	// TODO: [LATER] forking and unforking pipelines
	// TODO: [LATER] capture output of a running process (see: https://stackoverflow.com/q/19584825/98528)
	// TODO: [LATER] richer TUI:
	// - show # of read lines & kbytes
	// - show status (errorlevel) of process, or that it's still running (also with background colors)
	// - allow copying and pasting to/from command line
	// TODO: [LATER] allow connecting external editor (become server/engine via e.g. socket)
	// TODO: [LATER] become pluggable into http://luna-lang.org
	// TODO: [LATER][MAYBE] allow "plugins" ("combos" - commands with default options) e.g. for Lua `lua -e`+auto-quote, etc.
	// TODO: [LATER] make it more friendly to infrequent Linux users by providing "descriptive" commands like "search" etc.
	// TODO: [LATER] advertise on: HN, r/programming, r/golang, r/commandline, r/linux; data exploration? data science?
}

type Buf struct {
	bytes []byte
	// NOTE: n can be written only by Collect
	n     int
	nLock sync.Mutex
}

func NewBuf() *Buf {
	const bufsize = 40 * 1024 * 1024 // 40 MB
	return &Buf{bytes: make([]byte, bufsize)}
}

func (b *Buf) Collect(r io.Reader, signal func()) {
	// TODO: allow stopping - take context?
	for {
		n, err := r.Read(b.bytes[b.n:])
		b.nLock.Lock()
		b.n += n
		b.nLock.Unlock()
		go signal()
		if err == io.EOF {
			// TODO: mark work as complete
			return
		} else if err != nil {
			// TODO: better handling of errors
			panic(err)
		}
		if b.n == len(b.bytes) {
			return
		}
	}
}

func (b *Buf) Draw(tui tcell.Screen, y0 int, style BufDrawing) {
	b.nLock.Lock()
	buf := b.bytes[:b.n]
	b.nLock.Unlock()

	// PgDn/PgUp etc. support
	for ; style.Y > 0; style.Y-- {
		newline := bytes.IndexByte(buf, '\n')
		if newline != -1 {
			buf = buf[newline+1:]
		}
	}

	w, h := tui.Size()
	// TODO: handle runes properly, including their visual width (mattn/go-runewidth)
	x, y := 0, y0
	for len(buf) > 0 && y < h {
		ch, sz := utf8.DecodeRune(buf)
		buf = buf[sz:]
		switch ch {
		case '\n':
			b.endline(tui, x, y, w)
			x, y = 0, y+1
			continue
		case '\t':
			const tabwidth = 8
			b.putch(tui, x, y, ' ')
			for x%tabwidth < (tabwidth - 1) {
				x++
				if x >= w {
					break
				}
				b.putch(tui, x, y, ' ')
			}
		default:
			b.putch(tui, x, y, ch)
		}
		x++
		if x > w {
			// x, y = 0, y+1
			b.putch(tui, w-1, y, '»') // TODO: also «
		}
	}
	for ; y < h; y++ {
		b.endline(tui, 0, y, w)
	}
}

func (b *Buf) putch(tui tcell.Screen, x, y int, ch rune) {
	tui.SetCell(x, y, tcell.StyleDefault, ch)
}

func (b *Buf) endline(tui tcell.Screen, x, y, screenw int) {
	for ; x < screenw; x++ {
		b.putch(tui, x, y, ' ')
	}
}

func (b *Buf) Lines() int {
	b.nLock.Lock()
	n := b.n
	b.nLock.Unlock()
	newlines := bytes.Count(b.bytes[:n], []byte{'\n'})
	return newlines + 1
}

func (b *Buf) NewReader() io.Reader {
	// TODO: return EOF if input is fully buffered?
	i := 0
	return funcReader(func(p []byte) (n int, err error) {
		b.nLock.Lock()
		end := b.n
		b.nLock.Unlock()
		n = copy(p, b.bytes[i:end])
		i += n
		if n == 0 {
			// FIXME: GROSS HACK! To avoid busy-wait in caller, don't return
			// (0,nil), instead wait until at least 1 available, or return (0,
			// io.EOF) on completion
			time.Sleep(100 * time.Millisecond)
		}
		return n, nil
	})
}

type funcReader func([]byte) (int, error)

func (f funcReader) Read(p []byte) (int, error) { return f(p) }

type Editor struct {
	prompt []rune
	// TODO: make editor multiline. Reuse gocui or something for this?
	// TODO: rename 'command' to 'data' or 'value' or something more generic
	command []rune
	cursor  int
	// lastw is length of command on last Draw
	lastw int
}

func NewEditor(prompt string) *Editor {
	return &Editor{prompt: []rune(prompt)}
}

func (e *Editor) String() string {
	return string(e.command)
}

func (e *Editor) Draw(tui tcell.Screen, x, y int, setcursor bool) {
	promptStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
	for i, ch := range e.prompt {
		tui.SetCell(x+i, y, promptStyle, ch)
	}
	for i, ch := range e.command {
		tui.SetCell(x+len(e.prompt)+i, y, promptStyle, ch)
	}
	// clear remains of last command if needed
	for i := len(e.command); i < e.lastw; i++ {
		tui.SetCell(x+len(e.prompt)+i, y, tcell.StyleDefault, ' ')
	}
	if setcursor {
		tui.ShowCursor(x+len(e.prompt)+e.cursor, y)
	}
	e.lastw = len(e.command)
}

func (e *Editor) HandleKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune {
		e.insert(ev.Rune())
		return true
	}
	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// See https://github.com/nsf/termbox-go/issues/145
		e.delete(-1)
	case tcell.KeyDelete:
		e.delete(0)
	case tcell.KeyLeft:
		if e.cursor > 0 {
			e.cursor--
		}
	case tcell.KeyRight:
		if e.cursor < len(e.command) {
			e.cursor++
		}
	default:
		return false
	}
	return true
}

func (e *Editor) insert(ch rune) {
	// insert key into command (https://github.com/golang/go/wiki/SliceTricks#insert)
	e.command = append(e.command, 0)
	copy(e.command[e.cursor+1:], e.command[e.cursor:])
	e.command[e.cursor] = ch
	e.cursor++
}

func (e *Editor) delete(dx int) {
	pos := e.cursor + dx
	if pos < 0 || pos >= len(e.command) {
		return
	}
	e.command = append(e.command[:pos], e.command[pos+1:]...)
	e.cursor = pos
}

type Subprocess struct {
	Buf    *Buf
	cancel context.CancelFunc
}

func StartSubprocess(inputBuf *Buf, command string, signal func()) *Subprocess {
	ctx, cancel := context.WithCancel(context.TODO())
	s := &Subprocess{
		Buf:    NewBuf(),
		cancel: cancel,
	}
	r, w := io.Pipe()
	go s.Buf.Collect(r, signal)

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Stdin = inputBuf.NewReader()
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "up: %s", err)
		return s
	}
	go cmd.Wait()
	return s
}

func (s *Subprocess) Kill() {
	if s == nil {
		return
	}
	s.cancel()
}

type BufDrawing struct {
	// TODO: Wrap bool
	Y int // for pgup/pgdn scrolling)
	// TODO: X int (for left<->right scrolling)
}

func (b *BufDrawing) NormalizeY(nlines int) {
	if b.Y >= nlines {
		b.Y = nlines - 1
	}
	if b.Y < 0 {
		b.Y = 0
	}
}
