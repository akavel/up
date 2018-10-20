// Copyright 2018 The up Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// up is the Ultimate Plumber, a tool for writing Linux pipes in a
// terminal-based UI interactively, with instant live preview of command
// results.
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-isatty"
)

// TODO: some key shortcut to increase stdin capture buffer size (unless EOF already reached)
// TODO: show status infos:
//  - red fg + "up: process returned with error code %d" -- when subprocess returned an error
//  - yellow fg -- when process is still not finished
// TODO: readme, asciinema
// TODO: on github: add issues, incl. up-for-grabs / help-wanted
// TODO: [LATER] make it work on Windows; maybe with mattn/go-shellwords ?
// TODO: [LATER] Ctrl-O shows input via `less` or $PAGER
// TODO: properly show all licenses of dependencies on --version
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
// TODO: [LATER] advertise on: HN, r/programming, r/golang, r/commandline, r/linux, up-for-grabs.net; data exploration? data science?

func main() {
	// Handle command-line flags
	parseFlags()

	// Initialize TUI infrastructure
	tui := initTUI()
	defer tui.Fini()

	// Initialize 3 main UI parts
	var (
		// The top line of the TUI is an editable command, which will be used
		// as a pipeline for data we read from stdin
		commandEditor = NewEditor("| ")
		// The rest of the screen is a view of the results of the command
		commandOutput = BufView{}
		// Sometimes, a message may be displayed at the bottom of the screen, with help or other info
		message = `^X exit (^C nosave)  PgUp/PgDn/Up/Dn/^</^> scroll  ^S pause (^Q end)  [Ultimate Plumber v0.1 by akavel]`
	)

	// Initialize main data flow
	var (
		// We capture data piped to 'up' on standard input into an internal buffer
		// When some new data shows up on stdin, we raise a custom signal,
		// so that main loop will refresh the buffers and the output.
		stdinCapture = NewBuf().StartCapturing(os.Stdin, func() { triggerRefresh(tui) })
		// Then, we pass this data as input to a subprocess.
		// Initially, no subprocess is running, as no command is entered yet
		commandSubprocess *Subprocess = nil
	)
	// Intially, for user's convenience, show the raw input data, as if `cat` command was typed
	commandOutput.Buf = stdinCapture

	// Main loop
	lastCommand := ""
	for {
		// If user edited the command, immediately run it in background, and
		// kill the previously running command.
		command := commandEditor.String()
		if command != lastCommand {
			commandSubprocess.Kill()
			if command != "" {
				commandSubprocess = StartSubprocess(command, stdinCapture, func() { triggerRefresh(tui) })
				commandOutput.Buf = commandSubprocess.Buf
			} else {
				// If command is empty, show original input data again (~ equivalent of typing `cat`)
				commandSubprocess = nil
				commandOutput.Buf = stdinCapture
			}
		}
		lastCommand = command

		// Draw UI
		w, h := tui.Size()
		stdinCapture.DrawStatus(TuiRegion(tui, 0, 0, 1, 1))
		commandEditor.DrawTo(TuiRegion(tui, 1, 0, w-1, 1),
			func(x, y int) { tui.ShowCursor(x+1, 0) })
		commandOutput.DrawTo(TuiRegion(tui, 0, 1, w, h-1))
		drawText(TuiRegion(tui, 0, h-1, w, 1), whiteOnBlue, message)
		tui.Show()

		// Handle UI events
		switch ev := tui.PollEvent().(type) {
		// Key pressed
		case *tcell.EventKey:
			// Is it a command editor key?
			if commandEditor.HandleKey(ev) {
				message = ""
				continue
			}
			// Is it a command output view key?
			if commandOutput.HandleKey(ev, h-1) {
				message = ""
				continue
			}
			// Some other global key combinations
			switch getKey(ev) {
			case key(tcell.KeyCtrlS),
				ctrlKey(tcell.KeyCtrlS):
				stdinCapture.Pause(true)
				triggerRefresh(tui)
			case key(tcell.KeyCtrlQ),
				ctrlKey(tcell.KeyCtrlQ):
				stdinCapture.Pause(false)
				lastCommand = ":" // Make sure we restart current command
			case key(tcell.KeyCtrlC),
				ctrlKey(tcell.KeyCtrlC),
				key(tcell.KeyCtrlD),
				ctrlKey(tcell.KeyCtrlD):
				// Quit
				// TODO: print the command in case user did this accidentally
				return
			case key(tcell.KeyCtrlX),
				ctrlKey(tcell.KeyCtrlX):
				// Write script 'upN.sh' and quit
				writeScript(commandEditor.String(), tui)
				return
			}
		}
	}
}

