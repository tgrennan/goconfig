// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"github.com/tgrennan/fixme"
	"github.com/tgrennan/quotation"
	"github.com/tgrennan/sos"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type GoConfig struct {
	GoConfiguration string

	Package string
	Dir     string
	Begin   string
	End     string
	Entry   map[string]*Entry
	Entries []*Entry
}

type GoCommand struct {
	Name        string
	Flags       map[string]bool
	StringFlags map[string]string
}

const ALL = "all"

var Menu map[string]func(interface{}) error
var mutex = &sync.Mutex{}
var veritas = true
var goos, goarch, goconfiguration string
var goconfig_goos_goarch, goconfig_goarch, goconfig_goos, goconfig string
var goconfigs []string
var GoBuildFlags = map[string]bool{
	"a":    true,
	"n":    true,
	"race": true,
	"v":    true,
	"x":    true,
}
var GoConfigurableBuildFlags = map[string]bool{
	"race": true,
}
var GoBuildStringFlags = map[string]bool{
	"o":             true,
	"p":             true,
	"ccflags":       true,
	"compiler":      true,
	"gccgoflags":    true,
	"gcflags":       true,
	"installsuffix": true,
	"ldflags":       true,
}
var GoConfigurableBuildStringFlags = map[string]bool{
	"ccflags":       true,
	"compiler":      true,
	"gccgoflags":    true,
	"gcflags":       true,
	"installsuffix": true,
	"ldflags":       true,
}
var GoTestFlags = map[string]bool{
	"c": true,
	"i": true,
}

func init() {
	goos = runtime.GOOS
	if s := os.Getenv("GOOS"); s != "" {
		goos = s
	}
	goarch = runtime.GOARCH
	if s := os.Getenv("GOARCH"); s != "" {
		goarch = s
	}
	goconfig_goos_goarch = "goconfig_" + goos + "_" + goarch + ".yaml"
	goconfig_goarch = "goconfig_" + goarch + ".yaml"
	goconfig_goos = "goconfig_" + goos + ".yaml"
	goconfig = "goconfig.yaml"
	goconfiguration = "goconfiguration_" + goos + "_" + goarch + ".yaml"
	goconfigs = []string{
		goconfig_goos_goarch,
		goconfig_goarch,
		goconfig_goos,
		goconfig,
	}
}

func IsGoFlag(name string) bool {
	if _, ok := GoBuildFlags[name]; ok {
		return true
	}
	if _, ok := GoBuildStringFlags[name]; ok {
		return true
	}
	if _, ok := GoTestFlags[name]; ok {
		return true
	}
	return false
}

func NewGoConfig(pkg string) (*GoConfig, error) {
	var err error
	g := new(GoConfig)
	g.Entry = make(map[string]*Entry)
	if pkg == "" || pkg == "." {
		b, err := exec.Command("go", "list").CombinedOutput()
		if err == nil {
			g.Package = strings.TrimSpace(string(b))
		} else {
			return nil, err
		}
	} else {
		g.Package = pkg
	}
	if g.Package == ALL {
		err = g.listALL()
	} else {
		err = g.unmarshal(g.Package)
	}
	if err != nil {
		g.Clean()
	} else {
		g.Entries = make([]*Entry, 0)
		for name := g.Begin; name != ""; name = g.Entry[name].next {
			e := g.Entry[name]
			e.Name = name
			g.Entries = append(g.Entries, e)
		}
	}
	return g, err
}

func NewGoCommand(args sos.SoS) (*GoCommand, sos.SoS) {
	c := new(GoCommand)
	c.Flags = make(map[string]bool)
	c.StringFlags = make(map[string]string)
	args, c.Name = args.Pop()
	if c.Name == "go" {
		args, c.Name = args.Pop()
	}
	for k := range GoBuildFlags {
		var t bool
		if args, t = args.Flag(k); t {
			c.Flags[k] = t
		}
	}
	if c.Name == "test" {
		for k := range GoTestFlags {
			var t bool
			if args, t = args.Flag(k); t {
				c.Flags[k] = t
			}
		}
	}
	for k := range GoBuildStringFlags {
		var s string
		if args, s = args.Arg(k); s != "" {
			c.StringFlags[k] = s
		}
	}
	return c, args
}

