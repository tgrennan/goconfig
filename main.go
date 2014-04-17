// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"gopkg.in/tgrennan/fixme.v0"
	"gopkg.in/tgrennan/sos.v0"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type mainT struct {
	a sos.SoS
	f *os.File
	b *bytes.Buffer
	g *GoConfig
}

const usageSrc = `
Usage:
	goconfig [flags] [-cli] [package]{{if .WebServer}}
	goconfig [flags] -http=<server:port>{{end}}
	goconfig [flags] -show [-all] [package]
	goconfig [flags] <[go] command> [go flags] [package [args]]

Flags:
	-fixme[=<file>]
		Print debugging messages on stderr or in the given file.

	-config[=<file>]
		Load configuration from stdin or the given file instead of the
		default, goconfiguration_GOOS_GOARCH.yaml

Options:{{.TUI}}{{.WebServer}}
	-show [-all]
		Instead of a menu, print the configured [or all] entries.

	go <command> [build and test flags] [package [args]]
		Run the given command with the configured constraints and
		strings.

Goconfig operates on one package per execution unless given ` +
	"`all`" + `
where it makes a menu of packages within GOPATH containing:
` + "`goconfig[_GOOS][_GOARCH].yaml`" + `
`

var (
	egress  = errors.New("egress")
	exit    = os.Exit
	usage   *template.Template
	Version string
)

func init() {
	usage = template.Must(template.New("Usage").Parse(usageSrc[1:]))
}

func main() {
	var err error
	m := new(mainT)
	log.SetFlags(0)
	log.SetPrefix("goconfig: ")
	log.SetOutput(os.Stderr)
	defer func() {
		if m.f != nil {
			m.f.Close()
		}
		if err != nil && err != egress {
			log.Print(err)
			exit(1)
		}
	}()
	m.a = sos.New(os.Args[1:]...)
	for _, f := range []func() error{
		m.help,
		m.version,
		m.fixme,
		m.config,
		m.gotool,
		m.show,
		m.webserver,
		m.tui,
		m.cli,
	} {
		if err = f(); err != nil {
			return
		}
	}
}

func (m *mainT) arg(name string) string {
	var s string
	m.a, s = m.a.Arg(name)
	return s
}

func (m *mainT) cli() (err error) {
	if cli, ok := Menu["cli"]; ok {
		if err = m.goconfig(); err == nil {
			if err = cli(m.g); err == nil {
				err = egress
			}
		}
	}
	return
}

func (m *mainT) config() (err error) {
	var name string
	var file *os.File
	if m.flag("config") {
		file = os.Stdin
		name = "Stdin"
	} else if name = m.arg("config"); name != "" {
		file, err = os.Open(name)
		if err != nil {
			return
		}
		defer file.Close()
	}
	if file != nil {
		m.b = new(bytes.Buffer)
		_, err = m.b.ReadFrom(file)
	}
	return
}

func (m *mainT) flag(name string) bool {
	var flag bool
	m.a, flag = m.a.Flag(name)
	return flag
}

func (m *mainT) fixme() (err error) {
	if m.flag("fixme") {
		fixme.Enable()
	} else if name := m.arg("fixme"); name != "" {
		if m.f, err = os.Create(name); err == nil {
			fixme.SetWriter(m.f)
			fixme.Enable()
		}
	}
	return
}

func (m *mainT) goconfig() (err error) {
	var pkg string
	m.a, pkg = m.a.Pop()
	if strings.HasPrefix(pkg, "-") {
		return fmt.Errorf("invalid flag: %s", pkg)
	}
	if m.g, err = NewGoConfig(pkg); err != nil {
		return
	}
	if pkg == ALL {
		fixme.Println(err)
		return
	}
	err = m.g.Load(m.b)
	return
}

func (m *mainT) gotool() (err error) {
	for _, s := range []string{"go", "build", "run", "test", "install"} {
		if m.a.String(0) == s {
			var c *GoCommand
			c, m.a = NewGoCommand(m.a)
			if err = m.goconfig(); err == nil {
				var b []byte
				if b, err = m.g.GoTool(c, m.a); err == nil {
					os.Stdout.Write(b)
					err = egress
				} else {
					os.Stderr.Write(b)
				}
			}
			break
		}
	}
	return
}

func (m *mainT) help() (err error) {
	var opt struct{ TUI, WebServer string }
	if _, ok := Menu["tui"]; ok {
		opt.TUI = `
	-cli
		The default action is a terminal user interface unless running
		on a DUMB terminal or this flag is given to revert to a command
		line interface.
`

	}
	if _, ok := Menu["webserver"]; ok {
		opt.WebServer = `
	-http=<server:port>
		Runs a web server at the given address instead of a TUI or CLI.
`
	}
	if m.a.String(0) == "help" || m.flag("help") {
		if err = usage.Execute(os.Stdout, opt); err == nil {
			err = egress
		}
	}
	return
}

func (m *mainT) show() (err error) {
	var all bool
	if m.a.String(0) == "show" {
		m.a, _ = m.a.Pop()
	} else if !m.flag("show") {
		return
	}
	m.a, all = m.a.Flag("all")
	if err = m.goconfig(); err != nil {
		return
	}
	err = egress
	for _, e := range m.g.Entries {
		if all || !e.Value.Equal(e.Init) {
			yaml := e.Value.YAML()
			fmt.Print(e.Name, ": ", yaml, "\n")
		}
	}
	return
}

func (m *mainT) tui() (err error) {
	if !m.flag("cli") {
		if term := os.Getenv("TERM"); term != "" && term != "DUMB" {
			if tui, ok := Menu["tui"]; ok {
				if err = m.goconfig(); err == nil {
					if err = tui(m.g); err == nil {
						err = egress
					}
				}
			}
		}
	}
	return
}

func (m *mainT) version() (err error) {
	if m.a.String(0) == "version" || m.flag("version") {
		if Version == "" {
			Version = "unknown"
		}
		os.Stdout.Write([]byte("goconfig version " + Version + "\n"))
		err = egress
	}
	return
}

func (m *mainT) webserver() (err error) {
	var address string
	if m.a, address = m.a.Arg("http"); address == "" {
		return
	}
	if colon := strings.Index(address, ":"); colon < 0 {
		err = fmt.Errorf("invalid service address: %s", address)
	} else if _, err = strconv.Atoi(address[colon+1:]); err == nil {
		if webserver, ok := Menu["webserver"]; ok {
			if err = webserver(address); err == nil {
				err = egress
			}
		} else {
			err = errors.New("built without webserver")
		}
	}
	return
}
