// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !noWebServer

package main

import (
	"errors"
	html "html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type wsgT struct { // WebServer GoConfig
	G       *GoConfig
	mutex   *sync.Mutex
	Version int
}

type wshT struct { // WebServer Handler
	Name    string
	String  string
	command string
	path    string
	tmpl    string
	Heading html.HTML
	Body    html.HTML
	Index   int
	status  int
	err     error
	WSG     *wsgT
}

const wsPrefix = "/goconfig/"
const wsSource = `
{{define "__top__"}}
{{$black := "#000000"}}
{{$softwhite := "#D0D0D0"}}
{{$brightred := "#E00000"}}
<!DOCTYPE html>
<html>
<head>
<meta http-equiv="cache-control" content="no-cache">
<meta name="robots" content="none">
<title>goconfig: {{.}}</title>
<style>
a.goconfig {
	background-color: {{$black}};
	color: {{$softwhite}};
	text-decoration: none;
}
body {
	background-color: {{$black}};
	color: {{$softwhite}};
	font-family: Arial,sans-serif;
	font-size: 100%;
}
button {
	background-color: {{$black}};
	color: {{$softwhite}};
	border-size: 2px;
	border-radius: 7px;
	font-family: Arial,sans-serif;
	font-size: 90%;
}
button.entry {
	background-color: {{$black}};
	color: {{$softwhite}};
	border: none;
	font-family: Monaco,monospace;
	font-size: 90%;
	padding: 0 0 0 0;
}
code, tt {
	font-family: Monaco,monospace;
	font-size: 90%;
}
error {
	background-color: {{$black}};
	color: {{$brightred}};
	font-family: Arial,sans-serif;
	font-size: 110%;
}
h1 {
	font-size: 110%;
	font-weight: bold;
	font-family: Monaco,monospace;
}
input.button {
	background-color: {{$black}};
	color: {{$softwhite}};
	border-size: 2px;
	border-radius: 7px;
	font-family: Arial,sans-serif;
	font-size: 90%;
}
input.text {
	background-color: {{$black}};
	color: {{$softwhite}};
	border-size: 2px;
	border-radius: 7px;
	font-family: Monaco,monospace;
	font-size: 90%;
}
</style>
</head>
<body>
<h1>goconfig: {{.}}</h1>
{{end}}

{{define "__bottom__"}}
</body>
</html>
{{end}}

{{define "change"}}
{{template "__top__" .WSG.G.Package}}
<form	method="POST">
<input	type="hidden"
	name="version"
	value="{{.WSG.Version}}">
<code>{{.Name}}</code>:
<input
	class="text"
	name="s"
	type="text"
	size="55"
	value="{{.String}}"
	autofocus><br>
<button	type="submit"
	name="set"
	value="{{.Index}}"
>set</button>
<button	type="submit"
	name="reinitialize"
	value="{{.Index}}"
>reinitialize</button>
<button	type="submit"
	name="cancel"
	value="set"
>cancel</button>
</form>
{{template "__bottom__"}}
{{end}}

{{define "conflict"}}
{{template "__top__" .WSG.G.Package}}
<p><b>Warning!</b></p>
<form	method="POST">
<p>The configuration was changed by another session;<br>
<button	type="submit"
	name="cancel"
	value="results"
	autofocus
>resume</button> to review then resubmit.</p>
</form>
{{template "__bottom__"}}
{{end}}

{{define "go"}}
{{template "__top__" .WSG.G.Package}}
<form	method="POST">
<input	type="hidden"
	name="version"
	value="{{.WSG.Version}}">
go command:
<input	class="text"
	name="s"
	type="text"
	size="55"
	autofocus
><br>
<button
	type="submit"
	name="run"
	value="Go"
>go</button>
<button	type="submit"
	name="cancel"
	value="go"
>cancel</button>
</form>
{{template "__bottom__"}}
{{end}}

{{define "list"}}
{{template "__top__" .WSG.G.Package}}
<p>Select one of these packages:</p>
<p>
{{range $E := .WSG.G.Entries}}
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code><a
	class="goconfig"
	href="/goconfig/{{$E.Name}}"
>{{$E.Name}}</a><br>
{{end}}</p>
{{template "__bottom__"}}
{{end}}

{{define "results"}}
{{template "__top__" .WSG.G.Package}}
{{with .Heading}}{{.}}{{end}}
{{with .Body}}{{.}}{{end}}
<form	method="POST">
<button	type="submit"
	name="cancel"
	value="results"
	autofocus
>resume</button>
</form>
{{template "__bottom__"}}
{{end}}

{{define "view"}}
{{template "__top__" .WSG.G.Package}}
{{$WS := .}}
<p>
<form	method="POST">
<input	type="hidden"
	name="version"
	value="{{.WSG.Version}}">
{{range $I, $E := .WSG.G.Entries}}
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code>
<button	class="entry"
	type="submit"
	name="info"
	value="{{$I}}"
>{{$E.Name}}</button>:
<button	class="entry"
	type="submit"
	name="change"
	value="{{$I}}"
>{{with $E.Value.String}}{{.}}{{else}}nil{{end}}</button><br>
{{end}}
</p>
<p>
go command:
<input	class="text"
	name="go"
	type="text"
	size="55"
	autofocus
></p>
<p>
Select...<br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code>
an entry name for info;<br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code>
a value to change it;<br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code>
<button	type="submit"
	name="gotool"
	value="go"
>go</button>
to execute the given build, test, run, or install command;<br>
<code>&nbsp;&nbsp;&nbsp;&nbsp;</code>
<button	type="submit"
	name="reinitialize"
	value="all"
>reinitialize</button>
or
<button	type="submit"
	name="save"
	value="save"
>save</button>
this package configuration.
</p>
</form>
{{template "__bottom__"}}
{{end}}
`

