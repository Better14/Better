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
	if t == nil {
		return nil, false
	}
	t = Unalias(t)
	if o, ok := t.(*Optional); ok {
		return o, true
	}
	if o, ok := t.Underlying().(*Optional); ok {
		return o, true
	}
	return nil, false
}

// OptionalDestElem returns the element type of an optional destination T?
// (either *Optional or a lowered optional struct).
func OptionalDestElem(t Type) (elem Type, ok bool) {
	if t == nil {
		return nil, false
	}
	if o, ok := AsOptional(t); ok {
		return o.Elem(), true
	}
	return optionalStructElem(t)
}

// OptionalSrcElem returns the element type of an optional source T?
// (either *Optional or a lowered optional struct).
func OptionalSrcElem(t Type) (elem Type, ok bool) {
	return OptionalDestElem(t)
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

// withNilableNarrow runs f with extra nilable narrowing in effect.
func (check *Checker) withNilableNarrow(narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, f func()) {
	if len(narrowVars) == 0 && len(narrowSels) == 0 {
		f()
		return
	}
	oldV := check.nilableNarrow
	oldS := check.nilableNarrowSel
	if len(narrowVars) > 0 {
		merged := make(map[*Var]Type, len(oldV)+len(narrowVars))
		for v, t := range oldV {
			merged[v] = t
		}
		for v, t := range narrowVars {
			merged[v] = t
		}
		check.nilableNarrow = merged
	}
	if len(narrowSels) > 0 {
		merged := make(map[nilableSelKey]Type, len(oldS)+len(narrowSels))
		for k, t := range oldS {
			merged[k] = t
		}
		for k, t := range narrowSels {
			merged[k] = t
		}
		check.nilableNarrowSel = merged
	}
	f()
	check.nilableNarrow = oldV
	check.nilableNarrowSel = oldS
}

// parseNilableGuard recognizes x != nil, nil != x, and x == nil for nilable x.
// It also handles `x != nil && …` and `x == nil || …`.
func (check *Checker) parseNilableGuard(cond ast.Expr) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, nonNil bool, ok bool) {
	if op, ok := ast.Unparen(cond).(*ast.BinaryExpr); ok {
		switch op.Op {
		case token.LAND:
			if nv, ns, nn, gok := check.parseNilableGuard(op.X); gok && nn {
				return nv, ns, true, true
			}
		case token.LOR:
			if nv, ns, nn, gok := check.parseNilableGuard(op.X); gok && !nn {
				return nv, ns, true, true
			}
		}
	}
	op, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return nil, nil, false, false
	}
	switch op.Op {
	case token.NEQ:
		return check.nilableGuardExpr(op.X, op.Y, true)
	case token.EQL:
		return check.nilableGuardExpr(op.X, op.Y, false)
	default:
		return nil, nil, false, false
	}
}

func (check *Checker) nilableGuardExpr(x, y ast.Expr, nonNil bool) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, nonNilOut bool, ok bool) {
	if check.isNil(x) {
		x, y = y, x
	} else if !check.isNil(y) {
		return nil, nil, false, false
	}
	x = ast.Unparen(x)
	switch e := x.(type) {
	case *ast.Ident:
		obj := check.lookup(e.Name)
		if v, ok := obj.(*Var); ok {
			if elem := nilableElem(v.typ); elem != nil {
				return map[*Var]Type{v: elem}, nil, nonNil, true
			}
		}
	case *ast.SelectorExpr:
		var xop operand
		check.expr(nil, &xop, e)
		if elem := nilableElem(xop.typ()); elem != nil {
			if key, ok := check.nilableSelKeyFrom(e); ok {
				return nil, map[nilableSelKey]Type{key: elem}, nonNil, true
			}
		}
	}
	return nil, nil, false, false
}

func (check *Checker) nilableSelKeyFrom(e *ast.SelectorExpr) (nilableSelKey, bool) {
	name, ok := ast.Unparen(e.X).(*ast.Ident)
	if !ok {
		return nilableSelKey{}, false
	}
	obj := check.lookup(name.Name)
	if obj == nil {
		return nilableSelKey{}, false
	}
	return nilableSelKey{obj: obj, sel: e.Sel.Name}, true
}

// nilableGuardEarlyReturnNarrow reports narrowing for `if x == nil { <terminating> }` with no else.
func (check *Checker) nilableGuardEarlyReturnNarrow(s *ast.IfStmt) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, ok bool) {
	if s.Init != nil || s.Else != nil {
		return nil, nil, false
	}
	narrowVars, narrowSels, nonNil, guardOk := check.parseNilableGuard(s.Cond)
	if !guardOk || nonNil || (len(narrowVars) == 0 && len(narrowSels) == 0) {
		return nil, nil, false
	}
	if !check.isTerminating(s.Body, "") {
		return nil, nil, false
	}
	return narrowVars, narrowSels, true
}

