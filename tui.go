// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !noTUI

package main

import (
	"bytes"
	"code.google.com/p/goncurses"
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

type tuiT struct {
	G       *GoConfig
	Name    string
	rows    int
	cols    int
	row     int
	msg     string
	pkg     string
	scr     *goncurses.Window
	command map[goncurses.Key]func(*tuiT, int)
	help    *template.Template
	show    func(string, int, goncurses.Char)
}

const (
	NormalAttr = goncurses.A_NORMAL
	EntryAttr  = goncurses.A_STANDOUT
	StatusAttr = goncurses.A_BOLD
)

const tuiEntryHelpSrc = `
goconfig keys:
    EOF		Exit goconfig.
    ENTER	Set '{{.Name}}' with the prompted text.
		If this text is 'true', 'false' or 'nil', then '{{.Name}}'
		is set to the respective value.  You may quote such text to
		force string values; for example: "true", "false", "nil".
		In addition, you may set empty strings with paired quotes
		(i.e. "").
    SPACE	Toggle boolean build tags.
    [N]DOWN	Advance N (1) entries.
    [N]UP	Go back N (1) entries.
    [0]^R	Reinitialize '{{.Name}}' or all entries with 0 prefix.
    >		Save to {{.G.GoConfiguration}}
    !		Run the prompted command, If the command is "go", the first
		period ('.') argument is replaced by the package name and is
		prepended by appropriate goconfigured build flags.

{{.Marshal}}
`
const tuiPkgHelpSrc = `
goconfig keys:
    EOF		Exit goconfig.
    ENTER	goconfig '{{.Name}}'
    [N]DOWN	Advance N (1) entries.
    [N]UP	Go back N (1) entries.
`

var tuiEntryHelp, tuiPkgHelp *template.Template
var tuiEntryCommands = map[goncurses.Key]func(*tuiT, int){
	'?':                  tuiHelp,
	goncurses.KEY_UP:     tuiBackward,
	ctrlP:                tuiBackward,
	shiftTab:             tuiBackward,
	'k':                  tuiBackward,
	'-':                  tuiBackward,
	goncurses.KEY_DOWN:   tuiForward,
	ctrlN:                tuiForward,
	'\t':                 tuiForward,
	'j':                  tuiForward,
	'+':                  tuiForward,
	goncurses.KEY_PAGEUP: tuiPageUp,
	ctrlB:                tuiPageUp,
	metaV:                tuiPageUp,
	goncurses.KEY_PAGEDOWN: tuiPageDown,
	ctrlF:              tuiPageDown,
	ctrlV:              tuiPageDown,
	goncurses.KEY_HOME: tuiHome,
	metaLT:             tuiHome,
	'H':                tuiHome,
	goncurses.KEY_END:  tuiEnd,
	metaGT:             tuiEnd,
	ctrlL:              tuiRefresh,

	goncurses.KEY_RETURN: tuiSet,
	goncurses.KEY_ENTER:  tuiSet,
	ctrlR:                tuiReinit,
	' ':                  tuiToggle,
	'>':                  tuiStore,
	'!':                  tuiExec,
}
var tuiPkgCommands = map[goncurses.Key]func(*tuiT, int){
	'?':                  tuiHelp,
	goncurses.KEY_UP:     tuiBackward,
	ctrlP:                tuiBackward,
	shiftTab:             tuiBackward,
	'k':                  tuiBackward,
	'-':                  tuiBackward,
	goncurses.KEY_DOWN:   tuiForward,
	ctrlN:                tuiForward,
	'\t':                 tuiForward,
	'j':                  tuiForward,
	'+':                  tuiForward,
	goncurses.KEY_PAGEUP: tuiPageUp,
	ctrlB:                tuiPageUp,
	metaV:                tuiPageUp,
	goncurses.KEY_PAGEDOWN: tuiPageDown,
	ctrlF:              tuiPageDown,
	ctrlV:              tuiPageDown,
	goncurses.KEY_HOME: tuiHome,
	metaLT:             tuiHome,
	'H':                tuiHome,
	goncurses.KEY_END:  tuiEnd,
	metaGT:             tuiEnd,
	ctrlL:              tuiRefresh,

	goncurses.KEY_RETURN: tuiGoConfig,
	goncurses.KEY_ENTER:  tuiGoConfig,
}

func init() { AddMenu("tui", __tui__) }

func __tui__(v interface{}) (err error) {
	tuiEntryHelp = template.Must(template.New("tuiEntryHelp").Parse(
		tuiEntryHelpSrc[1:]))
	tuiPkgHelp = template.Must(template.New("tuiPkgHelp").Parse(
		tuiPkgHelpSrc[1:]))
	tui := new(tuiT)
	tui.G = v.(*GoConfig)
	if err = tui.init(); err != nil {
		return err
	}
outerLoop:
	for n := -1; true; n = -1 {
		goncurses.Cursor(0)
		goncurses.Echo(false)
		tui.show(tui.Name, tui.row, EntryAttr)
		tui.status(tui.msg)
	getkeyLoop:
		key := tui.scr.GetChar()
		if unicode.IsDigit(rune(key)) {
			if n < 0 {
				n = 0
			}
			n *= 10
			n += int(key) - '0'
			goto getkeyLoop
		}
		if n < 0 {
			n = 1
		}
		tui.show(tui.Name, tui.row, NormalAttr)
		if key == ctrlD || key == 'q' {
			break outerLoop
		} else if f, ok := tui.command[key]; ok {
			f(tui, n)
		} else {
			tui.msg = "Ignored unassigned key."

		}
	}
	goncurses.End()
	return
}

func tuiBackward(tui *tuiT, n int) {
	for i := 0; i < n; i++ {
		e, ok := tui.G.Entry[tui.Name]
		if !ok {
			break
		}
		s := e.prev
		if s == "" {
			break
		}
		tui.Name = s
		if tui.row -= 1; tui.row == -1 {
			tui.row = 0
			tui.scr.ScrollOk(true)
			tui.scr.Scroll(-1)
			tui.scr.ScrollOk(false)
			tui.show(s, tui.row, NormalAttr)
		}
	}
}

func tuiEnd(tui *tuiT, _ int) {
	for tui.Name != tui.G.End {
		tuiForward(tui, 1)
	}
}

func tuiExec(tui *tuiT, _ int) {
	if b, _ := tui.G.Exec(tui.prompt("! ")); len(b) != 0 {
		s := string(b)
		if strings.Index(s, "\n") >= 0 {
			tui.popup(func(args ...interface{}) {
				fmt.Fprintln(tui, args...)
			}, s)
		} else {
			tui.msg = s
		}
	}
}

func tuiForward(tui *tuiT, n int) {
	for i := 0; i < n; i++ {
		e, ok := tui.G.Entry[tui.Name]
		if !ok {
			break
		}
		s := e.next
		if s == "" {
			break
		}
		tui.Name = s
		if tui.row += 1; tui.row == tui.rows-1 {
			tui.scr.Move(tui.row, 0)
			tui.scr.ClearToEOL()
			tui.show(s, tui.row, NormalAttr)
			tui.scr.ScrollOk(true)
			tui.scr.Scroll(1)
			tui.scr.ScrollOk(false)
			tui.row -= 1
		}
	}
}

func tuiGoConfig(tui *tuiT, _ int) {
	if g, err := NewGoConfig(tui.Name); err == nil {
		if err = g.Load(nil); err != nil {
			tui.Error(err)
		}
		tui.G = g
		tui.Name = g.Begin
		tui.command = tuiEntryCommands
		tui.help = tuiEntryHelp
		tui.show = tui.showEntry
		tui.row = 0
		tui.refresh()
	} else {
		tui.Error(err)
	}
}

func tuiHelp(tui *tuiT, _ int) {
	tui.popup(func(_ ...interface{}) {
		if err := tui.help.Execute(tui, tui); err != nil {
			fmt.Fprintln(tui, err)
		}
	})
}

func tuiHome(tui *tuiT, _ int) {
	tui.Name = tui.G.Begin
	tui.row = 0
	tui.refresh()
}

func tuiRefresh(tui *tuiT, _ int) {
	tui.refresh()
}

func tuiPageUp(tui *tuiT, _ int) {
	tuiBackward(tui, tui.row)
	tui.row = tui.rows - 2
	for i, name := tui.row, tui.Name; i > 0; i -= 1 {
		e, ok := tui.G.Entry[name]
		if !ok {
			break
		}
		if name = e.prev; name == "" {
			tui.Name = tui.G.Begin
			tui.row = 0
			break
		}
	}
	tui.refresh()
}

func tuiPageDown(tui *tuiT, _ int) {
	tuiForward(tui, tui.rows-tui.row-3)
	tui.row = 0
	tui.refresh()
}

func tuiReinit(tui *tuiT, n int) {
	if n == 0 {
		tui.G.Reinit()
		tui.refresh()
	} else if e, ok := tui.G.Entry[tui.Name]; ok {
		e.Reinit()
		tui.show(tui.Name, tui.row, NormalAttr)
		tuiForward(tui, 1)
	}
}

func tuiSet(tui *tuiT, _ int) {
	var postXset bool
	name := tui.Name
	tui.show(name, tui.row, EntryAttr)
	if s := tui.prompt(": "); len(s) > 0 {
		postXset = tui.G.Set(name, s)
	}
	tui.show(tui.Name, tui.row, NormalAttr)
	tuiForward(tui, 1)
	if postXset {
		tui.refresh()
	}
}

func tuiStore(tui *tuiT, _ int) {
	if err := tui.G.Store(); err != nil {
		tui.Error(err)
	} else {
		tui.msg = "Wrote " + tui.G.GoConfiguration
	}
}

func tuiToggle(tui *tuiT, _ int) {
	if e, ok := tui.G.Entry[tui.Name]; ok && e != nil {
		if v := e.Value; v != nil {
			var refresh bool
			if v.IsTrue() {
				refresh = tui.G.Set(tui.Name, "false")
			} else if v.IsFalse() {
				refresh = tui.G.Set(tui.Name, "true")
			}
			tui.show(tui.Name, tui.row, EntryAttr)
			if refresh {
				tui.refresh()
			}
		}
	}
}

func (tui *tuiT) anyKey() {
	tui.status("Press any key to continue.")
	_ = tui.scr.GetChar()
	tui.scr.Move(tui.rows-1, 0)
	tui.scr.ClearToEOL()
	tui.scr.NoutRefresh()
	goncurses.Update()
}

func (tui *tuiT) Error(err error) {
	tui.msg = "error: " + err.Error()
}

func (tui *tuiT) init() (err error) {
	if tui.scr, err = goncurses.Init(); err != nil {
		return
	}
	tui.rows, tui.cols = tui.scr.MaxYX()
	tui.scr.Keypad(true)
	if tui.G.IsList() {
		tui.command = tuiPkgCommands
		tui.help = tuiPkgHelp
		tui.show = tui.showPkg
	} else {
		tui.command = tuiEntryCommands
		tui.help = tuiEntryHelp
		tui.show = tui.showEntry
	}
	tui.Name = tui.G.Begin
	tui.refresh()
	return
}

func (tui *tuiT) Marshal() string {
	return tui.G.Marshal(tui.Name)
}

func (tui *tuiT) popup(f func(...interface{}), args ...interface{}) {
	row := tui.row
	tui.row = 0
	tui.scr.Move(tui.row, 0)
	tui.scr.Clear()
	f(args...)
	tui.anyKey()
	tui.row = row
	tui.refresh()
}

func (tui *tuiT) prompt(a ...interface{}) string {
	var err error
	s := fmt.Sprint(a...)
	n := tui.cols - len(s)
	tui.status(s)
	goncurses.Cursor(1)
	goncurses.Echo(true)
	s, err = tui.scr.GetString(n)
	goncurses.Cursor(0)
	goncurses.Echo(false)
	if err == nil {
		return s
	}
	return ""
}

func (tui *tuiT) refresh() {
	tui.scr.Move(0, 0)
	tui.scr.Clear()
	for i, name := tui.row, tui.Name; name != "" && i >= 0; i -= 1 {
		tui.show(name, i, NormalAttr)
		name = tui.G.Entry[name].prev
	}
	for i, name := tui.row, tui.Name; name != "" && i < tui.rows-1; i += 1 {
		tui.show(name, i, NormalAttr)
		if name = tui.G.Entry[name].next; name == "" {
			break
		}
	}
	tui.msg = "goconfig " + tui.G.Package + "; press ? for help."
}

func (tui *tuiT) showEntry(name string, row int, attr goncurses.Char) {
	const sep = ": "
	const ellipsis = "..."
	e, ok := tui.G.Entry[name]
	if !ok {
		return
	}
	val := e.Value.YAML()
	if nl := strings.IndexAny(val, "\n\r"); nl >= 0 {
		val = val[:nl] + "..."
	}
	max := tui.cols - len(name) - len(sep)
	if len(val) > max {
		val = val[:max-len(ellipsis)] + ellipsis
	}
	tui.scr.Move(row, 0)
	tui.scr.ClearToEOL()
	if attr != NormalAttr {
		tui.scr.Print(name, sep)
		tui.scr.AttrOn(attr)
		tui.scr.Print(val)
		tui.scr.AttrOff(attr)
	} else {
		tui.scr.Print(name, sep, val)
	}
}

func (tui *tuiT) showPkg(name string, row int, attr goncurses.Char) {
	tui.scr.Move(row, 0)
	tui.scr.ClearToEOL()
	if attr != NormalAttr {
		tui.scr.AttrOn(attr)
		tui.scr.Print(name)
		tui.scr.AttrOff(attr)
	} else {
		tui.scr.Print(name)
	}
}

func (tui *tuiT) status(a ...interface{}) {
	tui.scr.Move(tui.rows-1, 0)
	tui.scr.ClearToEOL()
	tui.scr.AttrOn(StatusAttr)
	tui.scr.Print(a...)
	tui.scr.AttrOff(StatusAttr)
	tui.scr.NoutRefresh()
	goncurses.Update()
}

// This is a primitive pager for the Help and Exec output.
func (tui *tuiT) Write(b []byte) (n int, err error) {
	for len(b) > 0 && err == nil {
		if tui.row == tui.rows-1 {
			tui.anyKey()
			tui.row = 0
			tui.scr.Move(tui.row, 0)
			tui.scr.Clear()
		}
		if nl := bytes.IndexByte(b, '\n'); nl >= 0 {
			next := nl + 1
			tui.scr.Print(string(b[:next]))
			n += len(b[:next])
			b = b[next:]
			tui.row += 1
		} else {
			tui.scr.Print(string(b))
			n += len(b)
			break
		}
	}
	return
}