func parseFlags() {
	log.SetOutput(ioutil.Discard)
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		debug, err := os.Create("up.debug")
		if err != nil {
			die(err.Error())
		}
		log.SetOutput(debug)
	}
}

func initTUI() tcell.Screen {
	// TODO: Without below block, we'd hang when nothing is piped on input (see
	// github.com/peco/peco, mattn/gof, fzf, etc.)
	if isatty.IsTerminal(os.Stdin.Fd()) {
		die("up requires some data piped on standard input, for example try: `echo hello world | up`")
	}

	// Init TUI code
	// TODO: maybe try gocui or termbox?
	tui, err := tcell.NewScreen()
	if err != nil {
		die(err.Error())
	}
	err = tui.Init()
	if err != nil {
		die(err.Error())
	}
	return tui
}

func triggerRefresh(tui tcell.Screen) {
	tui.PostEvent(tcell.NewEventInterrupt(nil))
}

func die(message string) {
	os.Stderr.WriteString("error: " + message + "\n")
	os.Exit(1)
}

func NewEditor(prompt string) *Editor {
	return &Editor{prompt: []rune(prompt)}
}

type Editor struct {
	// TODO: make editor multiline. Reuse gocui or something for this?
	prompt []rune
	value  []rune
	cursor int
	// lastw is length of value on last Draw; we need it to know how much to erase after backspace
	lastw int
}

func (e *Editor) String() string { return string(e.value) }

func (e *Editor) DrawTo(region Region, setcursor func(x, y int)) {
	// Draw prompt & the edited value - use white letters on blue background
	style := whiteOnBlue
	for i, ch := range e.prompt {
		region.SetCell(i, 0, style, ch)
	}
	for i, ch := range e.value {
		region.SetCell(len(e.prompt)+i, 0, style, ch)
	}

	// Clear remains of last value if needed
	for i := len(e.value); i < e.lastw; i++ {
		region.SetCell(len(e.prompt)+i, 0, tcell.StyleDefault, ' ')
	}
	e.lastw = len(e.value)

	// Show cursor if requested
	if setcursor != nil {
		setcursor(len(e.prompt)+e.cursor, 0)
	}
}

func (e *Editor) HandleKey(ev *tcell.EventKey) bool {
	// If a character is entered, with no modifiers except maybe shift, then just insert it
	if ev.Key() == tcell.KeyRune && ev.Modifiers()&(^tcell.ModShift) == 0 {
		e.insert(ev.Rune())
		return true
	}
	// Handle editing & movement keys
	switch getKey(ev) {
	case key(tcell.KeyBackspace), key(tcell.KeyBackspace2):
		// See https://github.com/nsf/termbox-go/issues/145
		e.delete(-1)
	case key(tcell.KeyDelete):
		e.delete(0)
	case key(tcell.KeyLeft):
		if e.cursor > 0 {
			e.cursor--
		}
	case key(tcell.KeyRight):
		if e.cursor < len(e.value) {
			e.cursor++
		}
	default:
		// Unknown key/combination, not handled
		return false
	}
	return true
}

func (e *Editor) insert(ch rune) {
	// Insert character into value (https://github.com/golang/go/wiki/SliceTricks#insert)
	e.value = append(e.value, 0)
	copy(e.value[e.cursor+1:], e.value[e.cursor:])
	e.value[e.cursor] = ch
	e.cursor++
}

func (e *Editor) delete(dx int) {
	pos := e.cursor + dx
	if pos < 0 || pos >= len(e.value) {
		return
	}
	e.value = append(e.value[:pos], e.value[pos+1:]...)
	e.cursor = pos
}

type BufView struct {
	// TODO: Wrap bool
	Y   int // Y of the view in the Buf, for down/up scrolling
	X   int // X of the view in the Buf, for left/right scrolling
	Buf *Buf
}

