// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Entry struct {
	Init    *Union
	Value   *Union
	Name    string
	Help    string
	Choices []string
	Set     map[string]*Union
	Reset   map[string]*Union
	next    string
	prev    string
}

func (e *Entry) Reinit() {
	if e.Init == nil {
		e.Init = NewUnion("")
	}
	if e.Value == nil {
		e.Value = new(Union)
	}
	e.Value.Copy(e.Init)
}
