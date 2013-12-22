// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !noCLI

package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/tgrennan/fixme"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

type cliT struct {
	G       *GoConfig
	Name    string
	rows    int
	cols    int
	row     int
	scanner *bufio.Scanner
	command map[rune]func(*cliT, int, string)
	help    *template.Template
	prompt  func()
}

const cliEntryHelpSrc = `
goconfig commands:
    EOF		Exit goconfig.
    ENTER	Advance to next entry.
    [N]+	Advance N (1) entries.
    [N]-	Go back N (1) entries.
    #		Show the configuration.
    [0]^	Reinitialize '{{.Name}}' or all entries with 0 prefix.
    >		Save to {{.G.GoConfiguration}}
    !<command>
    !go <command> [go flags] . [target flags]
		Run the given command.  If the command is "go", the first
		period ('.') argument is replaced by the package name and is
		prepended by appropriate goconfigured build flags.

Anything else sets '{{.Name}}' to the given text.  If this text is 'true',
'false' or 'nil', then '{{.Name}}' is set to the respective value.  You may
quote such text to force string values; for example: "true", "false", "nil".
In addition, you may set empty strings with paired quotes (i.e. "").

{{.Marshal}}
`
const cliPkgHelpSrc = `
goconfig commands:
    EOF		Exit goconfig.
    ENTER	goconfig {{.Name}}
    [N]+	Advance N (1) entries.
    [N]-	Go back N (1) entries.
`

var cliEntryHelp, cliPkgHelp *template.Template
var cliErrorCommand = errors.New("unknown command")
var cliEntryCommands = map[rune]func(*cliT, int, string){
	0:   cliForward1,
	'?': cliShowHelp,
	'!': cliExec,
	'+': cliForward,
	'-': cliBackward,
	'>': cliStore,
	'#': cliShow,
	'^': cliReinit,
}
var cliPkgCommands = map[rune]func(*cliT, int, string){
	0:   cliGoConfig,
	'?': cliShowHelp,
	'!': cliExec,
	'+': cliForward,
	'-': cliBackward,
}

func init() { AddMenu("cli", __cli__) }

func __cli__(v interface{}) error {
	cliEntryHelp = template.Must(template.New("cliEntryHelp").Parse(
		cliEntryHelpSrc[1:]))
	cliPkgHelp = template.Must(template.New("cliPkgHelp").Parse(
		cliPkgHelpSrc[1:]))
	cli := new(cliT)
	cli.G = v.(*GoConfig)
	cli.init()
	cli.title()
	for {
		cli.row = 0
		cli.prompt()
		if !cli.scanner.Scan() {
			println()
			return cli.scanner.Err()
		}
		n := 1
		t := cli.scanner.Text()
		if len(t) > 0 {
			r := rune(t[0])
			if unicode.IsDigit(r) {
			digitLoop:
				for i, c := range t {
					if r = rune(c); !unicode.IsDigit(r) {
						x, err := strconv.Atoi(t[:i])
						if err == nil && x >= 0 {
							n = x
						}
						t = t[i:]
						break digitLoop
					}
				}
			}
		}
		r := rune(0)
		args := ""
		if len(t) > 0 {
			r = rune(t[0])
			args = strings.TrimSpace(t[1:])
		}
		if f, ok := cli.command[r]; ok {
			f(cli, n, args)
		} else if cli.G.IsList() {
			cli.Error(cliErrorCommand)
		} else {
			cli.G.Set(cli.Name, strings.TrimSpace(t))
			if s := cli.G.Entry[cli.Name].next; s != "" {
				cli.Name = s
			}
		}
	}
}

func cliBackward(cli *cliT, n int, _ string) {
	for i := 0; i < n; i++ {
		e, ok := cli.G.Entry[cli.Name]
		if !ok {
			break
		}
		if s := e.prev; s != "" {
			cli.Name = s
		} else {
			break
		}
	}
}

func cliExec(cli *cliT, _ int, s string) {
	b, _ := cli.G.Exec(s)
	cli.row = 0
	cli.Write(b)
}

func cliForward(cli *cliT, n int, _ string) {
	for i := 0; i < n; i++ {
		e, ok := cli.G.Entry[cli.Name]
		if !ok {
			break
		}
		if s := e.next; s != "" {
			cli.Name = s
		} else {
			break
		}
	}
}

