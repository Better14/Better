// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "cmd/compile/internal/syntax"

type linqRecvKind int

const (
	linqRecvInvalid linqRecvKind = iota
	linqRecvSlice
	linqRecvArray
	linqRecvLazy
)

type linqMethodInfo struct {
	slice string // package func when receiver is slice/array
	lazy  string // package func when receiver is linq.Lazy[T]
}

var linqMethods = map[string]linqMethodInfo{
	"Where":               {slice: "Where", lazy: "LazyWhere"},
	"Select":              {slice: "Select", lazy: "LazySelectBy"},
	"OrderBy":             {slice: "OrderBy", lazy: "LazyOrderBy"},
	"OrderByDescending":   {slice: "OrderByDescending", lazy: "LazyOrderByDescending"},
	"Take":                {slice: "Take", lazy: "LazyTake"},
	"Skip":                {slice: "Skip", lazy: "LazySkip"},
	"Distinct":            {slice: "Distinct", lazy: "LazyDistinct"},
	"GroupBy":             {slice: "GroupBy", lazy: "LazyGroupBy"},
	"ToList":              {slice: "ToList", lazy: "ToListLazy"},
	"First":               {slice: "First", lazy: "FirstLazy"},
	"FirstOrDefault":      {slice: "FirstOrDefault", lazy: "FirstOrDefaultLazy"},
	"Sum":                 {slice: "Sum", lazy: "LazySum"},
	"Any":                 {slice: "Any", lazy: "LazyAny"},
	"All":                 {slice: "All", lazy: "LazyAll"},
	"Aggregate":           {slice: "Aggregate", lazy: "LazyAggregate"},
}

func (check *Checker) linqRecvKind(typ Type) linqRecvKind {
	typ = Unalias(typ)
	switch u := typ.Underlying().(type) {
	case *Slice:
		return linqRecvSlice
	case *Array:
		return linqRecvArray
	case *Named:
		obj := u.Origin().obj
		if obj != nil && obj.pkg != nil && obj.pkg.path == "linq" && obj.name == "Lazy" {
			return linqRecvLazy
		}
	}
	return linqRecvInvalid
}

func (check *Checker) isLinqPackageCall(call *syntax.CallExpr) bool {
	sel, ok := call.Fun.(*syntax.SelectorExpr)
	if !ok {
		return false
	}
	name, ok := sel.X.(*syntax.Name)
	if !ok {
		return false
	}
	if pn, ok := check.lookup(name.Value).(*PkgName); ok && pn.imported != nil && pn.imported.path == "linq" {
		return true
	}
	return false
}

func (check *Checker) linqPkgName() *PkgName {
	for _, pn := range check.imports {
		if pn.imported != nil && pn.imported.path == "linq" {
			return pn
		}
	}
	return nil
}

func (check *Checker) tryLinqCall(x *operand, call *syntax.CallExpr, sel *syntax.SelectorExpr) (exprKind, bool) {
	method := sel.Sel.Value
	info, ok := linqMethods[method]
	if !ok {
		return statement, false
	}
	if !check.verifyVersionf(call, go1_27, "LINQ method %s", method) {
		x.invalidate()
		x.expr = call
		return statement, true
	}

	// Do not intercept package-qualified calls such as constanttime.Select.
	if name, ok := sel.X.(*syntax.Name); ok {
		if _, ok := check.lookup(name.Value).(*PkgName); ok {
			return statement, false
		}
	}

	var recv operand
	check.expr(nil, &recv, sel.X)
	if !recv.isValid() {
		x.invalidate()
		x.expr = call
		return statement, true
	}

	rk := check.linqRecvKind(recv.typ())
	if rk == linqRecvInvalid {
		return statement, false
	}

	funcName := info.slice
	if rk == linqRecvLazy {
		funcName = info.lazy
	}

	pname := check.linqPkgName()
	if pname == nil {
		return statement, false
	}

	recvExpr := sel.X
	if rk == linqRecvArray {
		recvExpr = &syntax.SliceExpr{X: sel.X}
	}

	check.typecheckLinqLambdas(call, pname, funcName, recv.typ())

	argList := make([]syntax.Expr, 1+len(call.ArgList))
	argList[0] = recvExpr
	copy(argList[1:], call.ArgList)

	call.Fun = &syntax.SelectorExpr{
		X:   syntax.NewName(call.Pos(), pname.name),
		Sel: syntax.NewName(call.Pos(), funcName),
	}
	call.ArgList = argList
	if n, ok := call.Fun.(*syntax.SelectorExpr).X.(*syntax.Name); ok {
		check.recordUse(n, pname)
	}
	return check.callExpr(x, call), true
}

func linqElemType(typ Type) Type {
	typ = Unalias(typ)
	switch u := typ.Underlying().(type) {
	case *Slice:
		return u.elem
	case *Array:
		return u.elem
	case *Named:
		obj := u.Origin().obj
		if obj != nil && obj.pkg != nil && obj.pkg.path == "linq" && obj.name == "Lazy" {
			if targs := u.TypeArgs(); targs != nil && targs.Len() > 0 {
				return targs.At(0)
			}
		}
	}
	return nil
}

func (check *Checker) typecheckLinqLambdas(call *syntax.CallExpr, pname *PkgName, funcName string, recv Type) {
	obj := pname.imported.scope.Lookup(funcName)
	fn, ok := obj.(*Func)
	if !ok || fn.typ() == nil {
		return
	}
	sig, ok := fn.typ().(*Signature)
	if !ok {
		return
	}
	instSig := sig
	if elem := linqElemType(recv); elem != nil && sig.TypeParams().Len() > 0 {
		instSig = check.instantiateSignature(call.Pos(), call, sig, []Type{elem}, nil)
		if instSig == nil {
			return
		}
	}
	for i, e := range call.ArgList {
		lam, ok := e.(*syntax.LambdaExpr)
		if !ok {
			continue
		}
		par := i + 1 // receiver is the first argument after desugaring
		if par >= instSig.params.Len() {
			continue
		}
		var x operand
		check.rawExpr(nil, &x, lam, instSig.params.At(par).typ, false)
		if x.isValid() {
			check.recordTypeAndValue(lam, value, x.typ(), nil)
		}
	}
}
