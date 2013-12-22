## goconfig - usage
```
Usage:
    goconfig [flags] [-cli] [package]
    goconfig [flags] -http=<server:port>
    goconfig [flags] -show [-all] [package]
    goconfig [flags] <[go] command> [go flags] [package [args]]

Flags:
   -fixme[=<file>]
    Print debugging messages on stderr or in the given file.

   -config[=<file>]
    Load configuration from stdin or the given file instead of the default,
    goconfiguration_GOOS_GOARCH.yaml

Options:
   -cli
    The default action is a terminal user interface unless running on a
    DUMB terminal or this flag is given to revert to a command line
    interface.

   -http=<server:port>
    Runs a web server at the given address instead of a TUI or CLI.

   -show [-all]
    Instead of a menu, print the configured [or all] entries.

    go <command> [build and test flags] [package [args]]
    Run the given command with the configured constraints and strings.

Goconfig operates on one package per execution unless given "all"
where it makes a menu of packages within GOPATH containing:
    goconfig[_GOOS][_GOARCH].yaml
```