func (g *GoConfig) Clean() {
	for name, e := range g.Entry {
		if e.Init != nil {
			e.Init.SetFalse()
		}
		if e.Value != nil {
			e.Value.SetFalse()
		}
		e.Choices = nil
		for x, u := range e.Set {
			if u != nil {
				u.SetFalse()
			}
			delete(e.Set, x)
		}
		for x, u := range e.Reset {
			if u != nil {
				u.SetFalse()
			}
			delete(e.Reset, x)
		}
		delete(g.Entry, name)
	}
	g.Entries = g.Entries[:0]
}

func (g *GoConfig) Exec(s string) ([]byte, error) {
	a := quotation.Fields(s)
	if a[0] == "go" && len(a) > 1 {
		c, sos := NewGoCommand(sos.SoS(a))
		return g.GoTool(c, sos)
	}
	return exec.Command(a[0], a[1:]...).CombinedOutput()
}

func (g *GoConfig) GoTool(c *GoCommand, a sos.SoS) ([]byte, error) {
	for _, f := range []func(*GoCommand, sos.SoS) (sos.SoS, error){
		g.pushSubject,
		g.pushBuildStringFlags,
		g.pushBuildLDFlag,
		g.pushBuildTagsFlag,
		g.pushTestFlags,
		g.pushBuildFlags,
	} {
		if xa, err := f(c, a); err != nil {
			return nil, err
		} else {
			a = xa
		}
	}
	a = a.Push(c.Name)
	a = a.Push("go")
	buf, err := exec.Command(a[0], a[1:]...).CombinedOutput()
	if nt, nok := c.Flags["n"]; nok && nt {
		s := "#\n# "
		for _, as := range a {
			if strings.Index(as, " ") >= 0 {
				s += " '" + as + "'"
			} else {
				s += " " + as
			}
		}
		s += "\n#\n"
		return append([]byte(s), buf...), err
	}
	return buf, err
}

func (g *GoConfig) Has(name string) bool {
	_, t := g.Entry[name]
	return t
}

func (g *GoConfig) importer(pkg string) error {
	var prefix string
	if pkg[0] == '.' {
		if i := strings.LastIndex(pkg, "/"); i > 0 {
			prefix = pkg[:i+1]
		}
		pkg = filepath.Clean(filepath.Join(g.Dir, pkg))
		// prefix = "_" + g.Dir
	}
	gi, err := NewGoConfig(pkg)
	if err != nil {
		return err
	}
	for name, e := range gi.Entry {
		var iname string
		if e == nil {
			fixme.Println(name, "from", pkg, "has nil e")
		} else if e.Value == nil {
			fixme.Println(name, "from", pkg, "has nil value")
		} else if name != "import" {
			if !e.Value.IsTag() && prefix != "" {
				iname = prefix + name
			} else if dot := strings.Index(name, "."); dot > 0 {
				iname = pkg + name[dot:]
			} else {
				iname = name
			}
			if g.Has(iname) {
				fixme.Println(g.Package, "ignoring duplicate",
					iname, "from", pkg)
			} else {
				g.Entry[iname] = e
				g.insert(iname)
			}
		}
		delete(gi.Entry, name)
	}
	return nil
}

func (g *GoConfig) insert(subject string) {
	subjectEntry := g.Entry[subject]
	if g.Begin == "" {
		g.Begin = subject
		g.End = subject
		return
	}
	if subjectEntry == nil {
		fixme.Println(subject, "has nil entry")
		return
	}
	if subjectEntry.Value == nil {
		fixme.Println(subject, "has nil value")
		return
	}
	for target := g.Begin; target != ""; target = g.Entry[target].next {
		targetEntry := g.Entry[target]
		if targetEntry == nil {
			fixme.Println(target, "has nil entry")
		} else if targetEntry.Value == nil {
			fixme.Println(target, "has nil value")
		} else if (subjectEntry.Value.IsTag() && !targetEntry.Value.IsTag()) ||
			(subject < target &&
				(!targetEntry.Value.IsTag() ||
					subjectEntry.Value.IsTag())) {
			subjectEntry.next = target
			subjectEntry.prev = targetEntry.prev
			targetEntry.prev = subject
			if target == g.Begin {
				g.Begin = subject
			} else {
				g.Entry[subjectEntry.prev].next = subject
			}
			return
		} else if target == g.End {
			subjectEntry.next = ""
			subjectEntry.prev = g.End
			g.Entry[g.End].next = subject
			g.End = subject
			return
		}
	}
}

