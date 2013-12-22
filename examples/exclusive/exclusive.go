// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is demonstration of a package with exclusive set of build tags.
package main

var t1, t2, t3 bool

func main() {
	print(
		"t1: ", t1, "\n",
		"t2: ", t2, "\n",
		"t3: ", t3, "\n",
	)
}
