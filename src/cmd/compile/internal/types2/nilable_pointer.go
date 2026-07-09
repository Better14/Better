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

func (check *Checker) moduleNilablePointersMode() nilablePointersMode {
	if check != nil && check.pkg != nil {
		return parseNilablePointersMode(check.pkg.nilablePointers)
	}
	return nptDisable
}

func (check *Checker) nilablePointersModeAt(pos syntax.Pos) nilablePointersMode {
	if check == nil {
		return nptDisable
	}
	if pos.IsKnown() {
		if f := check.fileAt(pos); f != nil {
			for _, r := range f.NilablePointersRegions {
				if pos.Cmp(r.Start) >= 0 && (!r.End.IsKnown() || pos.Cmp(r.End) < 0) {
					return parseNilablePointersMode(r.Mode)
				}
			}
		}
	}
	return check.moduleNilablePointersMode()
}

func (check *Checker) nilablePointersOnAt(pos syntax.Pos) bool {
	return check.nilablePointersModeAt(pos) != nptDisable
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

func isStrictPointerType(t Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*Pointer)
	return ok
}

func (check *Checker) collapseNilablePointerType(pos syntax.Pos, typ Type) Type {
	if elem, ok := nilablePointerElem(typ); ok && !check.nilablePointersOnAt(pos) {
		return elem
	}
	return typ
}

func (check *Checker) reportNilToStrictPointer(at poser, T Type) {
	msg := check.sprintf("cannot use nil as %s value", T)
	switch check.nilablePointersModeAt(at.Pos()) {
	case nptEnable:
		check.errorf(at, IncompatibleAssign, "%s", msg)
	default:
		check.softErrorf(at, IncompatibleAssign, "%s", msg)
	}
}

func (check *Checker) fileAt(pos syntax.Pos) *syntax.File {
	for _, f := range check.files {
		if f.Pos().Cmp(pos) <= 0 && pos.Cmp(f.EOF) <= 0 {
			return f
		}
	}
	return nil
}
