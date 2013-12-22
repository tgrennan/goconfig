// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !noTUI

package main

const (
	ctrlSpace = iota
	ctrlA
	ctrlB
	ctrlC
	ctrlD
	ctrlE
	ctrlF
	ctrlG
	ctrlH
	ctrlI
	ctrlJ
	ctrlK
	ctrlL
	ctrlM
	ctrlN
	ctrlO
	ctrlP
	ctrlQ
	ctrlR
	ctrlS
	ctrlT
	ctrlU
	ctrlV
	ctrlW
	ctrlX
	ctrlY
	ctrlZ
)

const (
	metaA = iota + '\u00a0'
	metaB
	metaC
	metaD
	metaE
	metaF
	metaG
	metaH
	metaI
	metaJ
	metaK
	metaL
	metaM
	metaN
	metaO
	metaP
	metaQ
	metaR
	metaS
	metaT
	metaU
	metaV
	metaW
	metaX
	metaY
	metaZ

	metaLT = '\u00bc'
	metaGT = '\u00be'
)

const (
	shiftTab = '\u0161'
)
