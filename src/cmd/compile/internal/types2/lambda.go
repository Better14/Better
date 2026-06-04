// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

func (check *Checker) lambdaExpr(x *operand, e *syntax.LambdaExpr, hint Type) {
	if hint == nil {
		check.errorf(e, InvalidSyntaxTree, "lambda expression requires type context")
		x.invalidate()
		return
	}
	hintSig := check.signatureFromHint(hint)
	if hintSig == nil {
		check.errorf(e, InvalidSyntaxTree, "lambda expression requires function type context")
		x.invalidate()
		return
	}
	if hintSig.params.Len() != len(e.ParamList) {
		check.errorf(e, WrongArgCount, "lambda has %d parameters, want %d", len(e.ParamList), hintSig.params.Len())
		x.invalidate()
		return
	}
	if hintSig.results.Len() != 1 {
		check.errorf(e, InvalidSyntaxTree, "lambda expression requires exactly one result")
		x.invalidate()
		return
	}

	params := make([]*Var, len(e.ParamList))
	for i, f := range e.ParamList {
		name := ""
		if f.Name != nil {
			name = f.Name.Value
		}
		params[i] = newVar(ParamVar, f.Pos(), check.pkg, name, hintSig.params.At(i).typ)
	}
	var recvTP, typeTP []*TypeParam
	if r := hintSig.RecvTypeParams(); r != nil {
		recvTP = r.list()
	}
	if t := hintSig.TypeParams(); t != nil {
		typeTP = t.list()
	}
	sig := NewSignatureType(nil, recvTP, typeTP, NewTuple(params...), hintSig.results, hintSig.variadic)

	ret := new(syntax.ReturnStmt)
	ret.SetPos(e.Body.Pos())
	ret.Results = e.Body
	body := new(syntax.BlockStmt)
	body.SetPos(e.Pos())
	body.List = []syntax.Stmt{ret}

	check.openScope(e, "function")
	sig.scope = check.scope
	check.scope.isFunc = true
	check.recordScope(e, check.scope)
	scopePos := syntax.EndPos(e)
	for i, f := range e.ParamList {
		if f.Name != nil && f.Name.Value != "" {
			check.declare(check.scope, f.Name, params[i], scopePos)
		}
	}
	if r := hintSig.results.vars[0]; r.name != "" {
		// result names are not in lambda syntax
	}
	if !check.conf.IgnoreFuncBodies {
		decl := check.decl
		iota := check.iota
		check.later(func() {
			check.funcBody(decl, "<lambda>", sig, body, iota)
		}).describef(e, "lambda")
	}
	check.closeScope()

	x.mode_ = value
	x.typ_ = sig
	x.expr = e
	check.recordTypeAndValue(e, value, sig, nil)
}

func (check *Checker) signatureFromHint(hint Type) *Signature {
	for {
		switch t := hint.(type) {
		case *Signature:
			return t
		case *Named:
			hint = t.Underlying()
		default:
			return nil
		}
	}
}
