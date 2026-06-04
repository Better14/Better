// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

var comparisonOpPairs = [][2]string{
	{"==", "!="},
	{"<", ">"},
	{"<=", ">="},
}

func (check *Checker) validateOperatorPairs() {
	for _, pair := range comparisonOpPairs {
		a := len(check.operatorFuncs(pair[0])) > 0
		b := len(check.operatorFuncs(pair[1])) > 0
		if a == b {
			continue
		}
		missing := pair[1]
		if !a {
			missing = pair[0]
		}
		for _, name := range pair {
			for _, fn := range check.operatorFuncs(name) {
				check.errorf(fn.pos, InvalidSyntaxTree, "operator %s requires paired operator %s", fn.name, missing)
			}
		}
	}
}

func (check *Checker) operatorFuncs(name string) []*Func {
	if check.overloadFuncs == nil {
		return nil
	}
	return check.overloadFuncs[name]
}

func (check *Checker) selectOperatorFunc(name string, nargs int, args []*operand) *Func {
	cands := check.operatorFuncs(name)
	if len(cands) == 0 {
		return nil
	}
	var matches []*Func
	for _, fn := range cands {
		if fn == nil || fn.typ == nil {
			continue
		}
		sig := fn.typ.(*Signature)
		if sig.params.Len() != nargs {
			continue
		}
		ok := true
		for i := 0; i < nargs; i++ {
			if okAssign, _ := args[i].assignableTo(check, sig.params.vars[i].typ, nil); !okAssign {
				ok = false
				break
			}
		}
		if ok {
			matches = append(matches, fn)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return nil
}

func (check *Checker) callOperator(x *operand, pos syntax.Pos, fn *Func, argExprs []syntax.Expr, args []*operand) {
	sig := fn.typ.(*Signature)
	call := &syntax.CallExpr{Fun: syntax.NewName(pos, fn.name), ArgList: argExprs}
	sig = check.arguments(call, sig, nil, nil, args, nil)
	if sig == nil || sig.results == nil {
		x.invalidate()
		return
	}
	switch sig.results.Len() {
	case 0:
		x.mode_ = novalue
	case 1:
		x.mode_ = value
		x.typ_ = sig.results.vars[0].typ
	default:
		x.mode_ = value
		x.typ_ = sig.results
	}
}

func (check *Checker) tryBinaryOperatorOverload(x *operand, e syntax.Expr, lhs, rhs syntax.Expr, op syntax.Operator) bool {
	if op == syntax.NullCoalesce || op == syntax.AndAnd || op == syntax.OrOr {
		return false
	}
	name := op.String()
	var l operand
	check.expr(nil, &l, lhs)
	if !l.isValid() {
		return false
	}
	var r operand
	check.expr(nil, &r, rhs)
	if !r.isValid() {
		return false
	}
	fn := check.selectOperatorFunc(name, 2, []*operand{&l, &r})
	if fn == nil {
		return false
	}
	check.callOperator(x, e.Pos(), fn, []syntax.Expr{lhs, rhs}, []*operand{&l, &r})
	if x.isValid() {
		x.expr = e
	}
	return true
}

func (check *Checker) tryUnaryOperatorOverload(x *operand, e *syntax.Operation) bool {
	op := e.Op
	if op == syntax.And || op == syntax.Recv || op == syntax.Mul || op == syntax.Tilde {
		return false
	}
	fn := check.selectOperatorFunc(op.String(), 1, []*operand{x})
	if fn == nil {
		return false
	}
	check.callOperator(x, e.Pos(), fn, []syntax.Expr{e.X}, []*operand{x})
	if x.isValid() {
		x.expr = e
	}
	return true
}

func (check *Checker) tryIncDecOperatorOverload(s *syntax.AssignStmt, op syntax.Operator) bool {
	name := "++"
	if op == syntax.Sub {
		name = "--"
	}
	var arg operand
	check.expr(nil, &arg, s.Lhs)
	if !arg.isValid() {
		return false
	}
	fn := check.selectOperatorFunc(name, 1, []*operand{&arg})
	if fn == nil {
		return false
	}
	var res operand
	check.callOperator(&res, s.Pos(), fn, []syntax.Expr{s.Lhs}, []*operand{&arg})
	if !res.isValid() {
		return false
	}
	check.assignVar(s.Lhs, nil, &res, "assignment")
	return true
}
