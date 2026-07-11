// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
	"sort"
	"strings"

	. "internal/types/errors"
)

type nilablePointersMode int

// nilableSelKey identifies a field selector v.f for nilable narrowing.
type nilableSelKey struct {
	obj Object
	sel string
}

const (
	nptDisable nilablePointersMode = iota
	nptWarn
	nptEnable
)

func parseNilablePointersMode(s string) nilablePointersMode {
	switch s {
	case "enable":
		return nptEnable
	case "warnings":
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

func (check *Checker) nilablePointersModeAt(pos token.Pos) nilablePointersMode {
	if check == nil {
		return nptDisable
	}
	if pos.IsValid() {
		if f := check.fileAt(pos); f != nil {
			for _, r := range f.NilablePointersRegions {
				if pos >= r.Start && (!r.End.IsValid() || pos < r.End) {
					return parseNilablePointersMode(r.Mode)
				}
			}
		}
	}
	return check.moduleNilablePointersMode()
}

func (check *Checker) nilablePointersOnAt(pos token.Pos) bool {
	return check.nilablePointersModeAt(pos) != nptDisable
}

func isNilablePointerType(t Type) bool {
	_, ok := nilablePointerElem(t)
	return ok
}

// nilablePointerElem returns T when t is T? and T is a pointer, slice, map, or channel type.
func nilablePointerElem(t Type) (Type, bool) {
	if t == nil {
		return nil, false
	}
	o, ok := t.Underlying().(*Optional)
	if !ok {
		return nil, false
	}
	switch o.elem.Underlying().(type) {
	case *Pointer, *Slice, *Map, *Chan:
		return o.elem, true
	default:
		return nil, false
	}
}

// isStrictPointerType reports whether t is a non-optional pointer, slice, map, or channel.
func isStrictPointerType(t Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *Pointer, *Slice, *Map, *Chan:
		return true
	default:
		return false
	}
}

func (check *Checker) collapseNilablePointerType(pos token.Pos, typ Type) Type {
	if elem, ok := nilablePointerElem(typ); ok && !check.nilablePointersOnAt(pos) {
		return elem
	}
	return typ
}

func (check *Checker) reportNilToStrictPointer(at positioner, T Type) {
	msg := check.sprintf("cannot use nil as %s value", T)
	switch check.nilablePointersModeAt(at.Pos()) {
	case nptEnable:
		check.errorf(at, IncompatibleAssign, "%s", msg)
	default:
		check.softErrorf(at, IncompatibleAssign, "%s", msg)
	}
}

func (check *Checker) reportNilableUseWithoutNilCheck(at positioner, t Type, verb string, code Code) {
	msg := check.sprintf("cannot %s %s without nil check", verb, t)
	switch check.nilablePointersModeAt(at.Pos()) {
	case nptEnable:
		check.errorf(at, code, invalidOp+"%s", msg)
	default:
		check.softErrorf(at, code, invalidOp+"%s", msg)
	}
}

func (check *Checker) fileAt(pos token.Pos) *ast.File {
	for _, f := range check.files {
		if f.FileStart <= pos && pos <= f.FileEnd {
			return f
		}
	}
	return nil
}

type nilablePointersDirective struct {
	pos  token.Pos
	mode string // enable, disable, warnings, end
}

func collectNilablePointersDirectives(file *ast.File) []nilablePointersDirective {
	if file == nil {
		return nil
	}
	var dirs []nilablePointersDirective
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if !strings.HasPrefix(text, "go:nilable_pointers ") {
				continue
			}
			f := strings.Fields(text)
			if len(f) != 2 {
				continue
			}
			switch f[1] {
			case "enable", "disable", "warnings", "end":
				dirs = append(dirs, nilablePointersDirective{pos: c.Pos(), mode: f[1]})
			}
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].pos < dirs[j].pos
	})
	return dirs
}

func buildNilablePointersRegions(dirs []nilablePointersDirective) []ast.NilablePointersRegion {
	var regions []ast.NilablePointersRegion
	var open *ast.NilablePointersRegion
	for _, d := range dirs {
		switch d.mode {
		case "end":
			if open != nil {
				open.End = d.pos
				regions = append(regions, *open)
				open = nil
			}
		case "enable", "disable", "warnings":
			if open != nil {
				open.End = d.pos
				regions = append(regions, *open)
			}
			open = &ast.NilablePointersRegion{Start: d.pos, Mode: d.mode}
		}
	}
	if open != nil {
		regions = append(regions, *open)
	}
	return regions
}