// nilableGuardDefaultAssignNarrow reports narrowing for `if x == nil { x = <non-nil> }` with no else.
// Statements after the if can treat x as non-nilable (strict *T or T).
func (check *Checker) nilableGuardDefaultAssignNarrow(s *ast.IfStmt) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, ok bool) {
	if s.Init != nil || s.Else != nil {
		return nil, nil, false
	}
	narrowVars, narrowSels, nonNil, guardOk := check.parseNilableGuard(s.Cond)
	if !guardOk || nonNil || (len(narrowVars) == 0 && len(narrowSels) == 0) {
		return nil, nil, false
	}
	if !check.nilDefaultAssignInBlock(s.Body, narrowVars, narrowSels) {
		return nil, nil, false
	}
	return narrowVars, narrowSels, true
}

// isSkipping reports whether s always exits the current control path
// (return, continue, break, goto, fallthrough, or panic).
func (check *Checker) isSkipping(s ast.Stmt, label string) bool {
	switch s := s.(type) {
	default:
		return false
	case *ast.ExprStmt:
		if call, ok := ast.Unparen(s.X).(*ast.CallExpr); ok && check.isPanic[call] {
			return true
		}
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		switch s.Tok {
		case token.CONTINUE, token.BREAK, token.GOTO, token.FALLTHROUGH:
			return true
		}
	case *ast.LabeledStmt:
		return check.isSkipping(s.Stmt, s.Label.Name)
	case *ast.BlockStmt:
		return check.isSkippingList(s.List, label)
	case *ast.IfStmt:
		if s.Else != nil {
			return false
		}
		return check.isSkipping(s.Body, label)
	}
	return false
}

func (check *Checker) isSkippingList(list []ast.Stmt, label string) bool {
	for i := len(list) - 1; i >= 0; i-- {
		if _, ok := list[i].(*ast.EmptyStmt); !ok {
			return check.isSkipping(list[i], label)
		}
	}
	return false
}

// nilableGuardEarlyContinueNarrow reports narrowing after `if x == nil { <skip> }` with no else.
func (check *Checker) nilableGuardEarlyContinueNarrow(s *ast.IfStmt) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, ok bool) {
	if s.Init != nil || s.Else != nil {
		return nil, nil, false
	}
	narrowVars, narrowSels, nonNil, guardOk := check.parseNilableGuard(s.Cond)
	if !guardOk || nonNil || (len(narrowVars) == 0 && len(narrowSels) == 0) {
		return nil, nil, false
	}
	if !check.isSkipping(s.Body, "") {
		return nil, nil, false
	}
	return narrowVars, narrowSels, true
}

// nilableGuardOrNilFirstNarrow reports narrowing after `if x == nil || … { <skip> }` with no else.
func (check *Checker) nilableGuardOrNilFirstNarrow(s *ast.IfStmt) (narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type, ok bool) {
	if s.Init != nil || s.Else != nil {
		return nil, nil, false
	}
	op, ok := ast.Unparen(s.Cond).(*ast.BinaryExpr)
	if !ok || op.Op != token.LOR {
		return nil, nil, false
	}
	narrowVars, narrowSels, nonNil, guardOk := check.parseNilableGuard(op.X)
	if !guardOk || nonNil || (len(narrowVars) == 0 && len(narrowSels) == 0) {
		return nil, nil, false
	}
	if !check.isSkipping(s.Body, "") && !check.isTerminating(s.Body, "") {
		return nil, nil, false
	}
	return narrowVars, narrowSels, true
}

func (check *Checker) nilDefaultAssignInBlock(body *ast.BlockStmt, narrowVars map[*Var]Type, narrowSels map[nilableSelKey]Type) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	assigned := make(map[*Var]bool, len(narrowVars))
	assignedSel := make(map[nilableSelKey]bool, len(narrowSels))
	for _, st := range body.List {
		check.collectNilDefaultAssigns(st, assigned, assignedSel)
	}
	for v := range narrowVars {
		if !assigned[v] {
			return false
		}
	}
	for k := range narrowSels {
		if !assignedSel[k] {
			return false
		}
	}
	return true
}

func (check *Checker) collectNilDefaultAssigns(stmt ast.Stmt, assigned map[*Var]bool, assignedSel map[nilableSelKey]bool) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok != token.ASSIGN && s.Tok != token.DEFINE {
			return
		}
		for i, lhs := range s.Lhs {
			rhs := rhsAt(s.Rhs, i)
			if rhs == nil || check.isNil(rhs) || !check.isNonNilPointerLike(rhs) {
				continue
			}
			switch e := ast.Unparen(lhs).(type) {
			case *ast.Ident:
				obj := check.lookup(e.Name)
				if v, ok := obj.(*Var); ok {
					assigned[v] = true
				}
			case *ast.SelectorExpr:
				if key, ok := check.nilableSelKeyFrom(e); ok {
					assignedSel[key] = true
				}
			}
		}
	case *ast.BlockStmt:
		for _, st := range s.List {
			check.collectNilDefaultAssigns(st, assigned, assignedSel)
		}
	}
}

func rhsAt(rhs []ast.Expr, i int) ast.Expr {
	if len(rhs) == 1 {
		return rhs[0]
	}
	if i < len(rhs) {
		return rhs[i]
	}
	return nil
}

func (check *Checker) isNonNilPointerLike(e ast.Expr) bool {
	e = ast.Unparen(e)
	if check.isNil(e) {
		return false
	}
	switch e.(type) {
	case *ast.UnaryExpr, *ast.CompositeLit, *ast.CallExpr:
		return true
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return true
	default:
		return false
	}
}
