// up is the ultimate pipe composer/editor. It helps building Linux pipelines
// in a terminal-based UI interactively, with live preview of command results.
package main

import (
	"io"
	"os"

	"github.com/jroimartin/gocui"
)

const (
	queryTag  = "query"
	outputTag = "output"
)

func main() {
	// In background, start collecting input from stdin to internal buffer of size 40 MB, then pause it
	go collect()

	// Init TUI code
	tui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer tui.Close()

	// Prepare TUI layout etc.
	tui.SetManagerFunc(layout)
	w, h := tui.Size()
	query, err := tui.SetView(queryTag, 0, 0, w-1, 3)
	if err != nil && err != gocui.ErrUnknownView {
		panic(err)
	}
	query.Title = "Command"
	query.BgColor = gocui.ColorCyan
	query.Editable = true
	// query.Editor = gocui.DefaultEditor
	output, err := tui.SetView(outputTag, 0, 3, w-1, h-1)
	if err != nil && err != gocui.ErrUnknownView {
		panic(err)
	}
	output.Title = "Output"
	output.Autoscroll = true
	err = tui.SetKeybinding(queryTag, gocui.KeyCtrlC, gocui.ModNone, func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	})
	if err != nil {
		panic(err)
	}
	tui.SetCurrentView(queryTag)
	err = tui.MainLoop()
	if err != nil && err != gocui.ErrQuit {
		panic(err)
	}

	// TODO: using tcell, edit a command in bash format in multiline input box (or jroimartin/gocui?)
	//       NOTE: gocui has trouble if we capture stdin. Try butchering ("total modding") peco/peco instead.
	// TODO: run it automatically in bg after first " " (or ^Enter), via `bash -c`
	// TODO: auto-kill the child process on any edit
	// TODO: [LATER] Ctrl-O shows input via `less` or $PAGER
	// TODO: ^X - save into executable file upN.sh (with #!/bin/bash) and quit
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
}

func collect() {
	const bufsize = 40 * 1024 * 1024 // 40 MB
	buf := make([]byte, bufsize)
	// TODO: read gradually what is available and show progress
	n, err := io.ReadFull(os.Stdin, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		panic(err)
	}
	buf = buf[:n]
	// TODO: use buf somewhere
}

func layout(tui *gocui.Gui) error {
	// w, h := tui.Size()
	return nil
}