func (v *BufView) DrawTo(region Region) {
	r := bufio.NewReader(v.Buf.NewReader(false))

	// PgDn/PgUp etc. support
	for y := v.Y; y > 0; y-- {
		line, err := r.ReadBytes('\n')
		switch err {
		case nil:
			// skip line
			continue
		case io.EOF:
			r = bufio.NewReader(bytes.NewReader(line))
			y = 0
			break
		default:
			panic(err)
		}
	}

	lclip := false
	drawch := func(x, y int, ch rune) {
		if x <= v.X && v.X != 0 {
			x, ch = 0, '«'
			lclip = true
		} else {
			x -= v.X
		}
		if x >= region.W {
			x, ch = region.W-1, '»'
		}
		region.SetCell(x, y, tcell.StyleDefault, ch)
	}
	endline := func(x, y int) {
		x -= v.X
		if x < 0 {
			x = 0
		}
		if x == 0 && lclip {
			x++
		}
		lclip = false
		for ; x < region.W; x++ {
			region.SetCell(x, y, tcell.StyleDefault, ' ')
		}
	}

	x, y := 0, 0
	// TODO: handle runes properly, including their visual width (mattn/go-runewidth)
	for {
		ch, _, err := r.ReadRune()
		if y >= region.H || err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		switch ch {
		case '\n':
			endline(x, y)
			x, y = 0, y+1
			continue
		case '\t':
			const tabwidth = 8
			drawch(x, y, ' ')
			for x%tabwidth < (tabwidth - 1) {
				x++
				if x >= region.W {
					break
				}
				drawch(x, y, ' ')
			}
		default:
			drawch(x, y, ch)
		}
		x++
	}
	for ; y < region.H; y++ {
		endline(0, y)
	}
}

func (v *BufView) HandleKey(ev *tcell.EventKey, scrollY int) bool {
	const scrollX = 8 // When user scrolls horizontally, move by this many characters
	switch getKey(ev) {
	//
	// Vertical scrolling
	//
	case key(tcell.KeyUp):
		v.Y--
		v.normalizeY()
	case key(tcell.KeyDown):
		v.Y++
		v.normalizeY()
	case key(tcell.KeyPgDn):
		// TODO: in top-right corner of Buf area, draw current line number & total # of lines
		v.Y += scrollY
		v.normalizeY()
	case key(tcell.KeyPgUp):
		v.Y -= scrollY
		v.normalizeY()
	//
	// Horizontal scrolling
	//
	case altKey(tcell.KeyLeft),
		ctrlKey(tcell.KeyLeft):
		v.X -= scrollX
		if v.X < 0 {
			v.X = 0
		}
	case altKey(tcell.KeyRight),
		ctrlKey(tcell.KeyRight):
		v.X += scrollX
	case altKey(tcell.KeyHome),
		ctrlKey(tcell.KeyHome):
		v.X = 0
	default:
		// Unknown key/combination, not handled
		return false
	}
	return true
}

func (v *BufView) normalizeY() {
	nlines := count(v.Buf.NewReader(false), '\n') + 1
	if v.Y >= nlines {
		v.Y = nlines - 1
	}
	if v.Y < 0 {
		v.Y = 0
	}
}

func count(r io.Reader, b byte) (n int) {
	buf := [256]byte{}
	for {
		i, err := r.Read(buf[:])
		n += bytes.Count(buf[:i], []byte{b})
		if err != nil {
			return
		}
	}
}

func NewBuf() *Buf {
	// TODO: make buffer size dynamic (growable by pressing a key)
	const bufsize = 40 * 1024 * 1024 // 40 MB
	buf := &Buf{bytes: make([]byte, bufsize)}
	buf.cond = sync.NewCond(&buf.mu)
	return buf
}

type Buf struct {
	bytes []byte

	mu     sync.Mutex // guards the following fields
	cond   *sync.Cond
	status bufStatus
	n      int
}

type bufStatus int

const (
	bufReading bufStatus = iota
	bufEOF
	bufPaused
)

func (b *Buf) StartCapturing(r io.Reader, notify func()) *Buf {
	go b.capture(r, notify)
	return b
}

