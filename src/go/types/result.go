// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	. "internal/types/errors"
)

// A Result represents a result value type T! (value + error).
type Result struct {
	elem Type
}

// NewResult returns a new result value type for the given element type.
func NewResult(elem Type) *Result { return &Result{elem: elem} }

// Elem returns the element type of r.
func (r *Result) Elem() Type { return r.elem }

func (r *Result) Underlying() Type { return r }
func (r *Result) String() string   { return TypeString(r, nil) }

// AsResult reports whether t is a *Result and returns it.
func AsResult(t Type) (*Result, bool) {
	r, _ := t.(*Result)
	return r, r != nil
}

// ResultStruct returns the struct type used to lower T!.
func ResultStruct(pkg *Package, res *Result) *Struct {
	pos := nopos
	return NewStruct([]*Var{
		newVar(FieldVar, pos, pkg, "value", res.elem),
		newVar(FieldVar, pos, pkg, "err", universeError),
	}, nil)
}

// canForceReturn reports whether ! may early-return an error from the current function.
func (check *Checker) canForceReturn() bool {
	if check.sig == nil {
		return false
	}
	res := check.sig.Results()
	if res == nil {
		return false
	}
	switch res.Len() {
	case 1:
		return Identical(res.At(0).Type(), universeError)
	case 2:
		return Identical(res.At(1).Type(), universeError)
	default:
		return false
	}
}

// checkForceReturn verifies the enclosing function can propagate errors for exprTyp,
// which must be error or a 2-tuple (T, error).
func (check *Checker) checkForceReturn(at positioner, exprTyp Type) bool {
	if check.canForceReturnFor(exprTyp) {
		return true
	}
	check.errorf(at, InvalidSyntaxTree, "invalid operation: ! requires enclosing function whose return type needs to be %s", check.forceReturnWant(exprTyp))
	return false
}

func (check *Checker) canForceReturnFor(exprTyp Type) bool {
	if check.sig == nil {
		return false
	}
	res := check.sig.Results()
	if res == nil || res.Len() == 0 {
		return false
	}

	if Identical(exprTyp, universeError) {
		switch res.Len() {
		case 1:
			return Identical(res.At(0).Type(), universeError)
		case 2:
			return Identical(res.At(1).Type(), universeError)
		default:
			return false
		}
	}

	tup, ok := exprTyp.(*Tuple)
	if !ok || tup.Len() != 2 || !Identical(tup.At(1).Type(), universeError) {
		return false
	}
	switch res.Len() {
	case 2:
		return Identical(res.At(1).Type(), universeError)
	case 1:
		_, ok := res.At(0).Type().Underlying().(*Result)
		return ok
	default:
		return false
	}
}

func (check *Checker) forceReturnWant(exprTyp Type) string {
	if Identical(exprTyp, universeError) {
		if check.sig != nil {
			res := check.sig.Results()
			if res != nil && res.Len() == 1 && !Identical(res.At(0).Type(), universeError) {
				return TypeString(NewResult(res.At(0).Type()), nil)
			}
		}
		return "error"
	}
	if tup, ok := exprTyp.(*Tuple); ok && tup.Len() >= 1 {
		return TypeString(NewResult(tup.At(0).Type()), nil)
	}
	return "error"
}

func (check *Checker) resultSelector(x *operand, e *ast.SelectorExpr) bool {
	res, ok := x.typ().Underlying().(*Result)
	if !ok {
		return false
	}
	sel := e.Sel.Name
	var ftyp Type
	var index int
	switch sel {
	case "value":
		ftyp = res.elem
		index = 0
	case "err":
		ftyp = universeError
		index = 1
	default:
		return false
	}
	v := newVar(FieldVar, e.Sel.Pos(), check.pkg, sel, ftyp)
	check.recordSelection(e, FieldVal, x.typ(), v, []int{index}, false)
	x.mode_ = value
	x.typ_ = ftyp
	x.expr = e
	return true
}
