// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"strings"

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

func isNilablePointerType(t Type) bool {
	_, ok := nilablePointerElem(t)
	return ok
}

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

func isStrictPointerType(t Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*Pointer)
	return ok
}

func (check *Checker) collapseNilablePointerType(typ Type) Type {
	if elem, ok := nilablePointerElem(typ); ok && !check.nilablePointersOn() {
		return elem
	}
	return typ
}

func (check *Checker) reportNilToStrictPointer(at positioner, T Type) {
	msg := check.sprintf("cannot use nil as %s value", T)
	switch check.nilablePointersMode() {
	case nptEnable:
		check.errorf(at, IncompatibleAssign, "%s", msg)
	default:
		check.softErrorf(at, IncompatibleAssign, "%s", msg)
	}
}

func (check *Checker) fileNilablePointersMode(file *ast.File) nilablePointersMode {
	if file != nil && file.NilablePointers != "" {
		return parseNilablePointersMode(file.NilablePointers)
	}
	return parseNilablePointersMode(check.conf.NilablePointers)
}

func nilablePointersDirective(file *ast.File) string {
	if file == nil {
		return ""
	}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.HasPrefix(text, "go:nilable_pointers ") {
				f := strings.Fields(text)
				if len(f) == 2 {
					return f[1]
				}
			}
		}
	}
	return ""
}