func (b *Buf) capture(r io.Reader, notify func()) {
	// TODO: allow stopping - take context?
	for {
		n, err := r.Read(b.bytes[b.n:])

		b.mu.Lock()
		for b.status == bufPaused {
			b.cond.Wait()
		}
		b.n += n
		if err == io.EOF {
			b.status = bufEOF
		}
		if b.n == len(b.bytes) {
			// TODO: remove this when we can grow the buffer
			err = io.EOF
		}
		b.cond.Broadcast()
		b.mu.Unlock()

		go notify()
		if err == io.EOF {
			log.Printf("capture EOF after: %q", b.bytes[:b.n]) // TODO: make sure no race here, and skipped if not debugging
			return
		} else if err != nil {
			// TODO: better handling of errors
			panic(err)
		}
	}
}

func (b *Buf) Pause(pause bool) {
	b.mu.Lock()
	if pause {
		if b.status == bufReading {
			b.status = bufPaused
			// trigger all readers to emit fake EOF
			b.cond.Broadcast()
		}
	} else {
		if b.status == bufPaused {
			b.status = bufReading
			// wake up the capture func
			b.cond.Broadcast()
		}
	}
	b.mu.Unlock()
}

func (b *Buf) DrawStatus(region Region) {
	status := '~' // default: still reading input

	b.mu.Lock()
	switch {
	case b.status == bufPaused:
		status = '#'
	case b.status == bufEOF:
		status = ' ' // all input read, nothing more to do
	case b.n == len(b.bytes):
		status = '+' // buffer full
	}
	b.mu.Unlock()

	region.SetCell(0, 0, whiteOnBlue, status)
}

func (b *Buf) NewReader(blocking bool) io.Reader {
	i := 0
	return funcReader(func(p []byte) (n int, err error) {
		b.mu.Lock()
		end := b.n
		for blocking && end == i && b.status == bufReading && end < len(b.bytes) {
			b.cond.Wait()
			end = b.n
		}
		b.mu.Unlock()

		n = copy(p, b.bytes[i:end])
		i += n
		if n > 0 {
			return n, nil
		} else {
			if blocking {
				log.Printf("blocking reader emitting EOF after: %q", b.bytes[:end])
			}
			return 0, io.EOF
		}
	})
}

type funcReader func([]byte) (int, error)

func (f funcReader) Read(p []byte) (int, error) { return f(p) }

type Subprocess struct {
	Buf    *Buf
	cancel context.CancelFunc
}

func StartSubprocess(command string, stdin *Buf, notify func()) *Subprocess {
	ctx, cancel := context.WithCancel(context.TODO())
	r, w := io.Pipe()
	p := &Subprocess{
		Buf:    NewBuf().StartCapturing(r, notify),
		cancel: cancel,
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Stdin = stdin.NewReader(true)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(w, "up: %s", err)
		w.Close()
		return p
	}
	log.Println(cmd.Path)
	go func() {
		err = cmd.Wait()
		if err != nil {
			fmt.Fprintf(w, "up: %s", err)
		}
		w.Close()
	}()
	return p
}

func (s *Subprocess) Kill() {
	if s == nil {
		return
	}
	s.cancel()
}

type key int32

func getKey(ev *tcell.EventKey) key { return key(ev.Modifiers())<<16 + key(ev.Key()) }
func altKey(base tcell.Key) key     { return key(tcell.ModAlt)<<16 + key(base) }
func ctrlKey(base tcell.Key) key    { return key(tcell.ModCtrl)<<16 + key(base) }

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

type Region struct {
	W, H    int
	SetCell func(x, y int, style tcell.Style, ch rune)
}

func TuiRegion(tui tcell.Screen, x, y, w, h int) Region {
	return Region{
		W: w, H: h,
		SetCell: func(dx, dy int, style tcell.Style, ch rune) {
			if dx >= 0 && dx < w && dy >= 0 && dy < h {
				tui.SetCell(x+dx, y+dy, style, ch)
			}
		},
	}
}

var (
	whiteOnBlue = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
)

func drawText(region Region, style tcell.Style, text string) {
	for x, ch := range text {
		region.SetCell(x, 0, style, ch)
	}
}
