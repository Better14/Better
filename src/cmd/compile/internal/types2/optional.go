// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

// An Optional represents a nullable type T? (value or nil).
type Optional struct {
	elem Type
}

// NewOptional returns a new nullable type for the given element type.
func NewOptional(elem Type) Type {
	return &Optional{elem: elem}
}

// Elem returns the element type of o.
func (o *Optional) Elem() Type { return o.elem }

func (o *Optional) Underlying() Type { return o }
func (o *Optional) String() string   { return TypeString(o, nil) }

// AsOptional reports whether t is a *Optional and returns it.
func AsOptional(t Type) (*Optional, bool) {
	o, _ := t.(*Optional)
	return o, o != nil
}

// isNullish reports whether t may be compared to nil and used with ?. and ??.
func isNullish(t Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *Optional, *Pointer, *Slice, *Map, *Chan, *Signature, *Interface:
		return true
	default:
		return false
	}
}

// optionalElem returns the element type for null-conditional access on t.
func optionalElem(t Type) Type {
	switch u := t.Underlying().(type) {
	case *Optional:
		return u.elem
	case *Pointer:
		return u.base
	case *Slice:
		return u.elem
	case *Map:
		return u.key
	default:
		return nil
	}
}

// ptrForNullish returns the type used for field lookup on a nullish t.
func ptrForNullish(t Type) Type {
	if _, ok := t.Underlying().(*Pointer); ok {
		return t
	}
	elem := optionalElem(t)
	if elem == nil {
		return Typ[Invalid]
	}
	return NewPointer(elem)
}

// optionalResultType is the result type of expr?.field when expr has type t.
func optionalResultType(t Type) Type {
	if !isValid(t) {
		return Typ[Invalid]
	}
	if _, ok := t.Underlying().(*Optional); ok {
		return t
	}
	elem := t
	if p, ok := t.Underlying().(*Pointer); ok {
		elem = p.base
	}
	return NewOptional(elem)
}

func (check *Checker) nullCondSelector(x *operand, e *syntax.SelectorExpr, nc *syntax.NullCondExpr) {
	var base operand
	check.expr(nil, &base, nc.X)
	if !base.isValid() {
		x.invalidate()
		return
	}
	if !isNullish(base.typ()) {
		check.errorf(nc, InvalidSyntaxTree, "invalid operation: ?. requires nullable operand, got %s", base.typ())
		x.invalidate()
		return
	}
	recvTyp := ptrForNullish(base.typ())
	sel := e.Sel.Value
	obj, index, indirect := lookupFieldOrMethod(recvTyp, false, check.pkg, sel, false)
	if obj == nil {
		check.errorf(e.Sel, MissingFieldOrMethod, "%s undefined", e.Sel.Value)
		x.invalidate()
		return
	}
	v, ok := obj.(*Var)
	if !ok {
		check.errorf(e.Sel, MissingFieldOrMethod, "%s.%s is not a field", base.expr, sel)
		x.invalidate()
		return
	}
	check.recordSelection(e, FieldVal, recvTyp, v, index, indirect)
	x.mode_ = value
	x.typ_ = optionalResultType(v.typ)
	x.expr = e
}

func (check *Checker) nullCondIndex(x *operand, e *syntax.IndexExpr, nc *syntax.NullCondExpr) {
	var base operand
	check.expr(nil, &base, nc.X)
	if !base.isValid() {
		x.invalidate()
		return
	}
	if !isNullish(base.typ()) {
		check.errorf(nc, InvalidSyntaxTree, "invalid operation: ?. requires nullable operand, got %s", base.typ())
		x.invalidate()
		return
	}
	*x = base
	x.typ_ = ptrForNullish(base.typ())
	if check.indexExpr(x, e) {
		check.funcInst(nil, e.Pos(), x, e, true)
	}
	if !x.isValid() {
		return
	}
	x.typ_ = optionalResultType(x.typ())
	if x.mode() == variable {
		x.mode_ = value
	}
	x.expr = e
}

