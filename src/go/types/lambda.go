// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	. "internal/types/errors"
)

func (check *Checker) lambdaExpr(x *operand, e *ast.LambdaExpr, hint Type) {
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
	if hintSig.params.Len() != len(e.Params) {
		check.errorf(e, WrongArgCount, "lambda has %d parameters, want %d", len(e.Params), hintSig.params.Len())
		x.invalidate()
		return
	}
	if hintSig.results.Len() != 1 {
		check.errorf(e, InvalidSyntaxTree, "lambda expression requires exactly one result")
		x.invalidate()
		return
	}

	params := make([]*Var, len(e.Params))
	for i, id := range e.Params {
		params[i] = newVar(ParamVar, id.Pos(), check.pkg, id.Name, hintSig.params.At(i).typ)
	}
	var recvTP, typeTP []*TypeParam
	if r := hintSig.RecvTypeParams(); r != nil {
		recvTP = r.list()
	}
	if t := hintSig.TypeParams(); t != nil {
		typeTP = t.list()
	}
	sig := NewSignatureType(nil, recvTP, typeTP, NewTuple(params...), hintSig.results, hintSig.variadic)

	ret := new(ast.ReturnStmt)
	ret.Return = e.Body.Pos()
	ret.Results = []ast.Expr{e.Body}
	body := new(ast.BlockStmt)
	body.Lbrace = e.Lparen
	body.List = []ast.Stmt{ret}

	check.openScope(e, "function")
	sig.scope = check.scope
	check.scope.isFunc = true
	check.recordScope(e, check.scope)
	scopePos := e.Arrow
	for i, id := range e.Params {
		if id.Name != "" {
			check.declare(check.scope, id, params[i], scopePos)
		}
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
