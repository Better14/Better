// Copyright authors of this Go fork
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
				check.errorf(fn.pos, BadDecl, "operator %s requires paired operator %s", fn.name, missing)
			}
		}
	}
}

func (check *Checker) operatorFuncs(name string) []*Func {
	return operatorFuncsInPackage(check.pkg, name)
}

func (check *Checker) operatorOverloads(name string) []*Func {
	if !check.pkgHasOperatorOverloads {
		return nil
	}
	if check.operatorOverloadsByName != nil {
		if funcs, ok := check.operatorOverloadsByName[name]; ok {
			return funcs
		}
	}
	funcs := check.computeOperatorOverloads(name)
	if check.operatorOverloadsByName == nil {
		check.operatorOverloadsByName = make(map[string][]*Func)
	}
	check.operatorOverloadsByName[name] = funcs
	return funcs
}

func (check *Checker) operatorPackagesForOperands(args ...*operand) []*Package {
	seen := make(map[*Package]bool)
	var pkgs []*Package
	add := func(p *Package) {
		if p != nil && !seen[p] {
			seen[p] = true
			pkgs = append(pkgs, p)
		}
	}
	add(check.pkg)
	for _, a := range args {
		add(check.pkgForRecv(a.typ()))
	}
	for _, imp := range check.imports {
		if imp != nil {
			add(imp.imported)
		}
	}
	return pkgs
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
	cands := check.operatorOverloads(name)
	if fn := check.lookupOverloadByArgTypes(name, args, cands); fn != nil {
		return fn
	}
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

// inOperatorOverloadImpl reports whether fn is the operator overload currently being implemented.
func (check *Checker) inOperatorOverloadImpl(fn *Func, name string) bool {
	return fn != nil && check.sig != nil && fn.name == name && fn.typ == check.sig
}

func (check *Checker) indexOperatorCall(pos syntax.Pos, name string, recvExpr syntax.Expr, indices []syntax.Expr, value syntax.Expr, fn *Func) *syntax.CallExpr {
	sig := fn.typ.(*Signature)
	if sig.Recv() != nil {
		args := append([]syntax.Expr{}, indices...)
		if value != nil {
			args = append(args, value)
		}
		return &syntax.CallExpr{
			Fun: &syntax.SelectorExpr{
				X:   recvExpr,
				Sel: syntax.NewName(pos, name),
			},
			ArgList: args,
		}
	}
	args := []syntax.Expr{recvExpr}
	args = append(args, indices...)
	if value != nil {
		args = append(args, value)
	}
	if pkgName := check.importNameFor(fn.pkg); pkgName != "" {
		return &syntax.CallExpr{
			Fun: &syntax.SelectorExpr{
				X:     syntax.NewName(pos, pkgName),
				Sel:   syntax.NewName(pos, name),
			},
			ArgList: args,
		}
	}
	return &syntax.CallExpr{
		Fun:     syntax.NewName(pos, name),
		ArgList: args,
	}
}

// selectIndexOperator picks a [] or []= overload for an index expression.
// A single generic candidate is accepted without assignability pre-check (inference
// happens in callOperator). Multiple candidates use silent overload resolution.
func (check *Checker) selectIndexOperator(cands []*Func, call *syntax.CallExpr, args []*operand) *Func {
	if len(cands) == 1 {
		return cands[0]
	}
	return check.selectOverloadSilent(call, cands, args)
}

func (check *Checker) importNameFor(pkg *Package) string {
	if pkg == nil || pkg == check.pkg {
		return ""
	}
	for _, imp := range check.imports {
		if imp != nil && imp.imported == pkg {
			return imp.name
		}
	}
	return ""
}

func (check *Checker) operatorCallExpr(pos syntax.Pos, fn *Func, argExprs []syntax.Expr) *syntax.CallExpr {
	if fn == nil {
		return &syntax.CallExpr{Fun: syntax.NewName(pos, ""), ArgList: argExprs}
	}
	if pkgName := check.importNameFor(fn.pkg); pkgName != "" {
		return &syntax.CallExpr{
			Fun: &syntax.SelectorExpr{
				X:     syntax.NewName(pos, pkgName),
				Sel:   syntax.NewName(pos, fn.name),
			},
			ArgList: argExprs,
		}
	}
	return &syntax.CallExpr{Fun: syntax.NewName(pos, fn.name), ArgList: argExprs}
}

func (check *Checker) callOperator(x *operand, pos syntax.Pos, fn *Func, call *syntax.CallExpr, argExprs []syntax.Expr, args []*operand, recordExpr syntax.Expr) *syntax.CallExpr {
	if call == nil {
		call = check.operatorCallExpr(pos, fn, argExprs)
	}
	check.expr(nil, x, call)
	if !x.isValid() {
		return nil
	}
	check.recordOperatorCall(recordExpr, call)
	return call
}

func (check *Checker) applyBinaryOperatorOverload(x, y *operand, e syntax.Expr, lhs, rhs syntax.Expr, op syntax.Operator) bool {
	if op == syntax.NullCoalesce || op == syntax.AndAnd || op == syntax.OrOr {
		return false
	}
	if !check.pkgHasOperatorOverloads {
		return false
	}
	name := op.String()
	if len(check.operatorOverloads(name)) == 0 {
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
	if check.inOperatorOverloadImpl(fn, name) {
		return false
	}
	pos := lhs.Pos()
	if e != nil {
		pos = e.Pos()
	}
	check.callOperator(x, pos, fn, nil, []syntax.Expr{lhs, rhs}, []*operand{x, y}, e)
	if x.isValid() {
		if e != nil {
			x.expr = e
		}
	}
	return true
}

func (check *Checker) tryUnaryOperatorOverload(x *operand, e *syntax.Operation) bool {
	op := e.Op
	if op == syntax.And || op == syntax.Recv || op == syntax.Mul || op == syntax.Tilde {
		return false
	}
	if !check.pkgHasOperatorOverloads {
		return false
	}
	name := op.String()
	if len(check.operatorOverloads(name)) == 0 {
		return false
	}
	fn := check.lookupUnaryOperatorExact(name, x)
	if fn == nil {
		fn = check.selectOperatorFunc(name, 1, []*operand{x})
	}
	if fn == nil {
		return false
	}
	if check.inOperatorOverloadImpl(fn, name) {
		return false
	}
	call := check.callOperator(x, e.Pos(), fn, nil, []syntax.Expr{e.X}, []*operand{x}, e)
	if x.isValid() {
		x.expr = e
	}
	return call != nil
}

// tryIndexOperatorOverload handles a[i] when the type of a defines func [](a, i...) U.
// base holds the type-checked receiver when indexExpr has already evaluated e.X.
func (check *Checker) tryIndexOperatorOverload(x *operand, e *syntax.IndexExpr, base *operand) bool {
	var l operand
	if base != nil && base.isValid() {
		l = *base
	} else {
		check.expr(nil, &l, e.X)
	}
	if !l.isValid() {
		return false
	}
	if !check.isComplete(l.typ()) || supportsBuiltinIndex(l.typ()) {
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
	call := check.indexOperatorCall(e.Pos(), "[]", e.X, indices, nil, cands[0])
	fn := check.selectIndexOperator(cands, call, preload)
	if fn == nil {
		return false
	}
	call = check.indexOperatorCall(e.Pos(), "[]", e.X, indices, nil, fn)
	check.expr(nil, x, call)
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
	if !l.isValid() {
		return false
	}
	if !check.isComplete(l.typ()) || supportsBuiltinIndex(l.typ()) {
		return false
	}
	indices := check.overloadIndices(idx)
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
	call := check.indexOperatorCall(lhs.Pos(), "[]=", idx.X, indices, rhs, cands[0])
	fn := check.selectIndexOperator(cands, call, preload)
	if fn == nil {
		return false
	}
	if x == nil {
		x = new(operand)
	}
	call = check.indexOperatorCall(lhs.Pos(), "[]=", idx.X, indices, rhs, fn)
	check.rawExpr(nil, x, call, nil, true)
	check.record(x)
	if !x.isValid() || x.mode() != novalue {
		return false
	}
	check.recordIndexAssignCall(idx, call)
	x.mode_ = novalue
	x.expr = lhs
	return true
}

func (check *Checker) tryIncDecOperatorOverload(s *syntax.AssignStmt, op syntax.Operator) bool {
	if !check.pkgHasOperatorOverloads {
		return false
	}
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
	if check.inOperatorOverloadImpl(fn, name) {
		return false
	}
	var res operand
	call := check.callOperator(&res, s.Pos(), fn, nil, []syntax.Expr{s.Lhs}, []*operand{&arg}, nil)
	if !res.isValid() || call == nil {
		return false
	}
	check.recordOperatorAssignCall(s, call)
	check.assignVar(s.Lhs, nil, &res, "assignment")
	return true
}
