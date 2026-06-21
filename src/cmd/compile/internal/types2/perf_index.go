// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "strings"

// operatorTypeKey identifies a binary operator overload by operand types.
type operatorTypeKey struct {
	left  string
	right string
}

func binaryOperatorTypeKey(l, r Type) operatorTypeKey {
	return operatorTypeKey{briefType(l), briefType(r)}
}

func operandTypesSuffix(args []*operand) string {
	if len(args) == 0 {
		return "0"
	}
	var b strings.Builder
	for i, a := range args {
		if i > 0 {
			b.WriteByte('_')
		}
		b.WriteString(briefType(a.typ()))
	}
	return b.String()
}

func overloadSigKey(name string, sig *Signature) string {
	return name + "·" + overloadParamSuffix(sig)
}

type overloadResolveKey struct {
	cand  *Func
	ncand int
	args  string
}

func containsFunc(funcs []*Func, fn *Func) bool {
	for _, f := range funcs {
		if f == fn {
			return true
		}
	}
	return false
}

func buildOperatorTypeIndexes(cands map[string][]*Func) (map[string]map[operatorTypeKey]*Func, map[string]map[string]*Func) {
	binary := make(map[string]map[operatorTypeKey]*Func)
	unary := make(map[string]map[string]*Func)
	for name, funcs := range cands {
		var binIdx map[operatorTypeKey]*Func
		var unIdx map[string]*Func
		for _, fn := range funcs {
			if fn == nil || fn.typ == nil {
				continue
			}
			sig, ok := fn.typ.(*Signature)
			if !ok || sig.params == nil {
				continue
			}
			switch sig.params.Len() {
			case 1:
				if unIdx == nil {
					unIdx = make(map[string]*Func)
				}
				unIdx[briefType(sig.params.At(0).Type())] = fn
			case 2:
				if binIdx == nil {
					binIdx = make(map[operatorTypeKey]*Func)
				}
				k := binaryOperatorTypeKey(sig.params.At(0).Type(), sig.params.At(1).Type())
				binIdx[k] = fn
			}
		}
		if binIdx != nil {
			binary[name] = binIdx
		}
		if unIdx != nil {
			unary[name] = unIdx
		}
	}
	return binary, unary
}

func buildOverloadBySig(cands map[string][]*Func) map[string]*Func {
	out := make(map[string]*Func)
	for name, funcs := range cands {
		for _, fn := range funcs {
			if fn == nil || fn.typ == nil {
				continue
			}
			if sig, ok := fn.typ.(*Signature); ok {
				out[overloadSigKey(name, sig)] = fn
			}
		}
	}
	return out
}

func ensurePackageOperatorFuncIndex(pkg *Package) {
	if pkg == nil || pkg.operatorFuncIndex != nil {
		return
	}
	idx := make(map[string][]*Func)
	for name, funcs := range pkg.overloadFuncs {
		idx[name] = append([]*Func(nil), funcs...)
	}
	if pkg.scope != nil {
		for _, n := range pkg.scope.Names() {
			base := n
			if i := strings.Index(n, "·"); i >= 0 {
				base = n[:i]
			}
			obj := pkg.scope.Lookup(n)
			fn, ok := obj.(*Func)
			if !ok {
				continue
			}
			if containsFunc(idx[base], fn) {
				continue
			}
			idx[base] = append(idx[base], fn)
		}
	}
	pkg.operatorFuncIndex = idx
}

func ensurePackageOperatorTypeIndex(pkg *Package) {
	if pkg == nil {
		return
	}
	if pkg.operatorExact != nil {
		return
	}
	ensurePackageOperatorFuncIndex(pkg)
	pkg.operatorExact, pkg.operatorUnaryExact = buildOperatorTypeIndexes(pkg.operatorFuncIndex)
}

func ensurePackageExtensionIndex(pkg *Package) {
	if pkg == nil || pkg.extensionByName != nil {
		return
	}
	idx := make(map[string][]*Func)
	if pkg.scope != nil {
		for _, n := range pkg.scope.Names() {
			obj := pkg.scope.Lookup(n)
			fn, ok := obj.(*Func)
			if !ok {
				continue
			}
			if !fn.IsExtension() && !extensionFuncShape(pkg, fn) {
				continue
			}
			base := n
			if i := strings.Index(n, "·"); i >= 0 {
				base = n[:i]
			}
			if containsFunc(idx[base], fn) {
				continue
			}
			idx[base] = append(idx[base], fn)
		}
	}
	pkg.extensionByName = idx
}

func (check *Checker) indexExtensionFunc(fn *Func) {
	if fn == nil {
		return
	}
	if !fn.IsExtension() && !extensionFuncShape(check.pkg, fn) {
		return
	}
	if check.pkg.extensionByName == nil {
		check.pkg.extensionByName = make(map[string][]*Func)
	}
	name := fn.name
	if !containsFunc(check.pkg.extensionByName[name], fn) {
		check.pkg.extensionByName[name] = append(check.pkg.extensionByName[name], fn)
	}
}