func (check *Checker) nullCoalesce(x *operand, e syntax.Expr, lhs, rhs syntax.Expr) {
	var y operand
	check.expr(nil, x, lhs)
	if !x.isValid() {
		return
	}
	if res, ok := x.typ().Underlying().(*Result); ok {
		check.expr(nil, &y, rhs)
		if !y.isValid() {
			x.invalidate()
			x.expr = y.expr
			return
		}
		if ok, _ := y.assignableTo(check, res.elem, nil); ok {
			x.typ_ = res.elem
		} else if vres, ok2 := y.typ().Underlying().(*Result); ok2 && Identical(res.elem, vres.elem) {
			x.typ_ = res.elem
		} else {
			check.errorf(e, MismatchedTypes, "invalid operation: ?? (cannot use %s as %s)", y.typ(), res.elem)
			x.invalidate()
		}
		if x.isValid() {
			x.mode_ = value
			x.expr = e
		}
		return
	}
	if !isNullish(x.typ()) {
		check.errorf(e, InvalidSyntaxTree, "invalid operation: ?? requires nullable or result left operand, got %s", x.typ())
		x.invalidate()
		return
	}
	check.expr(nil, &y, rhs)
	if !y.isValid() {
		x.invalidate()
		x.expr = y.expr
		return
	}

	if o, ok := x.typ().Underlying().(*Optional); ok {
		if o2, ok2 := y.typ().Underlying().(*Optional); ok2 {
			if Identical(o.elem, o2.elem) {
				x.typ_ = x.typ()
			} else {
				check.errorf(e, MismatchedTypes, "invalid operation: ?? (mismatched types %s and %s)", x.typ(), y.typ())
				x.invalidate()
			}
		} else if ok, _ := y.assignableTo(check, o.elem, nil); ok {
			x.typ_ = o.elem
		} else {
			check.errorf(e, MismatchedTypes, "invalid operation: ?? (cannot use %s as %s)", y.typ(), o.elem)
			x.invalidate()
		}
	} else if p, ok := x.typ().Underlying().(*Pointer); ok {
		elem := p.base
		if isUntyped(y.typ()) {
			if _, _, code := check.implicitTypeAndValue(&y, elem); code == 0 {
				x.typ_ = elem
			} else {
				check.errorf(e, MismatchedTypes, "invalid operation: ?? (cannot use %s as %s)", y.typ(), elem)
				x.invalidate()
			}
		} else if ok, _ := y.assignableTo(check, elem, nil); ok {
			x.typ_ = elem
		} else if o2, ok2 := y.typ().Underlying().(*Optional); ok2 && Identical(elem, o2.elem) {
			x.typ_ = elem
		} else {
			check.errorf(e, MismatchedTypes, "invalid operation: ?? (cannot use %s as %s)", y.typ(), elem)
			x.invalidate()
		}
	} else {
		if Identical(x.typ(), y.typ()) {
			x.typ_ = x.typ()
		} else if ok, _ := y.assignableTo(check, x.typ(), nil); ok {
			x.typ_ = y.typ()
		} else {
			check.errorf(e, MismatchedTypes, "invalid operation: ?? (mismatched types %s and %s)", x.typ(), y.typ())
			x.invalidate()
		}
	}
	if x.isValid() {
		x.mode_ = value
		x.expr = e
	}
}

// nullableElem returns the element type if t is T?, or nil.
func nullableElem(t Type) Type {
	if o, ok := t.Underlying().(*Optional); ok {
		return o.elem
	}
	return nil
}

// withNullableNarrow runs f with extra nullable variable narrowing in effect.
func (check *Checker) withNullableNarrow(narrow map[*Var]Type, f func()) {
	if len(narrow) == 0 {
		f()
		return
	}
	old := check.nullableNarrow
	merged := make(map[*Var]Type, len(old)+len(narrow))
	for v, t := range old {
		merged[v] = t
	}
	for v, t := range narrow {
		merged[v] = t
	}
	check.nullableNarrow = merged
	f()
	check.nullableNarrow = old
}

// parseNullableGuard recognizes v != nil, nil != v, and v == nil for nullable v.
func (check *Checker) parseNullableGuard(cond syntax.Expr) (v *Var, nonNil bool, ok bool) {
	op, ok := syntax.Unparen(cond).(*syntax.Operation)
	if !ok {
		return nil, false, false
	}
	switch op.Op {
	case syntax.Neq:
		return check.nullableGuardIdent(op.X, op.Y, true)
	case syntax.Eql:
		return check.nullableGuardIdent(op.X, op.Y, false)
	default:
		return nil, false, false
	}
}

func (check *Checker) nullableGuardIdent(x, y syntax.Expr, nonNil bool) (*Var, bool, bool) {
	if check.isNil(x) {
		x, y = y, x
	} else if !check.isNil(y) {
		return nil, false, false
	}
	name, ok := syntax.Unparen(x).(*syntax.Name)
	if !ok {
		return nil, false, false
	}
	obj := check.lookup(name.Value)
	v, ok := obj.(*Var)
	if !ok || nullableElem(v.typ) == nil {
		return nil, false, false
	}
	return v, nonNil, true
}
