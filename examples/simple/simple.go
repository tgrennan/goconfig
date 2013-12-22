// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is simple demonstration of a package with pairs of configurable tags
// and strings.
package main

var t1, t2 bool
var s1, s2 string

func main() {
	print(
		"t1: ", t1, "\n",
		"t2: ", t2, "\n",
		"main.s1: ", s1, "\n",
		"main.s2: ", s2, "\n",
	)
}
