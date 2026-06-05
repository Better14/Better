// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

// extensionSliceElemTypeParam reports whether rtyp is []T with a single
// identifier element name T that is not yet defined in scope.
func extensionSliceElemTypeParam(rtyp syntax.Expr) (*syntax.Name, bool) {
	st, ok := syntax.Unparen(rtyp).(*syntax.SliceType)
	if !ok {
		return nil, false
	}
	name, ok := syntax.Unparen(st.Elem).(*syntax.Name)
	if !ok || name.Value == "" || name.Value == "_" {
		return nil, false
	}
	return name, true
}

// isExtensionRecv reports whether typ is a valid extension receiver base type
// for a method declared in defPkg.
func (check *Checker) isExtensionRecv(typ Type) bool {
	typ, _ = deref(typ)
	typ = Unalias(typ)
	if !isValid(typ) {
		return false
	}
	switch T := typ.(type) {
	case *Basic:
		return true
	case *Slice, *Array, *Map, *Chan:
		return true
	case *Named:
		if T.obj == nil || T.obj.pkg == nil {
			return false
		}
		return T.obj.pkg != check.pkg || isCGoTypeObj(T.obj)
	default:
		return false
	}
}

// finishExtensionFunc lowers an extension method to a package-level function by
// moving the receiver to the first parameter and merging receiver type parameters
// into the function's type parameter list.
func (check *Checker) finishExtensionFunc(obj *Func, sig *Signature) {
	obj.isExtension_ = true

	recv := sig.recv
	sig.recv = nil

	if sig.rparams != nil {
		rlist := sig.rparams.list()
		sig.rparams = nil
		if sig.tparams == nil {
			sig.tparams = &TypeParamList{tparams: rlist}
		} else {
			combined := append(rlist, sig.tparams.list()...)
			for i, t := range combined {
				t.index = i
			}
			sig.tparams = &TypeParamList{tparams: combined}
		}
	}

	if sig.params == nil {
		sig.params = NewTuple(recv)
	} else {
		sig.params = NewTuple(append([]*Var{recv}, sig.params.vars...)...)
	}
}

// iterSeqElem returns the element type T if typ is (or is an instance of) iter.Seq[T].
func iterSeqElem(typ Type) Type {
	typ = Unalias(typ)
	n, ok := typ.(*Named)
	if !ok || n.obj == nil || n.obj.pkg == nil || n.obj.pkg.path != "iter" || n.obj.name != "Seq" {
		return nil
	}
	if targs := n.TypeArgs(); targs != nil && targs.Len() == 1 {
		return targs.At(0)
	}
	return nil
}

func sliceOrArrayElem(typ Type) Type {
	typ = Unalias(typ)
	switch u := typ.Underlying().(type) {
	case *Slice:
		return u.elem
	case *Array:
		return u.elem
	}
	return nil
}

type extensionMatch struct {
	fn      *Func
	pkgName *PkgName
	adapt   bool // wrap receiver with slices.Values
	slice   bool // convert array receiver to slice for []T extensions
}

// hasInstanceMethod reports whether typ has a concrete or interface method name.
func (check *Checker) hasInstanceMethod(typ Type, addressable bool, name string) bool {
	obj, _, _ := lookupFieldOrMethod(typ, addressable, check.pkg, name, false)
	_, ok := obj.(*Func)
	return ok
}

func (check *Checker) extensionCandidates(recvType Type, method string) []extensionMatch {
	var out []extensionMatch
	seen := make(map[*Func]bool)

	addPkg := func(pkg *Package, pkgName *PkgName) {
		if pkg == nil || pkg.scope == nil {
			return
		}
		scope := pkg.scope
		for _, name := range scope.Names() {
			if name != method {
				continue
			}
			obj := scope.Lookup(name)
			fn, ok := obj.(*Func)
			if !ok || !fn.IsExtension() || seen[fn] {
				continue
			}
			if m, ok := check.matchExtension(recvType, fn); ok {
				seen[fn] = true
				m.pkgName = pkgName
				out = append(out, m)
			}
		}
	}

	addPkg(check.pkg, nil)
	for _, imp := range check.imports {
		if imp.imported != nil {
			addPkg(imp.imported, imp)
		}
	}
	return out
}

