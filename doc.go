// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Goconfig is a utility to configure GO packages in the style of the Linux kernel
whereby the developer declares configurable parameters and the user runs this
much like `make menuconfig && make ...` to configure, build, test, and install.
Goconfig has CLI, TUI and webserver menus.

Goconfig should be functional on any GO supported OS/ARCH but is primarily
developed and tested with `linux_amd64`.  Similarly, the webserver is only
tested with Chrome, Safari, w3m, and elinks; however, its basic HTML and CSS
should be compatible with most browsers.

Files

Goconfig reads parameter declarations from these files within the top source
directory of the subject package.  With duplicate declarations, goconfig has
this top to bottom precedence.

	goconfig_GOOS_GOARCH.yaml
	goconfig_GOARCH.yaml
	goconfig_GOOS.yaml
	goconfig.yaml

Goconfig loads and stores configured parameters with a file named,

	goconfiguration_GOOS_GOARCH.yaml

Configurable Parameters

Configurable build flags, constraints (aka. tags), and strings may be declared
by simple mapped entries with the type implied from the initialized value. So,
the parameter is a boolean flag or tag if initialize with true or false; and a
string otherwise.

	t1: true
	t2: false
	main.s: hello world

You may also declare parameters with mapped fields where the type is implied
from the value of the `init` field.

	t1:
	    init: true
	t2:
	    init: false

Configurable Build Flags

You may declare a boolean `race` flag and these string flags: ccflags,
compiler, gccgoflags, gcflags, installsuffix, ldflags.

Since `race` is only supported on `amd64`, you should declare this in either of
these architectural specific files:

	goconfig_amd64.yaml
	goconfig_linux_amd64.yaml

Configurable Build Constraints

Build constraints or tags are declared with simple or `init` field boolean
values.

	t1: false
	t2:
	    init: false

You may also initialize constraints from a command result using either
`!!status` or `!!not-status` pseudo-types like these,

	TUI: !!status go list code.google.com/p/goncurses
	noTUI: !!not-status go list code.google.com/p/goncurses

Configurable Strings

You must preface string names with the GOPATH or relative package name;
for example:

	main.Hello: hello world
	github.com/tgrennan/fixme.prefix: "FIXME"
	../fixme.prefix: "FIXME"

You may initialize strings with command output using an `!!output` pseudo-type
like this,

	main.Version: !!output git describe --tags

Mapped Fields

Goconfig accepts these mapped declaration fields:

	init, help, choices, set, reset

As stated above, `init` is the initial value that implies the parameter type.

The `help` field is text displayed by the respective menu mode. You may the
include YAML '|' and '>' scalar indicators to preserve or modify formatting.

The set of permitted strings values are declared with a `choices` list:

	main.bufsize:
	    init: "2K"
	    choices: [ "2K", "4K", "8k" ]

The CLI, TUI, and webserer menus present pulldown selectors to limit strings to
these values. However, the package should validate all configurable parameters.

The `set` and `reset` fields are mapped values that get applied when the
parameter is set or reset through the respective menu. Use this to define
exclusive or dependent relationships between parameters.

	t1:
	    init: true
	    set:
	        t2: false
	        t3: false
	    reset:
	        t2: true
	t2:
	    init: false
	    set:
	        t1: false
	        t3: false
	    reset:
	        t1: true
	t3:
	    init: false
	    set:
	        t1: false
	        t2: false
	    reset:
	        t1: true

Import

You may import declarations from one or more dependent packages like these,

	import: github.com/tgrennan/goconfig/examples/importer/first

	import:
	    - github.com/tgrennan/goconfig/examples/importer/first
	    - github.com/tgrennan/goconfig/examples/importer/second
	    - github.com/tgrennan/goconfig/examples/importer/third

	import: [
	    github.com/tgrennan/goconfig/examples/importer/first,
	    github.com/tgrennan/goconfig/examples/importer/second,
	    github.com/tgrennan/goconfig/examples/importer/third
	]

or make relative import of local packages,

	import: [ ../first, ../second, ../third ]

References

...
    https://github.com/tgrennan/goconfig
    http://yaml.org/refcard.html
    http://en.wikipedia.org/wiki/YAML

*/
package main
