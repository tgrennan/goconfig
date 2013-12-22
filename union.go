// Copyright 2014 Tom Grennan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/tgrennan/quotation"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

const (
	SetOutput = iota
	SetStatus
	SetNotStatus
)

type Union struct{ p unsafe.Pointer }

func NewUnion(v interface{}) *Union {
	u := new(Union)
	switch t := v.(type) {
	case string:
		u.SetString(t)
	case bool:
		if t {
			u.SetTrue()
		} else {
			u.SetFalse()
		}
	}
	return u
}

func (u *Union) Copy(x *Union) {
	if x.IsTrue() {
		u.SetTrue()
	} else if x.IsFalse() {
		u.SetFalse()
	} else {
		u.SetString(x.String())
	}
}

func (u *Union) Equal(x *Union) bool {
	if u.IsTrue() {
		return x.IsTrue()
	} else if u.IsFalse() {
		return x.IsFalse()
	}
	return u.String() == x.String()
}

func (u *Union) IsFalse() bool {
	return u.p == nil
}

func (u *Union) IsString() bool {
	return u.p != nil && u.p != unsafe.Pointer(&veritas)
}

func (u *Union) IsTag() bool {
	return u.IsTrue() || u.IsFalse()
}

func (u *Union) IsTrue() bool {
	return u.p == unsafe.Pointer(&veritas)
}

func (u *Union) Set(v interface{}) {
	switch t := v.(type) {
	case nil:
		if u.IsTrue() {
			u.SetFalse()
		} else {
			u.SetString("")
		}
	case string:
		u.SetString(t)
	case bool:
		if t {
			u.SetTrue()
		} else {
			u.SetFalse()
		}
	}
}

func (u *Union) SetExec(set int, cmd string) {
	a := quotation.Fields(cmd)
	buf, err := exec.Command(a[0], a[1:]...).CombinedOutput()
	t := strings.TrimSpace(string(buf))
	switch set {
	case SetOutput:
		if err == nil {
			u.SetString(t)
		} else {
			u.SetString("")
		}
	case SetStatus:
		if err == nil {
			u.SetTrue()
		} else {
			u.SetFalse()
		}
	case SetNotStatus:
		if err == nil {
			u.SetFalse()
		} else {
			u.SetTrue()
		}
	}
}

func (u *Union) SetFalse() {
	u.p = nil
}

func (u *Union) SetString(s string) {
	u.p = (unsafe.Pointer(new(string)))
	*(*string)(u.p) = s
}

func (u *Union) SetTrue() {
	u.p = unsafe.Pointer(&veritas)
}

func (u *Union) SetYAML(t string, v interface{}) bool {
	switch t {
	case "!!bool":
		if v.(bool) {
			u.SetTrue()
		} else {
			u.SetFalse()
		}
	case "!!status":
		u.SetExec(SetStatus, v.(string))
	case "!!not-status":
		u.SetExec(SetNotStatus, v.(string))
	case "!!null":
		u.SetString("")
	case "!!str":
		u.SetString(v.(string))
	case "!!output":
		u.SetExec(SetOutput, v.(string))
	default:
		return false
	}
	return true
}

func (u *Union) String() string {
	if u.IsTag() {
		return strconv.FormatBool(u.IsTrue())
	}
	return *(*string)(u.p)
}

func (u *Union) YAML() string {
	if u.IsTrue() {
		return "true"
	}
	if u.IsFalse() {
		return "false"
	}
	s := u.String()
	if s == "" {
		return `""`
	}
	for _, x := range []string{
		"False", "false", "f",
		"True", "true", "t",
	} {
		if s == x {
			return `"` + s + `"`
		}
	}
	return s
}