func (check *Checker) matchExtension(recvType Type, fn *Func) (extensionMatch, bool) {
	sig := fn.Signature()
	if sig.params == nil || sig.params.Len() == 0 {
		return extensionMatch{}, false
	}
	param0 := sig.params.At(0).typ

	if check.extensionTypesMatch(recvType, param0) {
		return extensionMatch{fn: fn}, true
	}

	// [N]T can use extensions declared on []T.
	if arr, ok := Unalias(recvType).Underlying().(*Array); ok {
		if sl, ok := Unalias(param0).Underlying().(*Slice); ok && Identical(arr.elem, sl.elem) {
			return extensionMatch{fn: fn, slice: true}, true
		}
	}

	if seqElem := iterSeqElem(param0); seqElem != nil {
		if elem := sliceOrArrayElem(recvType); elem != nil && Identical(elem, seqElem) {
			return extensionMatch{fn: fn, adapt: true}, true
		}
	}
	return extensionMatch{}, false
}

func (check *Checker) extensionTypesMatch(recv, param Type) bool {
	if Identical(recv, param) {
		return true
	}
	var x operand
	x.mode_ = value
	x.typ_ = recv
	if ok, _ := x.assignableTo(check, param, nil); ok {
		return true
	}
	return false
}

func (check *Checker) tryExtensionCall(x *operand, call *syntax.CallExpr, sel *syntax.SelectorExpr) (exprKind, bool) {
	if !check.verifyVersionf(call, go1_27, "extension method %s", sel.Sel.Value) {
		return statement, false
	}

	// Do not intercept package-qualified calls.
	if name, ok := sel.X.(*syntax.Name); ok {
		if _, ok := check.lookup(name.Value).(*PkgName); ok {
			return statement, false
		}
	}

	var recv operand
	check.expr(nil, &recv, sel.X)
	if !recv.isValid() {
		return statement, false
	}

	method := sel.Sel.Value
	if check.hasInstanceMethod(recv.typ(), recv.mode() == variable, method) {
		return statement, false
	}

	matches := check.extensionCandidates(recv.typ(), method)
	if len(matches) == 0 {
		return statement, false
	}
	if len(matches) > 1 {
		check.errorf(call, AmbiguousSelector, "ambiguous extension call %s.%s", recv.expr, sel.Sel.Value)
		x.invalidate()
		x.expr = call
		return statement, true
	}
	m := matches[0]

	recvExpr := sel.X
	if m.slice {
		recvExpr = &syntax.SliceExpr{X: sel.X}
	}
	if m.adapt {
		if !check.verifyVersionf(call, go1_27, "slices.Values") {
			x.invalidate()
			x.expr = call
			return statement, true
		}
		var slicesPkg *PkgName
		for _, imp := range check.imports {
			if imp.imported != nil && imp.imported.path == "slices" {
				slicesPkg = imp
				break
			}
		}
		if slicesPkg == nil {
			check.errorf(call, UndeclaredName, "extension on slice requires import \"slices\"")
			x.invalidate()
			x.expr = call
			return statement, true
		}
		recvExpr = &syntax.CallExpr{
			Fun: &syntax.SelectorExpr{
				X:   syntax.NewName(call.Pos(), slicesPkg.name),
				Sel: syntax.NewName(call.Pos(), "Values"),
			},
			ArgList: []syntax.Expr{sel.X},
		}
		if n, ok := recvExpr.(*syntax.CallExpr).Fun.(*syntax.SelectorExpr).X.(*syntax.Name); ok {
			check.recordUse(n, slicesPkg)
		}
	}

	pkgIdent := m.pkgName
	if pkgIdent == nil {
		// same package
		call.Fun = &syntax.SelectorExpr{
			X:   syntax.NewName(call.Pos(), check.pkg.name),
			Sel: syntax.NewName(call.Pos(), sel.Sel.Value),
		}
	} else {
		call.Fun = &syntax.SelectorExpr{
			X:   syntax.NewName(call.Pos(), pkgIdent.name),
			Sel: syntax.NewName(call.Pos(), sel.Sel.Value),
		}
		check.recordUse(call.Fun.(*syntax.SelectorExpr).X.(*syntax.Name), pkgIdent)
	}

	argList := make([]syntax.Expr, 1+len(call.ArgList))
	argList[0] = recvExpr
	copy(argList[1:], call.ArgList)
	call.ArgList = argList

	return check.callExpr(x, call), true
}
