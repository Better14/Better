// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

type nilablePointersMode int

const (
	nptDisable nilablePointersMode = iota
	nptWarn
	nptEnable
)

func parseNilablePointersMode(s string) nilablePointersMode {
	switch s {
	case "enable":
		return nptEnable
	case "warn":
		return nptWarn
	default:
		return nptDisable
	}
}

func (check *Checker) nilablePointersMode() nilablePointersMode {
	if check == nil {
		return nptDisable
	}
	return check.nilablePointers
}

func (check *Checker) nilablePointersOn() bool {
	return check.nilablePointersMode() != nptDisable
}

// isNilablePointerType reports whether t is *T? (Optional wrapping a pointer).
func isNilablePointerType(t Type) bool {
	_, ok := nilablePointerElem(t)
	return ok
}

// nilablePointerElem returns *T when t is *T?, or nil.
func nilablePointerElem(t Type) (Type, bool) {
	if t == nil {
		return nil, false
	}
	o, ok := t.Underlying().(*Optional)
	if !ok {
		return nil, false
	}
	if _, ok := o.elem.Underlying().(*Pointer); !ok {
		return nil, false
	}
	return o.elem, true
}

// isStrictPointerType reports whether t is a plain *T (not *T?).
func isStrictPointerType(t Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*Pointer)
	return ok
}

// nilablePointerType returns *T? for plain pointer typ when NPT is on.
func (check *Checker) nilablePointerType(typ Type) Type {
	if p, ok := typ.Underlying().(*Pointer); ok && check.nilablePointersOn() {
		return NewOptional(p)
	}
	return typ
}

// collapseNilablePointerType returns *T when NPT is off and typ is *T?.
func (check *Checker) collapseNilablePointerType(typ Type) Type {
	if elem, ok := nilablePointerElem(typ); ok && !check.nilablePointersOn() {
		return elem
	}
	return typ
}

func (check *Checker) reportNilToStrictPointer(at poser, T Type) {
	msg := check.sprintf("cannot use nil as %s value", T)
	switch check.nilablePointersMode() {
	case nptEnable:
		check.errorf(at, IncompatibleAssign, "%s", msg)
	default:
		check.softErrorf(at, IncompatibleAssign, "%s", msg)
	}
}

func (check *Checker) reportNilablePointerMismatch(at poser, V, T Type, context string) {
	msg := check.sprintf("cannot use %s as %s value in %s", V, T, context)
	switch check.nilablePointersMode() {
	case nptEnable:
		check.errorf(at, IncompatibleAssign, "%s", msg)
	default:
		check.softErrorf(at, IncompatibleAssign, "%s", msg)
	}
}

// fileNilablePointersMode returns the effective mode for a file (file > package config).
func (check *Checker) fileNilablePointersMode(file *syntax.File) nilablePointersMode {
	if file != nil && file.NilablePointers != "" {
		return parseNilablePointersMode(file.NilablePointers)
	}
	return parseNilablePointersMode(check.conf.NilablePointers)
}