func (check *Checker) buildCheckerIndexes() {
	check.operatorExact, check.operatorUnaryExact = buildOperatorTypeIndexes(check.overloadFuncs)
	check.overloadBySig = buildOverloadBySig(check.overloadFuncs)
	for _, funcs := range check.overloadMeths {
		for _, fn := range funcs {
			if fn == nil || fn.typ == nil {
				continue
			}
			if sig, ok := fn.typ.(*Signature); ok {
				check.overloadBySig[overloadSigKey(fn.name, sig)] = fn
			}
		}
	}
	if len(check.overloadFuncs) > 0 {
		check.pkg.operatorExact = check.operatorExact
		check.pkg.operatorUnaryExact = check.operatorUnaryExact
		check.pkg.operatorFuncIndex = make(map[string][]*Func, len(check.overloadFuncs))
		for name, funcs := range check.overloadFuncs {
			check.pkg.operatorFuncIndex[name] = append([]*Func(nil), funcs...)
		}
	}
}

func (check *Checker) lookupBinaryOperatorExact(name string, l, r *operand) *Func {
	lt, rt := l.typ(), r.typ()
	if isUntyped(lt) {
		lt = Default(lt)
	}
	if isUntyped(rt) {
		rt = Default(rt)
	}
	key := binaryOperatorTypeKey(lt, rt)

	for _, pkg := range check.operatorPackagesForOperands(l, r) {
		if pkg == nil {
			continue
		}
		ensurePackageOperatorTypeIndex(pkg)
		if idx := pkg.operatorExact[name]; idx != nil {
			if fn := idx[key]; fn != nil {
				return fn
			}
		}
	}
	if idx := check.operatorExact[name]; idx != nil {
		if fn := idx[key]; fn != nil {
			return fn
		}
	}
	return nil
}

func (check *Checker) lookupUnaryOperatorExact(name string, x *operand) *Func {
	typ := x.typ()
	if isUntyped(typ) {
		typ = Default(typ)
	}
	k := briefType(typ)

	for _, pkg := range check.operatorPackagesForOperands(x) {
		if pkg == nil {
			continue
		}
		ensurePackageOperatorTypeIndex(pkg)
		if idx := pkg.operatorUnaryExact[name]; idx != nil {
			if fn := idx[k]; fn != nil {
				return fn
			}
		}
	}
	if idx := check.operatorUnaryExact[name]; idx != nil {
		if fn := idx[k]; fn != nil {
			return fn
		}
	}
	return nil
}

func (check *Checker) lookupOverloadByArgTypes(name string, args []*operand) *Func {
	if check.overloadBySig == nil {
		return nil
	}
	key := name + "·" + operandTypesSuffix(args)
	if fn, ok := check.overloadBySig[key]; ok {
		return fn
	}
	// Retry with default types for untyped operands.
	defaulted := make([]*operand, len(args))
	for i, a := range args {
		if a == nil {
			continue
		}
		op := *a
		if isUntyped(op.typ()) {
			op.typ_ = Default(op.typ())
		}
		defaulted[i] = &op
	}
	key = name + "·" + operandTypesSuffix(defaulted)
	if fn, ok := check.overloadBySig[key]; ok {
		return fn
	}
	return nil
}

func operatorFuncsInPackage(pkg *Package, name string) []*Func {
	if pkg == nil {
		return nil
	}
	EnsurePackageOperatorIndexes(pkg)
	if pkg.overloadFuncs != nil {
		if cands := pkg.overloadFuncs[name]; len(cands) > 0 {
			return cands
		}
	}
	return pkg.operatorFuncIndex[name]
}

// EnsurePackageOperatorIndexes builds operator overload indexes for pkg,
// including packages loaded from export data that lack checker metadata.
func EnsurePackageOperatorIndexes(pkg *Package) {
	if pkg == nil {
		return
	}
	ensurePackageOperatorFuncIndex(pkg)
	if pkg.operatorExact != nil {
		return
	}
	pkg.operatorExact, pkg.operatorUnaryExact = buildOperatorTypeIndexes(pkg.operatorFuncIndex)
	if pkg.overloadFuncs == nil && len(pkg.operatorFuncIndex) > 0 {
		pkg.overloadFuncs = make(map[string][]*Func, len(pkg.operatorFuncIndex))
		for name, funcs := range pkg.operatorFuncIndex {
			pkg.overloadFuncs[name] = append([]*Func(nil), funcs...)
		}
	}
}

func extensionFuncsInPackage(pkg *Package, method string) []*Func {
	if pkg == nil {
		return nil
	}
	ensurePackageExtensionIndex(pkg)
	return pkg.extensionByName[method]
}
