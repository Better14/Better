// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
	. "internal/types/errors"
)

// An Optional represents a nilable type T? (value or nil).
type Optional struct {
	elem Type
}

// NewOptional returns a new nilable type for the given element type.
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

// OptionalType returns the struct type used to lower T? to Option[T].
func OptionalType(pkg *Package, opt *Optional) *Struct {
	pos := nopos
	return NewStruct([]*Var{
		newVar(FieldVar, pos, pkg, "hasValue", Typ[Bool]),
		newVar(FieldVar, pos, pkg, "value", opt.elem),
	}, nil)
}

// optionalStructElem reports whether t is a lowered T? struct and returns its value type.
func optionalStructElem(t Type) (Type, bool) {
	if t == nil {
		return nil, false
	}
	s, ok := t.Underlying().(*Struct)
	if !ok || s.NumFields() != 2 {
		return nil, false
	}
	f0, f1 := s.Field(0), s.Field(1)
	if f0.Name() != "hasValue" || !Identical(f0.Type(), Typ[Bool]) || f1.Name() != "value" {
		return nil, false
	}
	return f1.Type(), true
}

// isNullish reports whether t may be compared to nil and used with ?. and ??.
func isNullish(t Type) bool {
	if t == nil {
		return false
	}
	if _, ok := optionalStructElem(t); ok {
		return true
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
	if elem, ok := optionalStructElem(t); ok {
		return elem
	}
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
	if o, ok := t.Underlying().(*Optional); ok {
		return o.elem
	}
	if elem, ok := optionalStructElem(t); ok {
		return elem
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
	if _, ok := optionalStructElem(t); ok {
		return t
	}
	return NewOptional(t)
}

func (check *Checker) nullCondSelector(x *operand, e *ast.SelectorExpr, nc *ast.NullCondExpr) {
	var base operand
	check.expr(nil, &base, nc.X)
	if !base.isValid() {
		x.invalidate()
		return
	}
	if !isNullish(base.typ()) {
		check.errorf(nc, InvalidSyntaxTree, "invalid operation: ?. requires nilable operand, got %s", base.typ())
		x.invalidate()
		return
	}
	recvTyp := ptrForNullish(base.typ())
	sel := e.Sel.Name
	obj, index, indirect := lookupFieldOrMethod(recvTyp, false, check.pkg, sel, false)
	if obj == nil {
		check.errorf(e.Sel, MissingFieldOrMethod, "%s undefined", e.Sel.Name)
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

func (check *Checker) nullCondIndex(x *operand, e *ast.IndexExpr, nc *ast.NullCondExpr) {
	var base operand
	check.expr(nil, &base, nc.X)
	if !base.isValid() {
		x.invalidate()
		return
	}
	if !isNullish(base.typ()) {
		check.errorf(nc, InvalidSyntaxTree, "invalid operation: ?. requires nilable operand, got %s", base.typ())
		x.invalidate()
		return
	}
	*x = base
	x.typ_ = ptrForNullish(base.typ())
	if ix := unpackIndexedExpr(e); check.indexExpr(x, ix) {
		check.funcInst(nil, e.Pos(), x, ix, true)
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

func (check *Checker) nullCoalesce(x *operand, e ast.Expr, lhs, rhs ast.Expr) {
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
		check.errorf(e, InvalidSyntaxTree, "invalid operation: ?? requires nilable or result left operand, got %s", x.typ())
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
	} else if elem, ok := optionalStructElem(x.typ()); ok {
		if o2, ok2 := y.typ().Underlying().(*Optional); ok2 {
			if Identical(elem, o2.elem) {
				x.typ_ = x.typ()
			} else {
				check.errorf(e, MismatchedTypes, "invalid operation: ?? (mismatched types %s and %s)", x.typ(), y.typ())
				x.invalidate()
			}
		} else if ok, _ := y.assignableTo(check, elem, nil); ok {
			x.typ_ = elem
		} else {
			check.errorf(e, MismatchedTypes, "invalid operation: ?? (cannot use %s as %s)", y.typ(), elem)
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

// nilableElem returns the element type if t is T? or *T?, or nil.
func nilableElem(t Type) Type {
	if elem, ok := nilablePointerElem(t); ok {
		return elem
	}
	if o, ok := t.Underlying().(*Optional); ok {
		return o.elem
	}
	if elem, ok := optionalStructElem(t); ok {
		return elem
	}
	return nil
}

// withNilableNarrow runs f with extra nilable variable narrowing in effect.
func (check *Checker) withNilableNarrow(narrow map[*Var]Type, f func()) {
	if len(narrow) == 0 {
		f()
		return
	}
	old := check.nilableNarrow
	merged := make(map[*Var]Type, len(old)+len(narrow))
	for v, t := range old {
		merged[v] = t
	}
	for v, t := range narrow {
		merged[v] = t
	}
	check.nilableNarrow = merged
	f()
	check.nilableNarrow = old
}

// parseNilableGuard recognizes v != nil, nil != v, and v == nil for nilable v.
func (check *Checker) parseNilableGuard(cond ast.Expr) (v *Var, nonNil bool, ok bool) {
	op, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return nil, false, false
	}
	switch op.Op {
	case token.NEQ:
		return check.nilableGuardIdent(op.X, op.Y, true)
	case token.EQL:
		return check.nilableGuardIdent(op.X, op.Y, false)
	default:
		return nil, false, false
	}
}

func (check *Checker) nilableGuardIdent(x, y ast.Expr, nonNil bool) (*Var, bool, bool) {
	if check.isNil(x) {
		x, y = y, x
	} else if !check.isNil(y) {
		return nil, false, false
	}
	name, ok := ast.Unparen(x).(*ast.Ident)
	if !ok {
		return nil, false, false
	}
	obj := check.lookup(name.Name)
	v, ok := obj.(*Var)
	if !ok || nilableElem(v.typ) == nil {
		return nil, false, false
	}
	return v, nonNil, true
}
