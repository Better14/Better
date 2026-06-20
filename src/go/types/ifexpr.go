// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	. "internal/types/errors"
)

func (check *Checker) ifExpr(x *operand, e *ast.IfExpr) {
	var cond, then, els operand
	check.expr(nil, &cond, e.Cond)
	if !cond.isValid() {
		x.invalidate()
		return
	}
	if !isBoolean(cond.typ()) {
		check.error(e.Cond, InvalidCond, "non-boolean condition in if expression")
		x.invalidate()
		return
	}
	check.expr(nil, &then, e.Then)
	check.expr(nil, &els, e.ElseBody)
	if !then.isValid() || !els.isValid() {
		x.invalidate()
		return
	}
	t := check.branchCommonType(e, &then, &els)
	if !isValid(t) {
		x.invalidate()
		return
	}
	x.mode_ = value
	x.typ_ = t
	x.expr = e
}

func (check *Checker) branchCommonType(e ast.Expr, a, b *operand) Type {
	if !a.isValid() || !b.isValid() {
		return Typ[Invalid]
	}
	at, bt := a.typ(), b.typ()
	if Identical(at, bt) {
		return at
	}
	if ok, _ := b.assignableTo(check, at, nil); ok {
		return at
	}
	if ok, _ := a.assignableTo(check, bt, nil); ok {
		return bt
	}
	if isUntyped(at) && !isUntyped(bt) {
		if _, _, code := check.implicitTypeAndValue(a, bt); code == 0 {
			return bt
		}
	}
	if isUntyped(bt) && !isUntyped(at) {
		if _, _, code := check.implicitTypeAndValue(b, at); code == 0 {
			return at
		}
	}
	check.errorf(e, MismatchedTypes, "mismatched types %s and %s", at, bt)
	return Typ[Invalid]
}

func (check *Checker) mergeBranchTypes(e ast.Expr, arms []*operand) Type {
	t := arms[0].typ()
	for i := 1; i < len(arms); i++ {
		var prev operand
		prev.mode_ = value
		prev.typ_ = t
		t = check.branchCommonType(e, &prev, arms[i])
		if !isValid(t) {
			return Typ[Invalid]
		}
	}
	return t
}
