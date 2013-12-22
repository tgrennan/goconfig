// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is demonstration of a package with GOPATH import of other package
// configurable parameter declarations.
package main

import (
	"github.com/tgrennan/goconfig/examples/importer/first"
	deuxième "github.com/tgrennan/goconfig/examples/importer/second"
	. "github.com/tgrennan/goconfig/examples/importer/third"
)

var s string

func main() {
	println("first:", first.First)
	println("second:", deuxième.Second)
	println("third:", Third)
	println("first.S:", first.S)
	println("second.S:", deuxième.S)
	println("third.S:", S)
	println("main.s:", s)
}
