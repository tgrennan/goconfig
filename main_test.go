// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// go test [-v] gopkg.in/tgrennan/goconfig.v0 [-- -fixme[=FILE]]

import (
	"bytes"
	"fmt"
	"gopkg.in/tgrennan/fixme.v0"
	"gopkg.in/tgrennan/sos.v0"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const (
	redirectStderr = " 2>&1"
)

var (
	testLog    *log.Logger
	testOutBuf *bytes.Buffer
	testOutC   chan int
	testInC    chan int
	failures   = 0
)

func init() {
	if a, flag := sos.SoS(os.Args).Flag("fixme"); flag {
		fixme.Enable()
	} else if _, name := a.Arg("fixme"); name != "" {
		if file, err := os.Create(name); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		} else {
			fixme.SetWriter(file)
		}
		fixme.Enable()
	}
	exit = func(_ int) {}
	testLog = log.New(os.Stderr, "", log.Lshortfile)
	testOutBuf = new(bytes.Buffer)
	testOutC = make(chan int)
	testInC = make(chan int)
}

func TestMain(t *testing.T) {
	test(`goconfig -help`, `
Usage:
    goconfig*`)
	test(`goconfig -foo 2>&1`, `
goconfig: invalid flag: -foo`)
	test(`goconfig -http foo.org wont_find 2>&1`, `
goconfig: invalid service address: foo.org`)
	test(`goconfig -http foo.org:FOO examples/wont_find 2>&1`, `
goconfig: strconv.ParseInt: parsing "FOO": invalid syntax`)
	test(`goconfig -http :FOO examples/wont_find 2>&1`, `
goconfig: strconv.ParseInt: parsing "FOO": invalid syntax`)
	test(`goconfig show examples/wont_find 2>&1`, `
goconfig: can't find examples/wont_find/goconfig.yaml`)
	test(`goconfig show -all ./examples/empty`, "")
	test(`goconfig build -o a.out ./examples/empty`, "")
	test("./a.out", "")
	test(`goconfig show -all ./examples/simple`, `
t1: false
t2: true
main.s1: The quick brown fox
main.s2: ""`)
	test("goconfig build -o a.out ./examples/simple", "")
	test("./a.out", `
t1: false
t2: true
main.s1: The quick brown fox
main.s2:`)
	test(`goconfig -config show -all ./examples/simple<
t1: true`, `
t1: true
t2: true
main.s1: The quick brown fox
main.s2: ""`)
	test(`goconfig -config build -o a.out ./examples/simple<
t1: true
`, "")
	test("./a.out", `
t1: true
t2: true
main.s1: The quick brown fox
main.s2:`)
	test(`goconfig -config show -all ./examples/simple<
t1: true
t2: false`, `
t1: true
t2: false
main.s1: The quick brown fox
main.s2: ""`)
	test(`goconfig -config build -o a.out ./examples/simple<
t1: true
t2: false`, "")
	test("./a.out", `
t1: true
t2: false
main.s1: The quick brown fox
main.s2:`)
	test(`goconfig -config show -all ./examples/simple<
main.s1: hello world`, `
t1: false
t2: true
main.s1: hello world
main.s2: ""`)
	test(`goconfig -config build -o a.out ./examples/simple<
main.s1: hello world`, "")
	test("./a.out", `
t1: false
t2: true
main.s1: hello world
main.s2:`)
	test(`goconfig -config show -all ./examples/simple<
main.s1: ""
main.s2: hello world`, `
t1: false
t2: true
main.s1: ""
main.s2: hello world`)
	test(`goconfig -config build -o a.out ./examples/simple<
main.s1: ""
main.s2: hello world`, "")
	test("./a.out", `
t1: false
t2: true
main.s1: 
main.s2: hello world`)
	test(`goconfig show -all ./examples/exclusive`, `
t1: true
t2: false
t3: false`)
	test(`goconfig build -o a.out ./examples/exclusive`, "")
	test("./a.out", `
t1: true
t2: false
t3: false`)
	test(`goconfig -config show -all ./examples/exclusive<
t1: false
t2: true`, `
t1: false
t2: true
t3: false`)
	test(`goconfig -config build -o a.out ./examples/exclusive<
t1: false
t2: true`, "")
	test("./a.out", `
t1: false
t2: true
t3: false`)
	test(`goconfig -config show -all ./examples/exclusive<
t1: false
t3: true`, `
t1: false
t2: false
t3: true`)
	test(`goconfig -config build -o a.out ./examples/exclusive<
t1: false
t3: true`, "")
	test("./a.out", `
t1: false
t2: false
t3: true`)
	test(`goconfig -config run examples/importer/rel
`, `
first: false
second: false
third: false
first.S: 
second.S: 
third.S: 
main.s: `)
	test(`goconfig -config show -all examples/importer/rel<
first: true
`, `
first: true
second: false
third: false
../first.S: ""
../second.S: ""
../third.S: ""
main.s: ""`)
	test(`goconfig -config show -all examples/importer/rel<
../first.S: hello world
`, `
first: false
second: false
third: false
../first.S: hello world
../second.S: ""
../third.S: ""
main.s: ""`)
	test(`goconfig -config show -all examples/buildflags<
race: false
`, `
race: false
compiler: gc`)
	test(`goconfig -config run -n examples/buildflags<
race: false
`, `
#
#  go run -n -compiler gc examples/buildflags/buildflags.go
#
*`)
	if failures > 0 {
		t.Fail()
	}
}

