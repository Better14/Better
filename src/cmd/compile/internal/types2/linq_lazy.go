// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "cmd/compile/internal/syntax"

type linqMethodInfo struct {
	slice string
	lazy  string
}

// linqMethods maps LINQ method calls to package functions when import "linq".
// Slice []T uses this desugaring; linq.Lazy[T] uses it for methods that cannot
// yet be generic receiver methods (Select, OrderBy, …) due to export ICEs.
var linqMethods = map[string]linqMethodInfo{
	"Where":             {slice: "Where", lazy: "LazyWhere"},
	"Select":            {slice: "Select", lazy: "LazySelectBy"},
	"OrderBy":           {slice: "OrderBy", lazy: "LazyOrderBy"},
	"OrderByDescending": {slice: "OrderByDescending", lazy: "LazyOrderByDescending"},
	"Take":              {slice: "Take", lazy: "LazyTake"},
	"Skip":              {slice: "Skip", lazy: "LazySkip"},
	"Distinct":          {slice: "Distinct", lazy: "LazyDistinct"},
	"GroupBy":           {slice: "GroupBy", lazy: "LazyGroupBy"},
	"ToList":            {slice: "ToList", lazy: "ToListLazy"},
	"First":             {slice: "First", lazy: "FirstLazy"},
	"FirstOrDefault":    {slice: "FirstOrDefault", lazy: "FirstOrDefaultLazy"},
	"Sum":               {slice: "Sum", lazy: "LazySum"},
	"Any":               {slice: "Any", lazy: "LazyAny"},
	"All":               {slice: "All", lazy: "LazyAll"},
	"Aggregate":         {slice: "Aggregate", lazy: "LazyAggregate"},
}

type linqRecvKind int

const (
	linqRecvInvalid linqRecvKind = iota
	linqRecvSlice
	linqRecvArray
	linqRecvLazy
)

func (check *Checker) linqRecvKind(typ Type) linqRecvKind {
	typ = Unalias(typ)
	if n, ok := typ.(*Named); ok {
		obj := n.Origin().obj
		if obj != nil && obj.pkg != nil && obj.pkg.path == "linq" && obj.name == "Lazy" {
			return linqRecvLazy
		}
	}
	switch typ.Underlying().(type) {
	case *Slice:
		return linqRecvSlice
	case *Array:
		return linqRecvArray
	}
	return linqRecvInvalid
}

func (check *Checker) linqPkgName() *PkgName {
	for _, pn := range check.imports {
		if pn.imported != nil && pn.imported.path == "linq" {
			return pn
		}
	}
	return nil
}

func (check *Checker) tryLazyCall(x *operand, call *syntax.CallExpr, sel *syntax.SelectorExpr) (exprKind, bool) {
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
