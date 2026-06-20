// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmp"
	"cmd/compile/internal/syntax"
	"fmt"
	. "internal/types/errors"
	"slices"
	"strings"
)

type methodKey struct {
	recvName string
	name     string
}

func sameParamSignature(a, b *Func) bool {
	if a == nil || b == nil || a.typ == nil || b.typ == nil {
		return false
	}
	sa := a.typ.(*Signature)
	sb := b.typ.(*Signature)
	if sa.variadic != sb.variadic {
		return false
	}
	pa := sa.Params()
	pb := sb.Params()
	if pa == nil || pb == nil {
		return pa == pb
	}
	if pa.Len() != pb.Len() {
		return false
	}
	for i := 0; i < pa.Len(); i++ {
		if !Identical(pa.At(i).Type(), pb.At(i).Type()) {
			return false
		}
	}
	return true
}

func briefType(t Type) string {
	t = Unalias(t)
	if n, ok := t.(*Named); ok && n.obj != nil {
		if n.obj.Exported() {
			return n.obj.name
		}
		if n.obj.pkg != nil {
			return n.obj.pkg.name + "." + n.obj.name
		}
		return n.obj.name
	}
	switch u := t.Underlying().(type) {
	case *Basic:
		return u.name
	case *Named:
		if u.obj != nil {
			if u.obj.Exported() {
				return u.obj.name
			}
			if u.obj.pkg != nil {
				return u.obj.pkg.name + "." + u.obj.name
			}
			return u.obj.name
		}
	case *Pointer:
		return "p" + briefType(u.base)
	case *Slice:
		return "s" + briefType(u.elem)
	case *Array:
		return fmt.Sprintf("a%d_%s", u.len, briefType(u.elem))
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', ',', '(', ')', '[', ']', '*', ';':
			return '_'
		default:
			return r
		}
	}, t.String())
}

func overloadParamSuffix(sig *Signature) string {
	params := sig.Params()
	if params == nil || params.Len() == 0 {
		if sig.Variadic() {
			return "v"
		}
		return "0"
	}
	var b strings.Builder
	for i := 0; i < params.Len(); i++ {
		if i > 0 {
			b.WriteByte('_')
		}
		b.WriteString(briefType(params.At(i).Type()))
	}
	if sig.Variadic() {
		b.WriteString("_dotdotdot")
	}
	return b.String()
}

func (check *Checker) overloadList(funcs []*Func) string {
	var b strings.Builder
	for i, fn := range funcs {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(check.funcString(fn, false))
	}
	return b.String()
}

func (check *Checker) identicalMethodSig(a, b *Func) bool {
	if a == nil || b == nil || a.typ == nil || b.typ == nil {
		return false
	}
	sa, ok := a.typ.(*Signature)
	if !ok {
		return false
	}
	sb, ok := b.typ.(*Signature)
	if !ok {
		return false
	}
	if sa.recv == nil || sb.recv == nil {
		return sa.recv == sb.recv
	}
	if !check.identicalReceiverTypes(sa.recv.typ, sb.recv.typ) {
		return false
	}
	if sa.variadic != sb.variadic {
		return false
	}
	pa, pb := sa.Params(), sb.Params()
	if pa == nil || pb == nil {
		if pa != pb {
			return false
		}
	} else {
		if pa.Len() != pb.Len() {
			return false
		}
		for i := 0; i < pa.Len(); i++ {
			if !Identical(pa.At(i).Type(), pb.At(i).Type()) {
				return false
			}
		}
	}
	ra, rb := sa.Results(), sb.Results()
	if ra == nil || rb == nil {
		return ra == rb
	}
	if ra.Len() != rb.Len() {
		return false
	}
	for i := 0; i < ra.Len(); i++ {
		if !Identical(ra.At(i).Type(), rb.At(i).Type()) {
			return false
		}
	}
	return true
}

func (check *Checker) identicalReceiverTypes(a, b Type) bool {
	if Identical(a, b) {
		return true
	}
	baseA := check.receiverNamedBase(a)
	baseB := check.receiverNamedBase(b)
	return baseA != nil && baseA == baseB
}

func (check *Checker) receiverNamedBase(t Type) *TypeName {
	t, _ = deref(t)
	n := asNamed(Unalias(t))
	if n == nil || n.obj == nil || n.obj.pkg != check.pkg {
		return nil
	}
	_, base := check.resolveBaseTypeName(false, syntax.NewName(nopos, n.obj.Name()))
	return base
}

func (check *Checker) recvBaseTypeName(recvName string) string {
	_, base := check.resolveBaseTypeName(false, syntax.NewName(nopos, recvName))
	if base != nil {
		return base.Name()
	}
	return recvName
}

