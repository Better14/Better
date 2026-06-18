// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
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
				check.errorf(atPos(fn.pos), InvalidSyntaxTree, "operator %s requires paired operator %s", fn.name, missing)
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

func (check *Checker) indexOperatorMethods(name string, recv Type) []*Func {
	recvName := recvBaseNameFromType(recv)
	if recvName == "" {
		return nil
	}
	key := methodKey{recvName: recvName, name: name}
	if cands := check.overloadMeths[key]; len(cands) > 0 {
		return cands
	}
	if pkg := check.pkgForRecv(recv); pkg != nil && pkg.overloadMeths != nil {
		return pkg.overloadMeths[key]
	}
	return nil
}

// operatorFuncsForRecv returns [] or []= overloads for recv, preferring receiver
// methods and falling back to package-level operators.
func (check *Checker) operatorFuncsForRecv(name string, recv Type) []*Func {
	if cands := check.indexOperatorMethods(name, recv); len(cands) > 0 {
		return cands
	}
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

func (check *Checker) selectOperatorFunc(name string, nargs int, args []*operand) *Func {
	if fn := check.lookupOverloadByArgTypes(name, args); fn != nil {
		return fn
	}
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

func (check *Checker) indexOperatorCall(pos token.Pos, name string, recvExpr ast.Expr, indices []ast.Expr, value ast.Expr, fn *Func) *ast.CallExpr {
	sig := fn.typ.(*Signature)
	if sig.Recv() != nil {
		args := append([]ast.Expr{}, indices...)
		if value != nil {
			args = append(args, value)
		}
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   recvExpr,
				Sel: astNewIdent(pos, name),
			},
			Args: args,
		}
	}
	args := []ast.Expr{recvExpr}
	args = append(args, indices...)
	if value != nil {
		args = append(args, value)
	}
	return &ast.CallExpr{
		Fun:  astNewIdent(pos, name),
		Args: args,
	}
}

// selectIndexOperator picks a [] or []= overload for an index expression.
// A single generic candidate is accepted without assignability pre-check (inference
// happens in callOperator). Multiple candidates use silent overload resolution.
func (check *Checker) selectIndexOperator(cands []*Func, call *ast.CallExpr, args []*operand) *Func {
	if len(cands) == 1 {
		return cands[0]
	}
	return check.selectOverloadSilent(call, cands, args)
}

