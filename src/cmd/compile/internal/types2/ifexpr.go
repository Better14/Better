// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

func (check *Checker) ifExpr(x *operand, e *syntax.IfExpr) {
	var cond, then, els operand
	check.expr(nil, &cond, e.Cond)
	if !cond.isValid() {
		x.invalidate()
		return
	}
	if !isBoolean(cond.typ()) {
		check.errorf(e.Cond, InvalidCond, "non-boolean condition in if expression")
		x.invalidate()
		return
	}
	check.expr(nil, &then, e.Then)
	check.expr(nil, &els, e.Else)
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

// branchCommonType returns the result type for if/switch expression arms.
func (check *Checker) branchCommonType(e syntax.Expr, a, b *operand) Type {
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