func (g *GoConfig) IsList() bool { return g.Package == ALL }

func (g *GoConfig) listALL() error {
	for _, path := range filepath.SplitList(os.Getenv("GOPATH")) {
		src := filepath.Join(path, "src")
		trim := src + string(os.PathSeparator)
		if fi, err := os.Stat(src); err != nil || !fi.IsDir() {
			continue
		}
		filepath.Walk(src,
			func(full string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				for _, name := range goconfigs {
					if filepath.Base(full) == name {
						dir := filepath.Dir(full)
						iPath := strings.TrimPrefix(dir,
							trim)
						if !g.Has(iPath) {
							e := new(Entry)
							e.Value = NewUnion(full)
							g.Entry[iPath] = e
							g.insert(iPath)
						}
					}
				}
				return nil
			})
	}
	return nil
}

func (g *GoConfig) Load(config *bytes.Buffer) (err error) {
	if config == nil {
		var file *os.File
		file, err = os.Open(g.GoConfiguration)
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return
		}
		defer file.Close()
		config = new(bytes.Buffer)
		_, err = config.ReadFrom(file)
		if err != nil {
			return
		}
	}
	m := make(map[string]interface{})
	if err = goyaml.Unmarshal(config.Bytes(), m); err != nil {
		return
	}
	for name, v := range m {
		if e, ok := g.Entry[name]; ok {
			e.Value.Set(v)
		} else {
			fixme.Println(name, "not found")
		}
		delete(m, name)
	}
	return
}

func (g *GoConfig) Marshal(name string) string {
	const nlFieldIndent = "\n    "
	const nlValIndent = "\n        "
	expandString := func(t string) (x string) {
		if strings.ContainsRune(t, '\n') {
			x += "!!str |"
			t = strings.Replace(t, "\n", nlValIndent, -1)
			n := len(t) - len(nlValIndent)
			if t[n:] == nlValIndent {
				t = t[:n]
			}
			x += nlValIndent + t
		} else if t == "" {
			x += `""`
		} else {
			x += t
		}
		return
	}
	expandMap := func(m map[string]*Union) (x string) {
		for k, u := range m {
			x += nlValIndent + k + ": " + expandString(u.String())
		}
		return
	}

	e, ok := g.Entry[name]
	if !ok {
		return ""
	}
	s := name + ":"
	if e.Help != "" {
		s += nlFieldIndent + "help: " + expandString(e.Help)
	}
	if e.Init != nil {
		s += nlFieldIndent + "init: " + expandString(e.Init.String())
	}
	if len(e.Choices) > 0 {
		s += nlFieldIndent + "choices: "
		for _, x := range e.Choices {
			s += nlValIndent + "- " + x
		}
	}
	if len(e.Set) > 0 {
		s += nlFieldIndent + "set:" + expandMap(e.Set)
	}
	if len(e.Reset) > 0 {
		s += nlFieldIndent + "reset:" + expandMap(e.Reset)
	}
	s += "\n"
	return s
}

func (g *GoConfig) pushBuildFlags(c *GoCommand, a sos.SoS) (sos.SoS, error) {
	for k := range GoBuildFlags {
		if t, ok := c.Flags[k]; ok && t {
			a = a.Push("-" + k)
		} else if _, ok := GoConfigurableBuildFlags[k]; ok {
			if e, ok := g.Entry[k]; ok {
				if e.Value.IsTrue() {
					a = a.Push("-" + k)
				}
			}
		}
	}
	return a, nil
}

func (g *GoConfig) pushBuildLDFlag(c *GoCommand, a sos.SoS) (sos.SoS, error) {
	var ldflags, space string
	for x := g.Begin; x != ""; x = g.Entry[x].next {
		if IsGoFlag(x) {
			continue
		}
		xv := g.Entry[x].Value
		if s := xv.String(); xv.IsString() && s != "" {
			name := x
			if name[0] == '.' {
				name = "_" +
					filepath.Clean(filepath.Join(g.Dir, x))
			}
			ldflags += space + `-X ` + name + ` "` + s + `"`
			space = " "
		}
	}
	if s, ok := c.StringFlags["ldflags"]; ok && s != "" {
		ldflags += space + s
	} else if e, ok := g.Entry["ldflags"]; ok {
		if s := e.Value.String(); s != "" {
			ldflags += space + s
		}
	}
	if ldflags != "" {
		a = a.Push("-ldflags", ldflags)
	}
	return a, nil
}

