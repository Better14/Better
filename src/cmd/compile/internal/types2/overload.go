package types2

import "cmd/compile/internal/syntax"

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

func (check *Checker) overloadCandidatesForCall(call *syntax.CallExpr) []*Func {
	switch fun := call.Fun.(type) {
	case *syntax.Name:
		return check.overloadFuncs[fun.Value]
	case *syntax.SelectorExpr:
		if id, ok := fun.X.(*syntax.Name); ok {
			if obj := check.lookup(id.Value); obj != nil {
				if _, isPkg := obj.(*PkgName); isPkg {
					// package selector (pkg.F) is not handled here
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