var (
	wsgMap map[string]*wsgT
	wstmpl *html.Template
)

func init() { AddMenu("webserver", __webserver__) }

func __webserver__(v interface{}) (err error) {
	address := v.(string)
	if wstmpl, err = html.New("ws").Parse(wsSource); err != nil {
		return
	}
	wsg := new(wsgT)
	if wsg.G, err = NewGoConfig(ALL); err != nil {
		return
	}
	wsg.mutex = new(sync.Mutex)
	wsgMap = make(map[string]*wsgT)
	wsgMap[ALL] = wsg
	http.HandleFunc("/", wsHandler)
	err = http.ListenAndServe(address, nil)
	return
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	wsh := &wshT{}
	defer func() {
		if wsh.tmpl != "" {
			wsh.err = wstmpl.ExecuteTemplate(w, wsh.tmpl, wsh)
		}
		if wsh.WSG != nil && wsh.WSG.mutex != nil {
			wsh.WSG.mutex.Unlock()
		}
		if wsh.err != nil {
			if wsh.status < http.StatusBadRequest {
				wsh.status = http.StatusInternalServerError
			}
			http.Error(w, wsh.err.Error(), wsh.status)
		} else if wsh.status == http.StatusNotFound {
			http.NotFound(w, r)
		} else if wsh.status >= http.StatusMultipleChoices {
			http.Redirect(w, r, wsPrefix+wsh.path, wsh.status)
		}
	}()
	if r.URL.Path == "/" {
		wsh.path, wsh.status = ALL, http.StatusFound
		return
	} else if strings.HasPrefix(r.URL.Path, wsPrefix) {
		wsh.path = strings.TrimPrefix(r.URL.Path, wsPrefix)
	} else {
		wsh.status = http.StatusNotFound
		return
	}
	if x, ok := wsgMap[wsh.path]; !ok {
		x = new(wsgT)
		wsh.WSG, wsgMap[wsh.path] = x, x
		if wsh.WSG.G, wsh.err = NewGoConfig(wsh.path); wsh.err != nil {
			wsh.status = http.StatusUnauthorized
			return
		}
		wsh.WSG.mutex = new(sync.Mutex)
	} else {
		wsh.WSG = x
	}
	wsh.WSG.mutex.Lock()
	if wsh.conflict(r) {
		return
	}
	for _, x := range []struct {
		c string
		f func(string, *http.Request)
	}{
		{"cancel", wsh.cancel},
		{"change", wsh.change},
		{"go", wsh.gotool},
		{"info", wsh.info},
		{"reinitialize", wsh.reinitialize},
		{"save", wsh.save},
		{"set", wsh.set},
	} {
		if s := r.FormValue(x.c); s != "" {
			x.f(s, r)
			return
		}
	}
	if wsh.path == ALL {
		wsh.status, wsh.tmpl = http.StatusOK, "list"
	} else {
		wsh.status, wsh.tmpl = http.StatusOK, "view"
	}
	return
}

