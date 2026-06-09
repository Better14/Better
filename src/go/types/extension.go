// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	. "internal/types/errors"
)

// extensionFuncShape reports whether fn looks like an exported extension
// (receiver lowered to the first parameter) from pkg.
func extensionFuncShape(pkg *Package, fn *Func) bool {
	if fn == nil || pkg == nil || fn.pkg != pkg {
		return false
	}
	sig := fn.Signature()
	if sig == nil || sig.Recv() != nil || sig.Params() == nil || sig.Params().Len() == 0 {
		return false
	}
	typ := sig.Params().At(0).Type()
	typ, _ = deref(typ)
	typ = Unalias(typ)
	switch typ.(type) {
	case *Basic, *Slice, *Array, *Map, *Chan:
		return true
	case *Named:
		if n := typ.(*Named); n.obj != nil && n.obj.pkg != nil {
			return n.obj.pkg != pkg
		}
	}
	return false
}

// extensionSliceElemTypeParam reports whether rtyp is []T with a single
// identifier element name T that is not yet defined in scope.
func extensionSliceElemTypeParam(rtyp ast.Expr) (*ast.Ident, bool) {
	st, ok := ast.Unparen(rtyp).(*ast.ArrayType)
	if !ok || st.Len != nil {
		return nil, false
	}
	name, ok := ast.Unparen(st.Elt).(*ast.Ident)
	if !ok || name.Name == "" || name.Name == "_" {
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
		return T.obj.pkg != check.pkg || isCGoTypeObj(check.fset, T.obj)
	default:
		return false
	}
}

// extensionSliceRecvTypeParam reports whether recv/rparams describe an extension
// on []T with T declared by the receiver element type.
func extensionSliceRecvTypeParam(recv *Var, rparams *TypeParamList) (*TypeParam, bool) {
	if recv == nil || rparams == nil || rparams.Len() != 1 {
		return nil, false
	}
	sl, ok := Unalias(recv.typ).Underlying().(*Slice)
	if !ok {
		return nil, false
	}
	tp := rparams.At(0)
	if Identical(sl.elem, tp) {
		return tp, true
	}
	return nil, false
}

// prepareReceiverMethodTypeParams handles a method type parameter list whose first
// entry restates a receiver type parameter (e.g. Where[T any] on []T or Lazy[T]).
// When required is true (extension []T), list must be non-empty and start with T.
// Additional entries are returned for collectTypeParams (e.g. Select[T, U any] → [U any]).
func (check *Checker) prepareReceiverMethodTypeParams(recvTPar *TypeParam, list []*ast.Field, at positioner, required bool) []*ast.Field {
	recvName := recvTPar.obj.name
	if len(list) == 0 {
		if required {
			check.errorf(at, BadDecl, "extension method on []%s must declare type parameter %s (e.g. …[%s any](…))", recvName, recvName, recvName)
		}
		return nil
	}
	if len(list[0].Names) == 0 || list[0].Names[0].Name != recvName {
		if required {
			check.errorf(atPos(list[0].Pos()), BadDecl, "first type parameter must be %s", recvName)
		}
		return list
	}
	bound := check.bound(list[0].Type)
	if isValid(bound) {
		recvTPar.SetConstraint(bound)
	}
	return list[1:]
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
	check.indexExtensionFunc(obj)
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
	adapt   bool // wrap receiver with slices.Names
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
		for _, fn := range extensionFuncsInPackage(pkg, method) {
			if seen[fn] {
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
	// Match concrete slices against extension signatures on []T (element type param).
	if recvSl, ok := Unalias(recv).Underlying().(*Slice); ok {
		if paramSl, ok := Unalias(param).Underlying().(*Slice); ok {
			if tp, ok := paramSl.elem.(*TypeParam); ok && isValid(recvSl.elem) {
				c := tp.Constraint()
				if !isValid(c) || Identical(Unalias(c), universeAny.Type()) {
					return true
				}
			}
			x.typ_ = recvSl.elem
			if ok, _ := x.assignableTo(check, paramSl.elem, nil); ok {
				return true
			}
		}
	}
	return false
}

func (check *Checker) tryExtensionCall(x *operand, call *ast.CallExpr, sel *ast.SelectorExpr) (exprKind, bool) {
	// Do not intercept package-qualified calls.
	if name, ok := sel.X.(*ast.Ident); ok {
		if _, ok := check.lookup(name.Name).(*PkgName); ok {
			return statement, false
		}
	}

	var recv operand
	check.expr(nil, &recv, sel.X)
	if !recv.isValid() {
		return statement, false
	}

	method := sel.Sel.Name
	if check.hasInstanceMethod(recv.typ(), recv.mode() == variable, method) {
		return statement, false
	}
	if obj, _, _ := lookupFieldOrMethod(recv.typ(), recv.mode() == variable, check.pkg, method, false); obj != nil {
		if _, ok := obj.(*Var); ok {
			return statement, false // field value (including func-typed fields)
		}
	}

	matches := check.extensionCandidates(recv.typ(), method)
	if len(matches) == 0 {
		return statement, false
	}

	if !check.verifyVersionf(call, go1_27, "extension method %s", sel.Sel.Name) {
		return statement, false
	}

	if len(matches) > 1 {
		check.errorf(call, AmbiguousSelector, "ambiguous extension call %s.%s", recv.expr, sel.Sel.Name)
		x.invalidate()
		x.expr = call
		return statement, true
	}
	m := matches[0]

	recvExpr := sel.X
	if m.slice {
		recvExpr = &ast.SliceExpr{X: sel.X}
	}
	if m.adapt {
		if !check.verifyVersionf(call, go1_27, "slices.Names") {
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
		recvExpr = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   astNewIdent(call.Pos(), slicesPkg.name),
				Sel: astNewIdent(call.Pos(), "Values"),
			},
			Args: []ast.Expr{sel.X},
		}
		if n, ok := recvExpr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident); ok {
			check.recordUse(n, slicesPkg)
		}
	}

	pkgIdent := m.pkgName
	if pkgIdent == nil {
		// same package: use an unqualified function name
		call.Fun = astNewIdent(call.Pos(), sel.Sel.Name)
	} else {
		call.Fun = &ast.SelectorExpr{
			X:   astNewIdent(call.Pos(), pkgIdent.name),
			Sel: astNewIdent(call.Pos(), sel.Sel.Name),
		}
		check.recordUse(call.Fun.(*ast.SelectorExpr).X.(*ast.Ident), pkgIdent)
	}

	argList := make([]ast.Expr, 1+len(call.Args))
	argList[0] = recvExpr
	copy(argList[1:], call.Args)
	call.Args = argList

	return check.callExpr(x, call), true
}
