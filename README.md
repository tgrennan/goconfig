Command `goconfig` is a utility to configure GO packages in the style of the
Linux kernel whereby the developer declares configurable parameters and the
user runs this much like `make menuconfig && make ...` to configure, build,
test, and install.  Goconfig has CLI, TUI and webserver menus.

Fetch, build and install this package with GO tool:

	go get gopkg.in/tgrennan/goconfig.v0

Import this package with:

	import "gopkg.in/tgrennan/goconfig.v0"

[USAGE](USAGE.md), [FAQ](FAQ.md)

[![GoDoc](https://godoc.org/gopkg.in/tgrennan/goconfig.v0?status.png)](
https://godoc.org/gopkg.in/tgrennan/goconfig.v0)
