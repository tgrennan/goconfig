// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is a demonstration of a package with many configurable strings used to
// test menu scrolling.
package main

import "fmt"

var s000, s001, s002, s003, s004, s005, s006, s007, s008, s009 string
var s010, s011, s012, s013, s014, s015, s016, s017, s018, s019 string
var s020, s021, s022, s023, s024, s025, s026, s027, s028, s029 string
var s030, s031, s032, s033, s034, s035, s036, s037, s038, s039 string
var s040, s041, s042, s043, s044, s045, s046, s047, s048, s049 string
var s050, s051, s052, s053, s054, s055, s056, s057, s058, s059 string
var s060, s061, s062, s063, s064, s065, s066, s067, s068, s069 string
var s070, s071, s072, s073, s074, s075, s076, s077, s078, s079 string
var s080, s081, s082, s083, s084, s085, s086, s087, s088, s089 string
var s090, s091, s092, s093, s094, s095, s096, s097, s098, s099 string

var s100, s101, s102, s103, s104, s105, s106, s107, s108, s109 string
var s110, s111, s112, s113, s114, s115, s116, s117, s118, s119 string
var s120, s121, s122, s123, s124, s125, s126, s127, s128, s129 string
var s130, s131, s132, s133, s134, s135, s136, s137, s138, s139 string
var s140, s141, s142, s143, s144, s145, s146, s147, s148, s149 string
var s150, s151, s152, s153, s154, s155, s156, s157, s158, s159 string
var s160, s161, s162, s163, s164, s165, s166, s167, s168, s169 string
var s170, s171, s172, s173, s174, s175, s176, s177, s178, s179 string
var s180, s181, s182, s183, s184, s185, s186, s187, s188, s189 string
var s190, s191, s192, s193, s194, s195, s196, s197, s198, s199 string

var strings = []*string{
	&s000, &s001, &s002, &s003, &s004, &s005, &s006, &s007, &s008, &s009,
	&s010, &s011, &s012, &s013, &s014, &s015, &s016, &s017, &s018, &s019,
	&s020, &s021, &s022, &s023, &s024, &s025, &s026, &s027, &s028, &s029,
	&s030, &s031, &s032, &s033, &s034, &s035, &s036, &s037, &s038, &s039,
	&s040, &s041, &s042, &s043, &s044, &s045, &s046, &s047, &s048, &s049,
	&s050, &s051, &s052, &s053, &s054, &s055, &s056, &s057, &s058, &s059,
	&s060, &s061, &s062, &s063, &s064, &s065, &s066, &s067, &s068, &s069,
	&s070, &s071, &s072, &s073, &s074, &s075, &s076, &s077, &s078, &s079,
	&s080, &s081, &s082, &s083, &s084, &s085, &s086, &s087, &s088, &s089,
	&s090, &s091, &s092, &s093, &s094, &s095, &s096, &s097, &s098, &s099,

	&s100, &s101, &s102, &s103, &s104, &s105, &s106, &s107, &s108, &s109,
	&s110, &s111, &s112, &s113, &s114, &s115, &s116, &s117, &s118, &s119,
	&s120, &s121, &s122, &s123, &s124, &s125, &s126, &s127, &s128, &s129,
	&s130, &s131, &s132, &s133, &s134, &s135, &s136, &s137, &s138, &s139,
	&s140, &s141, &s142, &s143, &s144, &s145, &s146, &s147, &s148, &s149,
	&s150, &s151, &s152, &s153, &s154, &s155, &s156, &s157, &s158, &s159,
	&s160, &s161, &s162, &s163, &s164, &s165, &s166, &s167, &s168, &s169,
	&s170, &s171, &s172, &s173, &s174, &s175, &s176, &s177, &s178, &s179,
	&s180, &s181, &s182, &s183, &s184, &s185, &s186, &s187, &s188, &s189,
	&s190, &s191, &s192, &s193, &s194, &s195, &s196, &s197, &s198, &s199,
}

func main() {
	for i, p := range strings {
		if *p != "" {
			fmt.Printf("main.s%03d: %s\n", i, *p)
		}
	}
}