func (wsh *wshT) entry(s string) {
	if wsh.Index, wsh.err = strconv.Atoi(s); wsh.err != nil {
		wsh.status = http.StatusBadRequest
	} else if wsh.Index >= len(wsh.WSG.G.Entries) {
		wsh.status = http.StatusRequestedRangeNotSatisfiable
		wsh.err = errors.New("index exceeds entries")
	} else {
		wsh.Name = wsh.WSG.G.Entries[wsh.Index].Name
	}
}

func (wsh *wshT) cancel(_ string, _ *http.Request) {
	wsh.status = http.StatusSeeOther
}

func (wsh *wshT) change(s string, _ *http.Request) {
	if wsh.entry(s); wsh.err == nil {
		wsh.status = http.StatusNotFound
		if entry := wsh.WSG.G.Entry[wsh.Name]; entry != nil {
			if value := entry.Value; value != nil {
				wsh.status = http.StatusOK
				if value.IsTrue() {
					wsh.WSG.G.Set(wsh.Name, "false")
					wsh.tmpl = "view"
					wsh.WSG.Version += 1
				} else if value.IsFalse() {
					wsh.WSG.G.Set(wsh.Name, "true")
					wsh.tmpl = "view"
					wsh.WSG.Version += 1
				} else {
					wsh.String = value.String()
					wsh.tmpl = "change"
				}
			}
		}
	}
}

func (wsh *wshT) conflict(r *http.Request) bool {
	if s := r.FormValue("version"); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			if i != wsh.WSG.Version {
				wsh.tmpl = "conflict"
				return true
			}
		}
	}
	return false
}

func (wsh *wshT) gotool(s string, _ *http.Request) {
	wsh.tmpl = "results"
	if b, err := wsh.WSG.G.Exec("go " + s); err != nil {
		wsh.Heading = `<error>Error:</error>`
		wsh.Body = html.HTML(`<pre>` + err.Error() + "\n" +
			string(b) + `</pre>`)
	} else {
		wsh.Heading = `<p>Results:</p>`
		wsh.Body = html.HTML(`<pre>` + string(b) + `</pre>`)
	}
}

func (wsh *wshT) info(s string, _ *http.Request) {
	if wsh.entry(s); wsh.err == nil {
		wsh.Body = html.HTML(`<pre>` +
			wsh.WSG.G.Marshal(wsh.Name) + `</pre>`)
		wsh.status, wsh.tmpl = http.StatusOK, "results"
	}
}

func (wsh *wshT) reinitialize(s string, _ *http.Request) {
	if s == "all" {
		wsh.status, wsh.tmpl = http.StatusOK, "view"
		wsh.WSG.G.Reinit()
	} else if wsh.entry(s); wsh.err == nil {
		wsh.status, wsh.tmpl = http.StatusOK, "view"
		wsh.WSG.G.Entry[wsh.Name].Reinit()
	}
	wsh.WSG.Version += 1
}

func (wsh *wshT) save(_ string, _ *http.Request) {
	wsh.tmpl = "results"
	if err := wsh.WSG.G.Store(); err != nil {
		wsh.Heading = `<error>Error:</error>`
		wsh.Body = html.HTML(`<pre>` + err.Error() + `</pre>`)
		wsh.status = http.StatusUnauthorized
	} else {
		wsh.Body = html.HTML(`<p>Wrote: <code>` +
			wsh.WSG.G.GoConfiguration + `</code></p>`)
		wsh.status = http.StatusOK
	}
}

func (wsh *wshT) set(s string, r *http.Request) {
	wsh.tmpl = "results"
	if wsh.entry(s); wsh.err == nil {
		wsh.WSG.G.Set(wsh.Name, r.FormValue("s"))
		wsh.status, wsh.tmpl = http.StatusOK, "view"
		wsh.WSG.Version += 1
	}
}