func (g *GoConfig) pushBuildStringFlags(c *GoCommand, a sos.SoS) (sos.SoS,
	error) {
	for k := range GoBuildStringFlags {
		// concatenate ldflags with build strings later
		if k != "ldflags" {
			if s, ok := c.StringFlags[k]; ok && s != "" {
				a = a.Push("-"+k, s)
			} else {
				_, ok := GoConfigurableBuildStringFlags[k]
				if ok {
					if e, ok := g.Entry[k]; ok {
						s = e.Value.String()
						if s != "" {
							a = a.Push("-"+k, s)
						}
					}
				}
			}
		}
	}
	return a, nil
}

func (g *GoConfig) pushBuildTagsFlag(c *GoCommand, a sos.SoS) (sos.SoS, error) {
	var tags, space string
	for k := g.Begin; k != ""; k = g.Entry[k].next {
		if _, ok := GoConfigurableBuildFlags[k]; ok {
			continue
		}
		if g.Entry[k].Value.IsTrue() {
			tags += space + k
			space = " "
		}
	}
	if tags != "" {
		a = a.Push("-tags", tags)
	}
	return a, nil
}

func (g *GoConfig) pushSubject(c *GoCommand, a sos.SoS) (sos.SoS, error) {
	if c.Name == "run" {
		fname := filepath.Join(g.Dir, "main.go")
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			fname = filepath.Join(g.Dir,
				filepath.Base(g.Package)+".go")
		}
		if pwd, err := os.Getwd(); err != nil {
			return a, err
		} else if fname, err = filepath.Rel(pwd, fname); err != nil {
			return a, err
		} else {
			a = a.Push(fname)
		}
	} else {
		a = a.Push(g.Package)
	}
	return a, nil
}

func (g *GoConfig) pushTestFlags(c *GoCommand, a sos.SoS) (sos.SoS, error) {
	if c.Name == "test" {
		for k := range GoTestFlags {
			if t, ok := c.Flags[k]; ok && t {
				a = a.Push("-" + k)
			}
		}
	}
	return a, nil
}

func (g *GoConfig) Reinit() {
	for _, e := range g.Entries {
		e.Reinit()
	}
}

func (g *GoConfig) search(pkg string) (string, error) {
	for _, base := range goconfigs {
		rel := filepath.Join(pkg, base)
		if _, err := os.Stat(rel); err == nil {
			return rel, nil
		}
		for _, path := range filepath.SplitList(os.Getenv("GOPATH")) {
			x := filepath.Join(path, "src", rel)
			if _, err := os.Stat(x); err == nil {
				return x, nil
			}
		}
	}
	return "", fmt.Errorf("can't find %s",
		filepath.Join(g.Package, goconfig))
}

func (g *GoConfig) Set(name string, s string) bool {
	var postXset map[string]*Union
	if e, ok := g.Entry[name]; ok {
		if e.Value.IsTag() {
			if t, err := strconv.ParseBool(s); err == nil {
				if t {
					e.Value.SetTrue()
					postXset = e.Set
				} else {
					e.Value.SetFalse()
					postXset = e.Reset
				}
			}
		} else if s == "" {
			e.Value.SetString("")
			postXset = e.Reset
		} else {
			n := len(s) - 1
			r0, rn := s[0], s[n]
			if (r0 == '"' && rn == '"') ||
				(r0 == '\'' && rn == '\'') {
				e.Value.SetString(s[1:n])
			} else {
				e.Value.SetString(s)
			}
			postXset = e.Set
		}
		if len(postXset) > 0 {
			for k, v := range postXset {
				ke := g.Entry[k]
				ke.Value.Copy(v)
			}
			return true
		}
	} else {
		fixme.Println(name, "not found")
	}
	return false
}