func (check *Checker) checkOverloadDuplicates(name string, cands []*Func, kind string) {
	seen := make(map[string]*Func)
	for _, fn := range cands {
		if fn == nil || fn.typ == nil {
			continue
		}
		sig, ok := fn.typ.(*Signature)
		if !ok {
			continue
		}
		key := overloadSigKey(name, sig)
		if prev, ok := seen[key]; ok {
			check.errorf(fn.pos, DuplicateDecl, "%s redeclared in this block", name)
			_ = prev
			continue
		}
		seen[key] = fn
	}
}

func (check *Checker) resolveRecvBaseName(recvName string) string {
	_, base := check.resolveBaseTypeName(false, syntax.NewName(nopos, recvName))
	if base != nil {
		return base.Name()
	}
	return recvName
}

func (check *Checker) checkMethodOverloadDuplicates() {
	merged := make(map[methodKey][]*Func)
	for key, cands := range check.overloadMeths {
		mk := methodKey{recvName: check.recvBaseTypeName(key.recvName), name: key.name}
		merged[mk] = append(merged[mk], cands...)
	}
	for key, cands := range merged {
		slices.SortFunc(cands, func(a, b *Func) int {
			return cmp.Compare(a.order(), b.order())
		})
		var seen []*Func
		for _, fn := range cands {
			if fn == nil || fn.typ == nil {
				continue
			}
			sig, ok := fn.typ.(*Signature)
			if !ok || sig.recv == nil || !isValid(sig.recv.typ) {
				continue
			}
			for _, prev := range seen {
				if check.identicalMethodSig(prev, fn) {
					if prev.Pos().IsKnown() {
						check.errorf(fn.pos, DuplicateMethod, "method %s.%s already declared at %v", key.recvName, key.name, prev.Pos())
					} else {
						check.errorf(fn.pos, DuplicateMethod, "method %s.%s already declared", key.recvName, key.name)
					}
					goto next
				}
			}
			seen = append(seen, fn)
		next:
		}
	}
}

func (check *Checker) assignOverloadSuffixes() {
	check.validateOperatorPairs()
	for name, cands := range check.overloadFuncs {
		check.checkOverloadDuplicates(name, cands, "function")
	}
	check.checkMethodOverloadDuplicates()

	check.buildCheckerIndexes()

	for _, cands := range check.overloadFuncs {
		if len(cands) <= 1 {
			continue
		}
		for _, fn := range cands {
			if sig, _ := fn.typ.(*Signature); sig != nil {
				fn.setLinkSuffix(overloadParamSuffix(sig))
			}
		}
	}
	for _, cands := range check.overloadMeths {
		if len(cands) <= 1 {
			continue
		}
		for _, fn := range cands {
			if sig, _ := fn.typ.(*Signature); sig != nil {
				fn.setLinkSuffix(overloadParamSuffix(sig))
			}
		}
	}
}

func (check *Checker) hasCallOverloads() bool {
	for _, cands := range check.overloadFuncs {
		if len(cands) > 1 {
			return true
		}
	}
	for _, cands := range check.overloadMeths {
		if len(cands) > 1 {
			return true
		}
	}
	return false
}

func (check *Checker) recvBaseNameFromExpr(x syntax.Expr) string {
	var recv operand
	check.rawExpr(nil, &recv, x, nil, true)
	if !recv.isValid() {
		return ""
	}
	t := recv.typ()
	if p, _ := t.Underlying().(*Pointer); p != nil {
		t = p.base
	}
	if n := asNamed(t); n != nil && n.obj != nil {
		return n.obj.name
	}
	return ""
}

func recvBaseNameFromType(t Type) string {
	if p, _ := t.Underlying().(*Pointer); p != nil {
		t = p.base
	}
	if n := asNamed(t); n != nil && n.obj != nil {
		return n.obj.name
	}
	return ""
}

func methodIndexInNamed(recv Type, fn *Func) int {
	if p, _ := recv.Underlying().(*Pointer); p != nil {
		recv = p.base
	}
	n := asNamed(recv)
	if n == nil {
		return -1
	}
	for i := 0; i < n.NumMethods(); i++ {
		if n.Method(i) == fn {
			return i
		}
	}
	return -1
}

func (check *Checker) overloadCandidatesForCall(call *syntax.CallExpr) []*Func {
	if !check.hasCallOverloads() {
		return nil
	}
	switch fun := call.Fun.(type) {
	case *syntax.Name:
		return check.overloadFuncs[fun.Value]
	case *syntax.SelectorExpr:
		if id, ok := fun.X.(*syntax.Name); ok {
			if obj := check.lookup(id.Value); obj != nil {
				if _, isPkg := obj.(*PkgName); isPkg {
					return nil
				}
			}
		}
		recvName := check.recvBaseNameFromExpr(fun.X)
		if recvName == "" {
			return nil
		}
		return check.overloadMeths[methodKey{recvName: recvName, name: fun.Sel.Value}]
	default:
		return nil
	}
}