func test(cmd, want string) {
	pout := &os.Stdout
	cmd = strings.TrimSpace(cmd)
	if i := strings.Index(cmd, redirectStderr); i > 0 {
		pout = &os.Stderr
		cmd = strings.Replace(cmd, redirectStderr, "", 1)
	}
	savedOut := *pout
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		testLog.Println("error:", err)
		return
	}
	*pout = outWriter
	testOutBuf.Reset()
	go func() {
		_, err = testOutBuf.ReadFrom(outReader)
		outReader.Close()
		if err != nil {
			testLog.Println("error:", err)
		}
		testOutC <- 1
	}()
	defer func() {
		outWriter.Close()
		*pout = savedOut
		<-testOutC
		got := strings.TrimSpace(testOutBuf.String())
		want = strings.TrimSpace(want)
		var match bool
		if strings.ContainsAny(want, "*?") {
			r := want
			if strings.Contains(want, "\n") {
				r = "(?s)" + want
				r = strings.Replace(r, ".", "\\.", -1)
				r = strings.Replace(r, "\n", ".", -1)
			}
			match = regexp.MustCompile(r).Match(testOutBuf.Bytes())
		} else {
			match = got == want
		}
		if !match {
			testLog.Output(3, fmt.Sprintln("failed:",
				markup(want, "<"), "\n---", markup(got, ">")))
			failures += 1
		} else if testing.Verbose() {
			testLog.Output(3, "passed")
		}
	}()
	if i := strings.Index(cmd, "<\n"); i > 0 {
		os.Args = strings.Split(cmd[:i], " ")
		inBuf := bytes.NewBuffer([]byte(cmd[i+2:]))
		savedStdin := os.Stdin
		inReader, inWriter, err := os.Pipe()
		if err != nil {
			testLog.Println("error:", err)
			return
		}
		os.Stdin = inReader
		go func() {
			_, err := inBuf.WriteTo(inWriter)
			inWriter.Close()
			testInC <- 1
			if err != nil {
				testLog.Println("error:", err)
			}
		}()
		defer func() {
			<-testInC
			inReader.Close()
			os.Stdin = savedStdin
		}()
	} else {
		os.Args = strings.Split(cmd, " ")
	}
	if os.Args[0] == "goconfig" {
		main()
	} else {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		if err := cmd.Run(); err != nil {
			fixme.Println("error:", err)
		}
	}
}

func markup(s, mark string) string {
	return "\n" + mark + strings.Replace(s, "\n", "\n"+mark, -1)
}
