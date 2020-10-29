<img align="left" width="183" height="154" src="logo.svg" />

# up - the Ultimate Plumber

**up** is the **Ultimate Plumber**, a tool for writing Linux pipes in a
terminal-based UI interactively, with instant live preview of command results.

The main **goal** of the Ultimate Plumber is to help **interactively and
incrementally explore textual data** in Linux, by making it easier to quickly
build complex pipelines, thanks to a **fast feedback loop**. This is achieved
by boosting any typical **Linux text-processing utils** such as `grep`, `sort`,
`cut`, `paste`, `awk`, `wc`, `perl`, etc., etc., by providing a quick,
**interactive, scrollable preview** of their results.

[![](up.gif)](https://asciinema.org/a/208538)

## Usage

**[Download *up* for Linux](https://github.com/akavel/up/releases/latest/download/up)**
&nbsp; | &nbsp; [ArchLinux](https://wiki.archlinux.org/index.php/Arch_User_Repository): [`aur/up`](https://aur.archlinux.org/packages/up/)
&nbsp; | &nbsp; FreeBSD: [`pkg install up`](https://www.freshports.org/textproc/up)
&nbsp; | &nbsp; macOS: [`brew install up`](https://formulae.brew.sh/formula/up)
&nbsp; | &nbsp; [Other OSes](https://github.com/akavel/up/releases)

To start using **up**, redirect any text-emitting command (or pipeline) into it
— for example:

    $ lshw |& ./up

then:

- use ***PgUp/PgDn*** and ***Ctrl-[←]/Ctrl-[→]*** for basic browsing through
  the command output;
- in the input box at the top of the screen, start **writing any bash
  pipeline**; then **press Enter to execute the command you typed**,
  and the Ultimate Plumber will immediately show you the output of
  the pipeline in the **scrollable window** below (replacing any
  earlier contents)
    - For example, you can try writing:
      `grep network -A2 | grep : | cut -d: -f2- | paste - -`
      — on my computer, after pressing *Enter*, the screen then shows
      the pipeline and a scrollable preview of its output like below:

             | grep network -A2 | grep : | cut -d: -f2- | paste - -
             Wireless interface      Centrino Advanced-N 6235
             Ethernet interface      RTL8111/8168/8411 PCI Express Gigabit Ethernet Controller

    - **WARNING: Please be careful when using it! It could be dangerous.**
      In particular, writing "rm" or "dd" into it could be like running around
      with a chainsaw. But you'd be careful writing "rm" anywhere in Linux
      anyway, no?
- when you are satisfied with the result, you can **press *Ctrl-X* to exit**
  the Ultimate Plumber, and the command you built will be **written into
  `up1.sh` file** in the current working directory (or, if it already existed,
  `up2.sh`, etc., until 1000, based on [Shlemiel the Painter's
  algorithm](https://www.joelonsoftware.com/2001/12/11/back-to-basics/)).
  Alternatively, you can press ***Ctrl-C*** to quit without saving.
- If the command you piped into *up* is long-running (in such case you will see
  a tilde `~` indicator character in the top-left corner of the screen, meaning
  that *up* is still waiting for more input), you may need to press
  ***Ctrl-S*** to temporarily freeze *up*'s input buffer (a freeze will be
  indicated by a `#` character in top-left corner), which will inject a fake
  EOF into the pipeline; otherwise, some commands in the pipeline may not print
  anything, waiting for full input (especially commands like `wc` or `sort`,
  but `grep`, `perl`, etc. may also show incomplete results). To unfreeze back,
  press ***Ctrl-Q***.

## Additional Notes

- The pipeline is passed verbatim to a `bash -c` command, so any bash-isms should work.
- The input buffer of the Ultimate Plumber is currently fixed at **40 MB**. If
  you reach this limit, a `+` character should get displayed in the top-left
  corner of the screen. (This is intended to be changed to a
  dynamically/manually growable buffer in a future version of *up*.)
- **MacOSX support:** I don't have a Mac, thus I have no idea if it works on
  one. You are welcome to try, and also to send PRs. If you're interested in
  me providing some kind of official-like support for MacOSX, please consider
  trying to find a way to send me some usable-enough Mac computer. Please note
  I'm not trying to "take advantage" of you by this, as I'm actually not at all
  interested in achieving a Mac otherwise. (Also, trying to commit to this kind
  of support will be an extra burden and obligation on me. Knowing someone out
  there cares enough to do a fancy physical gesture would really help alleviate
  this.) If you're serious enough to consider this option, please contact me by
  email (mailto:czapkofan@gmail.com) or keybase (https://keybase.io/akavel), so
  that we could try to research possible ways to achieve this.
  Thanks for understanding!
- **Prior art:** I was surprised no one seemed to write a similar tool before,
  that I could find. It should have been possible to write this since the dawn
  of Unix already, or earlier! And indeed, after I announced *up*, I got enough
  publicity that my attention was directed to one such earlier project already:
  **[Pipecut](http://pipecut.org/index.html)**. Looks interesting! You may like
  to check it too! (Thanks [@TronDD](https://lobste.rs/s/acpz00/up_tool_for_writing_linux_pipes_with#c_qxrgoa).)
- **Other influences:** I don't remember the fact too well already, but I'm
  rather sure that this must have been inspired in big part by The Bret Victor's Talk(s).

## Future Ideas

- I have quite a lot of ideas for further experimentation of development of
  *up*, including but not limited to:
    - [RIIR](https://rust-lang.org) (once I learn enough of Rust... at some
      point in future... maybe...) — esp. to hopefully make *up* be a smaller
      binary (and also to maybe finally learn some Rust); though I'm somewhat
      afraid if it might ossify the codebase and make harder to develop
      further..? ...but maybe actually converse?...
    - Maybe it could be made into an UI-less, RPC/REST/socket/text-driven
      service, like gocode or [Language Servers](https://langserver.org/), for
      integration with editors/IDEs (emacs? vim? VSCode?...) I'd be especially
      interested in eventually merging it into [Luna
      Studio](https://luna-lang.org/); RIIR may help in this. (Before this, as
      a simpler approach, multi-line editing may be needed, or at least
      left&right scrolling of the command editor input box. Also, some kind of
      jumping between words in the command line; readline's *Alt-b* & *Alt-f*?)
    - Make it possible to [capture output of already running
      processes](https://stackoverflow.com/a/19584979/98528)! (But maybe that
      could be better made as a separate, composable tool! In Rust?)
    - Adding tests... (ahem; see also
      [#1](https://github.com/akavel/up/issues/1)) ...also write `--help`...
    - Making it work on Windows,
      somehow[?](https://github.com/mattn/go-shellwords) Also, obviously, would
      be nice to have some CI infrastructure enabling porting it to MacOSX,
      BSDs, etc., etc...
    - Integration with [fzf](https://github.com/junegunn/fzf) and other TUI
      tools? I only have some vague thoughts and ideas about it as of now, not
      even sure how this could look like.
    - Adding more previews, for each `|` in the pipeline; also forking of
      pipelines, merging, feedback loops, and other mixing and matching (though
      I'd strongly prefer if [Luna](https://luna-lang.org) was to do it
      eventually).
- If you are interested in financing my R&D work, contact me by email at:
  czapkofan@gmail.com, or [on keybase.io as akavel](https://keybase.io/akavel).
  I suppose I will probably be developing the Ultimate Plumber further anyway,
  but at this time it's purely a hobby project, with all the fun and risks this
  entails.

— *Mateusz Czapliński*  
*October 2018*

*PS. The UP logo was conceived and generously sponsored
by [Thoai Nguyen](https://github.com/thoaionline)
and [GPU Exchange](https://gpu.exchange/),
with a helping hand from [Many Pixels](https://www.manypixels.co/).*
