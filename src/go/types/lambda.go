
package types

import (
	"go/ast"
	. "internal/types/errors"
)

func (check *Checker) lambdaExpr(x *operand, e *ast.LambdaExpr, hint Type) {
	if hint == nil && check.Types != nil {
		if tv, ok := check.Types[e]; ok && tv.Type != nil {
			x.mode_ = value
			x.typ_ = tv.Type
			x.expr = e
			return
		}
	}
	if hint == nil {
		if check.inExtensionProbe || check.inOverloadProbe {
			x.invalidate()
			return
		}
		check.error(e, InvalidSyntaxTree, "lambda expression requires type context")
		x.invalidate()
		return
	}
	hintSig := check.signatureFromHint(hint)
	if hintSig == nil {
		if check.inExtensionProbe || check.inOverloadProbe {
			x.invalidate()
			return
		}
		check.error(e, InvalidSyntaxTree, "lambda expression requires function type context")
		x.invalidate()
		return
	}
	if hintSig.params.Len() != len(e.Params) {
		if check.inExtensionProbe || check.inOverloadProbe {
			x.invalidate()
			return
		}
		check.errorf(e, WrongArgCount, "lambda has %d parameters, want %d", len(e.Params), hintSig.params.Len())
		x.invalidate()
		return
	}
	if hintSig.results.Len() != 1 {
		check.error(e, InvalidSyntaxTree, "lambda expression requires exactly one result")
		x.invalidate()
		return
	}

	params := make([]*Var, len(e.Params))
	for i, id := range e.Params {
		params[i] = newVar(ParamVar, id.Pos(), check.pkg, id.Name, hintSig.params.At(i).typ)
	}

	check.openScope(e, "function")
	sigScope := check.scope
	sigScope.isFunc = true
	check.recordScope(e, sigScope)
	scopePos := e.Arrow
	for i, id := range e.Params {
		if id.Name != "" {
			check.declare(sigScope, id, params[i], scopePos)
		}
	}

	resultType := hintSig.results.At(0).typ
	bodyChecked := false
	if lambdaResultNeedsInference(resultType) {
		var bodyVal operand
		check.expr(nil, &bodyVal, e.Body)
		if bodyVal.isValid() {
			if inferred, ok := lambdaInferResultFromBody(resultType, bodyVal.typ()); ok {
				resultType = inferred
				bodyChecked = true
			}
		}
		if !bodyChecked {
			check.closeScope()
			x.invalidate()
			x.expr = e
			return
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

	ret := new(ast.ReturnStmt)
	ret.Return = e.Body.Pos()
	ret.Results = []ast.Expr{e.Body}
	body := new(ast.BlockStmt)
	body.Lbrace = e.Lparen
	body.List = []ast.Stmt{ret}

	if !check.conf.IgnoreFuncBodies && !bodyChecked {
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

// lambdaResultNeedsInference reports whether hint result type needs body
// typing to infer a concrete result (e.g. U, []U, iter.Seq[U]).
func lambdaResultNeedsInference(hintResult Type) bool {
	if !isValid(hintResult) {
		return false
	}
	hintResult = Unalias(hintResult)
	if _, ok := hintResult.(*TypeParam); ok {
		return true
	}
	if sl, ok := hintResult.Underlying().(*Slice); ok {
		if _, ok := Unalias(sl.elem).(*TypeParam); ok {
			return true
		}
	}
	if elem := iterSeqElem(hintResult); elem != nil {
		if _, ok := Unalias(elem).(*TypeParam); ok {
			return true
		}
	}
	return false
}

// lambdaInferResultFromBody maps a generic lambda result hint to a concrete
// type from the body expression, when possible.
func lambdaInferResultFromBody(hintResult, bodyType Type) (Type, bool) {
	if !isValid(hintResult) || !isValid(bodyType) {
		return nil, false
	}
	hintResult = Unalias(hintResult)
	bodyType = Unalias(bodyType)

	if _, ok := hintResult.(*TypeParam); ok {
		return bodyType, true
	}
	if sl, ok := hintResult.Underlying().(*Slice); ok {
		if _, ok := Unalias(sl.elem).(*TypeParam); ok {
			if _, ok := bodyType.Underlying().(*Slice); ok {
				return bodyType, true
			}
			return nil, false
		}
	}
	if elem := iterSeqElem(hintResult); elem != nil {
		if _, ok := Unalias(elem).(*TypeParam); ok {
			if iterSeqElem(bodyType) != nil {
				return bodyType, true
			}
			return nil, false
		}
	}
	return nil, false
}