func (g *GoConfig) Store() error {
	w, err := os.Create(filepath.Join(g.Dir, goconfiguration))
	if err != nil {
		return err
	}
	defer w.Close()
	for s := g.Begin; s != ""; s = g.Entry[s].next {
		e := g.Entry[s]
		v := e.Value
		if v.IsTrue() ||
			(!v.IsFalse() &&
				(v.String() != "" ||
					e.Init.String() != "")) {
			fmt.Fprintf(w, "%s: %s\n", s, e.Value.YAML())
		}
	}
	return nil
}

func (g *GoConfig) unmarshal(pkg string) error {
	importList := make([]string, 0)
	if full, err := g.search(pkg); err != nil {
		return err
	} else if d, err := filepath.Abs(filepath.Dir(full)); err != nil {
		return err
	} else {
		g.Dir = d
		g.GoConfiguration = filepath.Join(g.Dir, goconfiguration)
	}
	buf := new(bytes.Buffer)
	for _, base := range goconfigs {
		full := filepath.Join(g.Dir, base)
		if file, err := os.Open(full); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		} else {
			buf.Reset()
			_, err = buf.ReadFrom(file)
			file.Close()
			if err != nil {
				return err
			}
		}
		for _, f := range []func([]byte, *[]string) error{
			g.unmarshal1, g.unmarshal2, g.unmarshal3, g.unmarshal4,
		} {
			if err := f(buf.Bytes(), &importList); err != nil {
				return err
			}
		}
		for _, s := range importList {
			if err := g.importer(s); err != nil {
				return err
			}
		}
	}
	return nil
}

// first pass for map of entries with mapped fields; i.e.:
//	a:
//	    init: true
//	    set:
//	        b: false
//	b:
//	    init: false
//	    set:
//	        a: false
func (g *GoConfig) unmarshal1(buf []byte, _ *[]string) error {
	m := make(map[string]*Entry)
	if err := goyaml.Unmarshal(buf, m); err != nil {
		return fmt.Errorf("first pass %s %v", g.Package, err)
	}
	for name, x := range m {
		if name != "import" && !g.Has(name) {
			if x == nil {
				x = new(Entry)
			}
			g.Entry[name] = x
			x.Reinit()
			g.insert(name)
		} else {
			delete(m, name)
		}
	}
	return nil
}

// second pass for simple map entries; i.e.:
//	t: false
//	s: hello world
// Also,
//	import: foo
func (g *GoConfig) unmarshal2(buf []byte, p *[]string) error {
	m := make(map[string]*Union)
	if err := goyaml.Unmarshal(buf, m); err != nil {
		return fmt.Errorf("second pass of %s %v", g.Package, err)
	}
	for name, x := range m {
		if name == "import" {
			*p = append(*p, x.String())
		} else if _, ok := g.Entry[name]; !ok {
			e := new(Entry)
			g.Entry[name] = e
			if x == nil {
				e.Init = NewUnion("")
			} else {
				e.Init = x
			}
			e.Reinit()
			g.insert(name)
		}
		delete(m, name)
	}
	return nil
}

// third pass for empty strings; i.e.:
//	s: ""
func (g *GoConfig) unmarshal3(buf []byte, p *[]string) error {
	m := make(map[string]string)
	if err := goyaml.Unmarshal(buf, m); err != nil {
		return fmt.Errorf("fourth pass of %s %v", g.Package, err)
	}
	for name, x := range m {
		if name == "import" {
			*p = append(*p, x)
		} else if _, ok := g.Entry[name]; !ok {
			e := new(Entry)
			g.Entry[name] = e
			e.Init = NewUnion(x)
			e.Reinit()
			g.insert(name)
		}
		delete(m, name)
	}
	return nil
}

// fourth pass to gather import list, i.e.:
//	import: [
//		import1.yaml,
//		alt import2.yaml,
//		. import3.yaml
//	]
func (g *GoConfig) unmarshal4(buf []byte, p *[]string) error {
	m := make(map[string][]string)
	if err := goyaml.Unmarshal(buf, m); err != nil {
		return fmt.Errorf("third pass of %s %v", g.Package, err)
	}
	if a, ok := m["import"]; ok {
		for _, s := range a {
			*p = append(*p, s)
		}
	}
	return nil
}

func AddMenu(name string, f func(interface{}) error) {
	mutex.Lock()
	if len(Menu) == 0 {
		Menu = make(map[string]func(interface{}) error)
	}
	Menu[name] = f
	mutex.Unlock()
}