func (check *Checker) callOperator(x *operand, pos token.Pos, fn *Func, call *ast.CallExpr, argExprs []ast.Expr, args []*operand) {
	sig := fn.typ.(*Signature)
	if call == nil {
		call = &ast.CallExpr{Fun: astNewIdent(pos, fn.name), Args: argExprs}
	}
	if args == nil || sig.TypeParams().Len() > 0 {
		args, _ = check.genericExprList(argExprs)
	}
	sig = check.arguments(call, sig, nil, nil, args, nil)
	if sig == nil {
		x.invalidate()
		return
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

func (check *Checker) applyBinaryOperatorOverload(x, y *operand, e ast.Expr, lhs, rhs ast.Expr, op token.Token) bool {
	if op == token.NULLCOALESCE || op == token.LAND || op == token.LOR {
		return false
	}
	name := op.String()
	if len(check.operatorFuncs(name)) == 0 {
		return false
	}
	if !x.isValid() || !y.isValid() {
		return false
	}
	fn := check.lookupBinaryOperatorExact(name, x, y)
	if fn == nil {
		fn = check.selectOperatorFunc(name, 2, []*operand{x, y})
	}
	if fn == nil {
		return false
	}
	pos := lhs.Pos()
	if e != nil {
		pos = e.Pos()
	}
	check.callOperator(x, pos, fn, nil, []ast.Expr{lhs, rhs}, []*operand{x, y})
	if x.isValid() {
		if e != nil {
			x.expr = e
		}
	}
	return true
}

func (check *Checker) tryUnaryOperatorOverload(x *operand, e *ast.UnaryExpr) bool {
	op := e.Op
	if op == token.AND || op == token.ARROW || op == token.MUL || op == token.TILDE {
		return false
	}
	name := op.String()
	if len(check.operatorFuncs(name)) == 0 {
		return false
	}
	fn := check.lookupUnaryOperatorExact(name, x)
	if fn == nil {
		fn = check.selectOperatorFunc(name, 1, []*operand{x})
	}
	if fn == nil {
		return false
	}
	check.callOperator(x, e.Pos(), fn, nil, []ast.Expr{e.X}, []*operand{x})
	if x.isValid() {
		x.expr = e
	}
	return true
}

// tryIndexOperatorOverload handles a[i] when the type of a defines func [](a, i...) U.
// base holds the type-checked receiver when indexExpr has already evaluated e.X.
func (check *Checker) tryIndexOperatorOverload(x *operand, e *indexedExpr, base *operand) bool {
	var l operand
	if base != nil && base.isValid() {
		l = *base
	} else {
		check.expr(nil, &l, e.x)
	}
	if !l.isValid() || supportsBuiltinIndex(l.typ()) {
		return false
	}
	indices := check.overloadIndices(e)
	if indices == nil {
		return false
	}
	var indexOps []*operand
	for _, index := range indices {
		var i operand
		check.expr(nil, &i, index)
		if !i.isValid() {
			return false
		}
		indexOps = append(indexOps, &i)
	}
	cands := check.operatorFuncsForRecv("[]", l.typ())
	if len(cands) == 0 {
		return false
	}
	var preload []*operand
	if cands[0].typ.(*Signature).Recv() != nil {
		preload = indexOps
	} else {
		preload = append([]*operand{&l}, indexOps...)
	}
	call := check.indexOperatorCall(e.Pos(), "[]", e.x, indices, nil, cands[0])
	fn := check.selectIndexOperator(cands, call, preload)
	if fn == nil {
		return false
	}
	call = check.indexOperatorCall(e.Pos(), "[]", e.x, indices, nil, fn)
	check.expr(nil, x, call)
	if !x.isValid() {
		return false
	}
	check.recordIndexOperatorCall(e.orig, call)
	x.expr = call
	return true
}

// tryIndexAssignOperatorOverload handles a[i] = v when a's type defines func []=(a, i..., v).
func (check *Checker) tryIndexAssignOperatorOverload(lhs ast.Expr, rhs ast.Expr, x *operand) bool {
	ix := unpackIndexedExpr(ast.Unparen(lhs))
	if ix == nil {
		return false
	}
	var l operand
	check.expr(nil, &l, ix.x)
	if !l.isValid() || supportsBuiltinIndex(l.typ()) {
		return false
	}
	indices := check.overloadIndices(ix)
	if indices == nil {
		return false
	}
	var indexOps []*operand
	for _, index := range indices {
		var i operand
		check.expr(nil, &i, index)
		if !i.isValid() {
			return false
		}
		indexOps = append(indexOps, &i)
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
	var preload []*operand
	if cands[0].typ.(*Signature).Recv() != nil {
		preload = append(indexOps, &v)
	} else {
		preload = append([]*operand{&l}, indexOps...)
		preload = append(preload, &v)
	}
	call := check.indexOperatorCall(lhs.Pos(), "[]=", ix.x, indices, rhs, cands[0])
	fn := check.selectIndexOperator(cands, call, preload)
	if fn == nil {
		return false
	}
	if x == nil {
		x = new(operand)
	}
	call = check.indexOperatorCall(lhs.Pos(), "[]=", ix.x, indices, rhs, fn)
	check.rawExpr(nil, x, call, nil, true)
	check.record(x)
	if !x.isValid() || x.mode() != novalue {
		return false
	}
	check.recordIndexAssignCall(ix.orig, call)
	x.mode_ = novalue
	x.expr = lhs
	return true
}

func (check *Checker) tryIncDecOperatorOverload(s *ast.AssignStmt, op token.Token) bool {
	if len(s.Lhs) != 1 {
		return false
	}
	lhs := s.Lhs[0]
	name := "++"
	if op == token.SUB {
		name = "--"
	}
	var arg operand
	check.expr(nil, &arg, lhs)
	if !arg.isValid() {
		return false
	}
	fn := check.selectOperatorFunc(name, 1, []*operand{&arg})
	if fn == nil {
		return false
	}
	var res operand
	check.callOperator(&res, s.Pos(), fn, nil, []ast.Expr{lhs}, []*operand{&arg})
	if !res.isValid() {
		return false
	}
	check.assignVar(lhs, nil, &res, "assignment")
	return true
}