func cliForward1(cli *cliT, _ int, _ string) {
	if e, ok := cli.G.Entry[cli.Name]; ok {
		if s := e.next; s != "" {
			cli.Name = s
		}
	}
}

func cliGoConfig(cli *cliT, _ int, _ string) {
	if g, err := NewGoConfig(cli.Name); err == nil {
		if err = g.Load(nil); err != nil {
			cli.Error(err)
		}
		cli.G = g
		cli.Name = g.Begin
		cli.command = cliEntryCommands
		cli.help = cliEntryHelp
		cli.prompt = cli.entry
		cli.title()
	} else {
		cli.Error(err)
	}
}

func cliReinit(cli *cliT, n int, _ string) {
	if n == 0 {
		cli.G.Reinit()
	} else {
		cli.G.Entry[cli.Name].Reinit()
		if s := cli.G.Entry[cli.Name].next; s != "" {
			cli.Name = s
		}
	}
}

func cliShow(cli *cliT, _ int, _ string) {
	for _, e := range cli.G.Entries {
		print(e.Name, ": ", e.Value.YAML(), "\n")
	}
}

func cliShowHelp(cli *cliT, _ int, _ string) {
	if err := cli.help.Execute(cli, cli); err != nil {
		fixme.Println(err)
	}
}

func cliStore(cli *cliT, _ int, _ string) {
	if err := cli.G.Store(); err != nil {
		cli.Error(err)
	} else {
		println("Wrote:", cli.G.GoConfiguration)
	}
}

func (cli *cliT) entry() {
	if e, ok := cli.G.Entry[cli.Name]; ok {
		print(cli.Name, ": ", e.Value.YAML(), "$ ")
	} else {
		print("$ ")
	}
}

func (cli *cliT) Error(err error) {
	println("error:", err.Error())
}

func (cli *cliT) init() {
	cli.Name = cli.G.Begin
	cli.resize()
	cli.scanner = bufio.NewScanner(os.Stdin)
	if cli.G.IsList() {
		cli.command = cliPkgCommands
		cli.help = cliPkgHelp
		cli.prompt = func() {
			print(cli.Name, "$ ")
		}
	} else {
		cli.command = cliEntryCommands
		cli.help = cliEntryHelp
		cli.prompt = cli.entry
	}
}

func (cli *cliT) Marshal() string {
	return cli.G.Marshal(cli.Name)
}

func (cli *cliT) resize() {
	cli.rows, cli.cols = 24, 80
	i, err := strconv.Atoi(os.Getenv("LINES"))
	if err == nil && i > 0 {
		cli.rows = i
		i, err = strconv.Atoi(os.Getenv("COLUMNS"))
		if err == nil && i > 0 {
			cli.cols = i
		}
	} else {
		cmd := exec.Command("stty", "size")
		cmd.Stdin = os.Stdin
		buf, _ := cmd.CombinedOutput()
		sbuf := strings.TrimSpace(string(buf))
		if space := strings.Index(sbuf, " "); space > 0 {
			i, err = strconv.Atoi(sbuf[:space])
			if err == nil && i > 0 {
				cli.rows = i
				i, err = strconv.Atoi(sbuf[space+1:])
				if err == nil && i > 0 {
					cli.cols = i
				}
			}
		}
	}
}

func (cli *cliT) title() {
	println("goconfig", cli.G.Package, "- enter ? for help.")
}

// This is a primitive pager for the Help and Exec output.
func (cli *cliT) Write(b []byte) (n int, err error) {
	for len(b) > 0 && err == nil {
		if cli.row == cli.rows-1 {
			os.Stdout.Write([]byte("Press Enter to continue."))
			cli.row = 0
			if !cli.scanner.Scan() {
				err = cli.scanner.Err()
				break
			}
			cli.scanner.Text()
		}
		x := 0
		if nl := bytes.IndexByte(b, '\n'); nl >= 0 {
			next := nl + 1
			x, err = os.Stdout.Write(b[:next])
			b = b[next:]
			n += x
			cli.row += 1
		} else {
			x, err = os.Stdout.Write(b)
			n += x
			break
		}
	}
	return
}
