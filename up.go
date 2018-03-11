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
	const (
		xscroll = 8
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
		switch ev := tui.PollEvent().(type) {
		case *tcell.EventKey:
			// handle command-line editing keys
			if editor.HandleKey(ev) {
				continue main_loop
			}
			// handle other keys
			switch (keymod{ev.Key(), ev.Modifiers()}) {
			case keymod{tcell.KeyCtrlC, 0},
				keymod{tcell.KeyCtrlC, tcell.ModCtrl}:
				// quit
				return
			case keymod{tcell.KeyCtrlX, 0},
				keymod{tcell.KeyCtrlX, tcell.ModCtrl}:
				// write script and quit
				writeScript(editor.String(), tui)
				return
			// TODO: move buf scroll handlers to Buf or BufDrawing struct
			case keymod{tcell.KeyUp, 0}:
				bufStyle.Y--
				bufStyle.NormalizeY(buf.Lines())
			case keymod{tcell.KeyDown, 0}:
				bufStyle.Y++
				bufStyle.NormalizeY(buf.Lines())
			case keymod{tcell.KeyPgDn, 0}:
				// TODO: in top-right corner of Buf area, draw current line number & total # of lines
				_, h := tui.Size()
				bufStyle.Y += h - bufY - 1
				bufStyle.NormalizeY(buf.Lines())
			case keymod{tcell.KeyPgUp, 0}:
				_, h := tui.Size()
				bufStyle.Y -= h - bufY - 1
				bufStyle.NormalizeY(buf.Lines())
			case keymod{tcell.KeyLeft, tcell.ModAlt},
				keymod{tcell.KeyLeft, tcell.ModCtrl}:
				bufStyle.X -= xscroll
				if bufStyle.X < 0 {
					bufStyle.X = 0
				}
			case keymod{tcell.KeyRight, tcell.ModAlt},
				keymod{tcell.KeyRight, tcell.ModCtrl}:
				bufStyle.X += xscroll
			case keymod{tcell.KeyHome, tcell.ModAlt},
				keymod{tcell.KeyHome, tcell.ModCtrl}:
				bufStyle.X = 0
			}
		}
	}

	// TODO: properly handle fully consumed buffers, to enable piping into `wc -l` or `uniq -c` etc.
	// TODO: [LATER] make it work on Windows; maybe with mattn/go-shellwords ?
	// TODO: [LATER] Ctrl-O shows input via `less` or $PAGER
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
	putch := func(x, y int, ch rune) {
		if x <= style.X && style.X != 0 {
			x, ch = 0, '«'
		} else {
			x -= style.X
		}
		if x >= w {
			x, ch = w-1, '»'
		}
		tui.SetCell(x, y, tcell.StyleDefault, ch)
	}
	endline := func(x, y int) {
		x -= style.X
		if x < 0 {
			x = 0
		}
		for ; x < w; x++ {
			tui.SetCell(x, y, tcell.StyleDefault, ' ')
		}
	}

	x, y := 0, y0
	// TODO: handle runes properly, including their visual width (mattn/go-runewidth)
	for len(buf) > 0 && y < h {
		ch, sz := utf8.DecodeRune(buf)
		buf = buf[sz:]
		switch ch {
		case '\n':
			endline(x, y)
			x, y = 0, y+1
			continue
		case '\t':
			const tabwidth = 8
			putch(x, y, ' ')
			for x%tabwidth < (tabwidth - 1) {
				x++
				if x >= w {
					break
				}
				putch(x, y, ' ')
			}
		default:
			putch(x, y, ch)
		}
		x++
	}
	for ; y < h; y++ {
		endline(0, y)
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
	if ev.Key() == tcell.KeyRune && ev.Modifiers()&(^tcell.ModShift) == 0 {
		e.insert(ev.Rune())
		return true
	}
	switch (keymod{ev.Key(), ev.Modifiers()}) {
	case keymod{tcell.KeyBackspace, 0}, keymod{tcell.KeyBackspace2, 0}:
		// See https://github.com/nsf/termbox-go/issues/145
		e.delete(-1)
	case keymod{tcell.KeyDelete, 0}:
		e.delete(0)
	case keymod{tcell.KeyLeft, 0}:
		if e.cursor > 0 {
			e.cursor--
		}
	case keymod{tcell.KeyRight, 0}:
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
	X int // for left<->right scrolling
}

func (b *BufDrawing) NormalizeY(nlines int) {
	if b.Y >= nlines {
		b.Y = nlines - 1
	}
	if b.Y < 0 {
		b.Y = 0
	}
}

type keymod struct {
	tcell.Key
	tcell.ModMask
}

func writeScript(command string, tui tcell.Screen) {
	var f *os.File
	var err error
	// TODO: if we hit loop end, panic with some message
	for i := 1; i < 1000; i++ {
		f, err = os.OpenFile(fmt.Sprintf("up%d.sh", i), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0755)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			// FIXME: don't panic, instead show error and let user try to copy & paste visually
			panic(err)
		} else {
			break
		}
	}
	_, err = fmt.Fprintf(f, "#!/bin/bash\n%s\n", command)
	if err != nil {
		// FIXME: don't panic, instead show error and let user try to copy & paste visually
		panic(err)
	}
	err = f.Close()
	if err != nil {
		// FIXME: don't panic, instead show error and let user try to copy & paste visually
		panic(err)
	}
	tui.Fini()
	fmt.Printf("up: command written to: %s\n", f.Name())
}
