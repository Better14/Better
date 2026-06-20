// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

func (check *Checker) lambdaExpr(x *operand, e *syntax.LambdaExpr, hint Type) {
	if hint == nil && check.Types != nil {
		if tv, ok := check.Types[e]; ok && tv.Type != nil {
			x.mode_ = value
			x.typ_ = tv.Type
			x.expr = e
			return
		}
	}
	if hint == nil {
		check.error(e, InvalidSyntaxTree, "lambda expression requires type context")
		x.invalidate()
		return
	}
	hintSig := check.signatureFromHint(hint)
	if hintSig == nil {
		check.error(e, InvalidSyntaxTree, "lambda expression requires function type context")
		x.invalidate()
		return
	}
	if hintSig.params.Len() != len(e.ParamList) {
		check.errorf(e, WrongArgCount, "lambda has %d parameters, want %d", len(e.ParamList), hintSig.params.Len())
		x.invalidate()
		return
	}
	if hintSig.results.Len() != 1 {
		check.error(e, InvalidSyntaxTree, "lambda expression requires exactly one result")
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

	check.openScope(e, "function")
	sigScope := check.scope
	sigScope.isFunc = true
	check.recordScope(e, sigScope)
	scopePos := syntax.EndPos(e)
	for i, f := range e.ParamList {
		if f.Name != nil && f.Name.Value != "" {
			check.declare(sigScope, f.Name, params[i], scopePos)
		}
	}

	resultType := hintSig.results.At(0).typ
	bodyChecked := false
	if _, ok := resultType.(*TypeParam); ok {
		var bodyVal operand
		check.expr(nil, &bodyVal, e.Body)
		if bodyVal.isValid() {
			resultType = bodyVal.typ()
			bodyChecked = true
		}
	}

	var recvTP, typeTP []*TypeParam
	if !bodyChecked {
		if r := hintSig.RecvTypeParams(); r != nil {
			recvTP = r.list()
		}
		if t := hintSig.TypeParams(); t != nil {
			typeTP = t.list()
		}
	}
	results := NewTuple(newVar(ResultVar, nopos, check.pkg, "", resultType))
	sig := NewSignatureType(nil, recvTP, typeTP, NewTuple(params...), results, hintSig.variadic)
	sig.scope = sigScope

	ret := new(syntax.ReturnStmt)
	ret.SetPos(e.Body.Pos())
	ret.Results = e.Body
	block := new(syntax.BlockStmt)
	block.SetPos(e.Pos())
	block.List = []syntax.Stmt{ret}

	if !check.conf.IgnoreFuncBodies && !bodyChecked {
		decl := check.decl
		iota := check.iota
		check.later(func() {
			check.funcBody(decl, "<lambda>", sig, block, iota)
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
