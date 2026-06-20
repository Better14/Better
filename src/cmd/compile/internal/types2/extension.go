// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
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

// extensionMapTypeParams reports whether rtyp is map[K]V with identifier
// key K and value V that are not yet defined in scope.
func extensionMapTypeParams(rtyp syntax.Expr) (key, val *syntax.Name, ok bool) {
	mt, ok := syntax.Unparen(rtyp).(*syntax.MapType)
	if !ok {
		return nil, nil, false
	}
	key, ok1 := syntax.Unparen(mt.Key).(*syntax.Name)
	val, ok2 := syntax.Unparen(mt.Value).(*syntax.Name)
	if !ok1 || !ok2 || key.Value == "" || key.Value == "_" || val.Value == "" || val.Value == "_" {
		return nil, nil, false
	}
	return key, val, true
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

// extensionSliceRecvTypeParam reports whether recv/rparams describe an extension
// on []T with T declared by the receiver element type.
func extensionSliceRecvTypeParam(recv *Var, rparams *TypeParamList) (*TypeParam, bool) {
	if recv == nil || rparams == nil || rparams.Len() != 1 {
		return nil, false
	}
	// Named types such as FlatArray[T] with underlying []T are ordinary methods, not extensions.
	if _, ok := Unalias(recv.typ).(*Named); ok {
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

// extensionMapRecvTypeParams reports whether recv/rparams describe an extension
// on map[K]V with K and V declared by the receiver map type.
func extensionMapRecvTypeParams(recv *Var, rparams *TypeParamList) (*TypeParam, *TypeParam, bool) {
	if recv == nil || rparams == nil || rparams.Len() != 2 {
		return nil, nil, false
	}
	if _, ok := Unalias(recv.typ).(*Named); ok {
		return nil, nil, false
	}
	m, ok := Unalias(recv.typ).Underlying().(*Map)
	if !ok {
		return nil, nil, false
	}
	k := rparams.At(0)
	v := rparams.At(1)
	if Identical(m.key, k) && Identical(m.elem, v) {
		return k, v, true
	}
	return nil, nil, false
}

// prepareReceiverMethodTypeParams handles a method type parameter list whose first
// entry restates a receiver type parameter (e.g. Where[T any] on []T or Lazy[T]).
// When required is true (extension []T), list must be non-empty and start with T.
// Additional entries are returned for collectTypeParams (e.g. Select[T, U any] → [U any]).
func (check *Checker) prepareReceiverMethodTypeParams(recvTPar *TypeParam, list []*syntax.Field, at poser, required bool) []*syntax.Field {
	recvName := recvTPar.obj.name
	if len(list) == 0 {
		if required {
			check.errorf(at, BadDecl, "extension method on []%s must declare type parameter %s (e.g. …[%s any](…))", recvName, recvName, recvName)
		}
		return nil
	}
	if list[0].Name == nil || list[0].Name.Value != recvName {
		if required {
			check.errorf(list[0].Pos(), BadDecl, "first type parameter must be %s", recvName)
		}
		return list
	}
	bound := check.bound(list[0].Type)
	if isValid(bound) {
		recvTPar.SetConstraint(bound)
	}
	return list[1:]
}

// prepareReceiverMapMethodTypeParams handles method type parameters for extensions
// on map[K]V; the first two entries must restate K and V from the receiver.
func (check *Checker) prepareReceiverMapMethodTypeParams(keyPar, valPar *TypeParam, list []*syntax.Field, at poser, required bool) []*syntax.Field {
	keyName := keyPar.obj.name
	valName := valPar.obj.name
	if len(list) < 2 {
		if required {
			check.errorf(at, BadDecl, "extension method on map[%s]%s must declare type parameters %s and %s (e.g. …[%s comparable, %s any](…))", keyName, valName, keyName, valName, keyName, valName)
		}
		return nil
	}
	if list[0].Name == nil || list[0].Name.Value != keyName {
		if required {
			check.errorf(list[0].Pos(), BadDecl, "first type parameter must be %s", keyName)
		}
		return list
	}
	if list[1].Name == nil || list[1].Name.Value != valName {
		if required {
			check.errorf(list[1].Pos(), BadDecl, "second type parameter must be %s", valName)
		}
		return list
	}
	if bound := check.bound(list[0].Type); isValid(bound) {
		keyPar.SetConstraint(bound)
	}
	if bound := check.bound(list[1].Type); isValid(bound) {
		valPar.SetConstraint(bound)
	}
	return list[2:]
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
	fn       *Func
	pkgName  *PkgName
	adapt    bool   // wrap receiver with slices.Values
	slice    bool   // convert array receiver to slice for []T extensions
	linqFast string // unexported linq []T fast path (compiler-only)
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
		if elem := sliceOrArrayElem(recvType); elem != nil && extensionSeqElemMatch(check, elem, seqElem) {
			m := extensionMatch{fn: fn, adapt: true}
			if fn.pkg != nil && fn.pkg.path == linqPkgPath {
				if fast, ok := linqSliceFastPath(fn.name); ok {
					m.linqFast = fast
				}
			}
			return m, true
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
			if extensionSeqElemMatch(check, recvSl.elem, paramSl.elem) {
				return true
			}
		}
	}
	// Match concrete maps against extension signatures on map[K]V.
	if recvMap, ok := Unalias(recv).Underlying().(*Map); ok {
		if paramMap, ok := Unalias(param).Underlying().(*Map); ok {
			if extensionSeqElemMatch(check, recvMap.key, paramMap.key) &&
				extensionSeqElemMatch(check, recvMap.elem, paramMap.elem) {
				return true
			}
		}
	}
	// Match concrete iter.Seq[E] against extension signatures on iter.Seq[T].
	if recvElem := iterSeqElem(recv); recvElem != nil {
		if paramElem := iterSeqElem(param); paramElem != nil {
			if extensionSeqElemMatch(check, recvElem, paramElem) {
				return true
			}
		}
	}
	return false
}

func extensionSeqElemMatch(check *Checker, elem, pattern Type) bool {
	if Identical(elem, pattern) {
		return true
	}
	if tp, ok := pattern.(*TypeParam); ok && isValid(elem) {
		c := tp.Constraint()
		if !isValid(c) || Identical(Unalias(c), universeAny.Type()) {
			return true
		}
		if check != nil && check.implements(elem, c, true, nil) {
			return true
		}
		var x operand
		x.mode_ = value
		x.typ_ = elem
		if ok, _ := x.assignableTo(check, pattern, nil); ok {
			return ok
		}
	}
	return false
}

// extensionMethodExists reports whether method is declared as an extension
// in the current package or any import, without type-checking receivers.
func (check *Checker) extensionMethodExists(method string) bool {
	if len(extensionFuncsInPackage(check.pkg, method)) > 0 {
		return true
	}
	for _, imp := range check.imports {
		if imp.imported != nil && len(extensionFuncsInPackage(imp.imported, method)) > 0 {
			return true
		}
	}
	return false
}

func (check *Checker) tryExtensionCall(x *operand, call *syntax.CallExpr, sel *syntax.SelectorExpr, inst *syntax.IndexExpr) (exprKind, bool) {
	method := sel.Sel.Value

	// Cheap name lookup only; avoids re-evaluating long receiver chains for
	// ordinary method calls (test/torture.go).
	if !check.extensionMethodExists(method) {
		return statement, false
	}

	// Extension probing evaluates the full receiver via exprOrType, which
	// re-enters callExpr on nested calls in long chains. When already probing,
	// the name check above is enough to gate nested work.
	if !check.inExtensionProbe {
		check.inExtensionProbe = true
		defer func() { check.inExtensionProbe = false }()
	}

	// Do not intercept package-qualified calls or other selector expressions
	// whose receiver is a bare identifier. Evaluating the identifier alone
	// would report "use of package X not in selector" before the ordinary
	// call checker can handle pkg.Func.
	if name, ok := sel.X.(*syntax.Name); ok {
		obj := check.lookup(name.Value)
		if obj == nil {
			return statement, false
		}
		if _, ok := obj.(*PkgName); ok {
			return statement, false
		}
	}

	var recv operand
	check.exprOrType(&recv, sel.X, true)
	if !recv.isValid() {
		return statement, false
	}
	if recv.mode() == typexpr {
		// Method expression (T.m)(args), not an extension call.
		return statement, false
	}

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
	if len(matches) > 1 {
		var direct []extensionMatch
		for _, m := range matches {
			if !m.adapt {
				direct = append(direct, m)
			}
		}
		if len(direct) > 0 {
			matches = direct
		}
	}

	if !check.verifyVersionf(call, go1_27, "extension method %s", sel.Sel.Value) {
		return statement, false
	}

	var m extensionMatch
	if len(matches) == 1 {
		m = matches[0]
	} else {
		funcs := make([]*Func, len(matches))
		for i, match := range matches {
			funcs[i] = match.fn
		}
		var argOps []*operand
		recvArg := recv
		argOps = append(argOps, &recvArg)
		for _, arg := range call.ArgList {
			var a operand
			check.expr(nil, &a, arg)
			if !a.isValid() {
				x.invalidate()
				x.expr = call
				return statement, true
			}
			argOps = append(argOps, &a)
		}
		fn := check.selectOverloadSilent(call, funcs, argOps)
		if fn == nil {
			check.errorf(call, AmbiguousSelector, "ambiguous extension call %s.%s", recv.expr, sel.Sel.Value)
			x.invalidate()
			x.expr = call
			return statement, true
		}
		m = matches[0]
		for _, cand := range matches {
			if cand.fn == fn {
				m = cand
				break
			}
		}
	}

	recvExpr := sel.X
	if m.slice {
		recvExpr = &syntax.SliceExpr{X: sel.X}
	}
	funcName := m.fn.LinkName()
	useLinqFast := m.linqFast != "" && (m.pkgName == nil || m.adapt)
	if useLinqFast && m.pkgName != nil && m.pkgName.imported.scope.Lookup(m.linqFast) == nil {
		useLinqFast = false
	}
	if useLinqFast {
		funcName = m.linqFast
	} else if m.adapt {
		if !check.verifyVersionf(call, go1_27, "slices.Values") {
			x.invalidate()
			x.expr = call
			return statement, true
		}
		var slicesPkg *PkgName
		slicesPkg = check.ensureImported(call.Pos(), "slices")
		if slicesPkg == nil {
			check.error(call, BrokenImport, "could not import slices")
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
		// same package: use an unqualified function name
		call.Fun = syntax.NewName(call.Pos(), funcName)
	} else {
		call.Fun = &syntax.SelectorExpr{
			X:   syntax.NewName(call.Pos(), pkgIdent.name),
			Sel: syntax.NewName(call.Pos(), funcName),
		}
		check.recordUse(call.Fun.(*syntax.SelectorExpr).X.(*syntax.Name), pkgIdent)
	}
	if inst != nil {
		call.Fun = &syntax.IndexExpr{X: call.Fun, Index: inst.Index}
	}

	argList := make([]syntax.Expr, 1+len(call.ArgList))
	argList[0] = recvExpr
	copy(argList[1:], call.ArgList)
	call.ArgList = argList

	return check.callExpr(x, call, nil), true
}
