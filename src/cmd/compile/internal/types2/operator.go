// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
	"strings"
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

func (check *Checker) pkgForRecv(typ Type) *Package {
	switch t := Unalias(typ).(type) {
	case *Named:
		if t.obj != nil {
			return t.obj.pkg
		}
	case *Pointer:
		return check.pkgForRecv(t.base)
	}
	return nil
}

// operatorFuncsForRecv returns [] or []= overloads from the current package or
// from the package that defines the receiver type.
func (check *Checker) operatorFuncsForRecv(name string, recv Type) []*Func {
	if cands := check.operatorFuncs(name); len(cands) > 0 {
		return cands
	}
	if pkg := check.pkgForRecv(recv); pkg != nil {
		if cands := operatorFuncsInPackage(pkg, name); len(cands) > 0 {
			return cands
		}
	}
	return nil
}

func operatorFuncsInPackage(pkg *Package, name string) []*Func {
	if pkg.overloadFuncs != nil {
		if cands := pkg.overloadFuncs[name]; len(cands) > 0 {
			return cands
		}
	}
	var cands []*Func
	for _, n := range pkg.scope.Names() {
		if n != name && !strings.HasPrefix(n, name+"·") {
			continue
		}
		if obj := pkg.scope.Lookup(n); obj != nil {
			if fn, ok := obj.(*Func); ok {
				cands = append(cands, fn)
			}
		}
	}
	return cands
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

// selectIndexOperator picks a package-level [] or []= overload for an index expression.
// A single generic candidate is accepted without assignability pre-check (inference
// happens in callOperator). Multiple candidates use silent overload resolution.
func (check *Checker) selectIndexOperator(cands []*Func, call *syntax.CallExpr, args []*operand) *Func {
	if len(cands) == 1 {
		return cands[0]
	}
	return check.selectOverloadSilent(call, cands, args)
}

func (check *Checker) callOperator(x *operand, pos syntax.Pos, fn *Func, call *syntax.CallExpr, argExprs []syntax.Expr, args []*operand) {
	sig := fn.typ.(*Signature)
	if call == nil {
		call = &syntax.CallExpr{Fun: syntax.NewName(pos, fn.name), ArgList: argExprs}
	}
	if args == nil || sig.TypeParams().Len() > 0 {
		args, _ = check.genericExprList(argExprs)
	}
	sig = check.arguments(call, sig, nil, nil, args, nil)
	if sig == nil {
		x.invalidate()
		return
	}
	if name, ok := call.Fun.(*syntax.Name); ok {
		check.recordUse(name, fn)
		if sig.TypeParams().Len() == 0 {
			check.recordTypeAndValue(call.Fun, value, sig, nil)
		}
	}
	switch {
	case sig.results == nil || sig.results.Len() == 0:
		x.mode_ = novalue
	case sig.results.Len() == 1:
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
	check.callOperator(x, e.Pos(), fn, nil, []syntax.Expr{lhs, rhs}, []*operand{&l, &r})
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
	check.callOperator(x, e.Pos(), fn, nil, []syntax.Expr{e.X}, []*operand{x})
	if x.isValid() {
		x.expr = e
	}
	return true
}

// tryIndexOperatorOverload handles a[i] when the type of a defines func [](a, i...) U.
func (check *Checker) tryIndexOperatorOverload(x *operand, e *syntax.IndexExpr) bool {
	var l operand
	check.expr(nil, &l, e.X)
	if !l.isValid() || supportsBuiltinIndex(l.typ()) {
		return false
	}
	index := check.singleIndex(e)
	if index == nil {
		return false
	}
	var i operand
	check.expr(nil, &i, index)
	if !i.isValid() {
		return false
	}
	cands := check.operatorFuncsForRecv("[]", l.typ())
	if len(cands) == 0 {
		return false
	}
	call := &syntax.CallExpr{
		Fun:     syntax.NewName(e.Pos(), "[]"),
		ArgList: []syntax.Expr{e.X, index},
	}
	fn := check.selectIndexOperator(cands, call, []*operand{&l, &i})
	if fn == nil {
		return false
	}
	check.callOperator(x, e.Pos(), fn, call, call.ArgList, nil)
	if !x.isValid() {
		return false
	}
	check.recordIndexOperatorCall(e, call)
	x.expr = call
	return true
}

// tryIndexAssignOperatorOverload handles a[i] = v when a's type defines func []=(a, i..., v).
func (check *Checker) tryIndexAssignOperatorOverload(lhs, rhs syntax.Expr, x *operand) bool {
	idx, ok := syntax.Unparen(lhs).(*syntax.IndexExpr)
	if !ok {
		return false
	}
	var l operand
	check.expr(nil, &l, idx.X)
	if !l.isValid() || supportsBuiltinIndex(l.typ()) {
		return false
	}
	index := check.singleIndex(idx)
	if index == nil {
		return false
	}
	var i operand
	check.expr(nil, &i, index)
	if !i.isValid() {
		return false
	}
	var v operand
	if x != nil {
		v = *x
	} else {
		check.expr(nil, &v, rhs)
	}
	if !v.isValid() {
		return false
	}
	cands := check.operatorFuncsForRecv("[]=", l.typ())
	if len(cands) == 0 {
		return false
	}
	call := &syntax.CallExpr{
		Fun:     syntax.NewName(lhs.Pos(), "[]="),
		ArgList: []syntax.Expr{idx.X, index, rhs},
	}
	fn := check.selectIndexOperator(cands, call, []*operand{&l, &i, &v})
	if fn == nil {
		return false
	}
	if x == nil {
		x = new(operand)
	}
	check.callOperator(x, lhs.Pos(), fn, call, call.ArgList, nil)
	if !x.isValid() {
		return false
	}
	check.recordIndexAssignCall(idx, call)
	x.mode_ = novalue
	x.expr = lhs
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
	check.callOperator(&res, s.Pos(), fn, nil, []syntax.Expr{s.Lhs}, []*operand{&arg})
	if !res.isValid() {
		return false
	}
	check.assignVar(s.Lhs, nil, &res, "assignment")
	return true
}
