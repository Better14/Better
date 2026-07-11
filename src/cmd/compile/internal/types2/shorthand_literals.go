// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	. "internal/types/errors"
)

func (check *Checker) compositeShorthandLit(x *operand, e *syntax.CompositeLit, hint Type) {
	if hasSpreadElems(e.ElemList) {
		lowered := check.lowerSpreadShorthandLit(e)
		check.rawExpr(nil, x, lowered, hint, false)
		return
	}

	switch e.Shorthand {
	case syntax.ShorthandArray:
		elem := check.inferShorthandElemType(e.ElemList)
		e.Shorthand = syntax.ShorthandNone
		check.compositeLit(x, e, &Slice{elem: elem})

	case syntax.ShorthandMap:
		key, val := check.inferShorthandMapTypes(e.ElemList)
		key = check.defaultingType(key)
		val = check.defaultingType(val)
		e.Type = &syntax.MapType{
			Key:   check.typeToSyntaxExpr(key),
			Value: check.typeToSyntaxExpr(val),
		}
		e.Shorthand = syntax.ShorthandNone
		check.compositeLit(x, e, &Map{key: key, elem: val})

	case syntax.ShorthandSet:
		elem := check.inferShorthandElemType(e.ElemList)
		e.Shorthand = syntax.ShorthandNone
		check.expr(nil, x, check.makeSetOfCall(e.Pos(), e.ElemList, elem))
		return

	default:
		check.compositeLit(x, e, hint)
	}
}

func hasSpreadElems(elems []syntax.Expr) bool {
	for _, el := range elems {
		if _, ok := el.(*syntax.SpreadExpr); ok {
			return true
		}
	}
	return false
}

func (check *Checker) inferShorthandElemType(elems []syntax.Expr) Type {
	var typ Type
	for _, e := range elems {
		if _, ok := e.(*syntax.SpreadExpr); ok {
			continue
		}
		var x operand
		check.genericExpr(&x, e, nil)
		if !x.isValid() {
			continue
		}
		if typ == nil {
			typ = x.typ()
		} else {
			typ = check.commonShorthandType(typ, x.typ())
		}
	}
	if typ == nil {
		return Typ[Int]
	}
	if isUntyped(typ) {
		return Default(typ)
	}
	return typ
}

func (check *Checker) inferShorthandMapTypes(elems []syntax.Expr) (key, val Type) {
	for _, e := range elems {
		kv, ok := e.(*syntax.KeyValueExpr)
		if !ok {
			continue
		}
		var kx, vx operand
		check.genericExpr(&kx, kv.Key, nil)
		check.genericExpr(&vx, kv.Value, nil)
		if kx.isValid() {
			if key == nil {
				key = kx.typ()
			} else {
				key = check.commonShorthandType(key, kx.typ())
			}
		}
		if vx.isValid() {
			if val == nil {
				val = vx.typ()
			} else {
				val = check.commonShorthandType(val, vx.typ())
			}
		}
	}
	if key == nil {
		key = Typ[Invalid]
	}
	if val == nil {
		val = Typ[Invalid]
	}
	return key, val
}

func (check *Checker) commonShorthandType(a, b Type) Type {
	if Identical(a, b) {
		return a
	}
	if isUntyped(b) && !isUntyped(a) {
		return a
	}
	if isUntyped(a) && !isUntyped(b) {
		return b
	}
	return a
}

func (check *Checker) setTypeForElem(pos syntax.Pos, elem Type) Type {
	pkg := check.importPackage(pos, "set", "")
	if pkg == nil {
		return Typ[Invalid]
	}
	obj := pkg.Scope().Lookup("Set")
	if obj == nil {
		check.errorf(pos, BrokenImport, "set package missing type Set")
		return Typ[Invalid]
	}
	inst, err := Instantiate(check.ctxt, obj.Type(), []Type{elem}, true)
	if err != nil {
		check.errorf(pos, BrokenImport, "invalid set type: %v", err)
		return Typ[Invalid]
	}
	return inst
}

func (check *Checker) makeSetOfCall(pos syntax.Pos, elems []syntax.Expr, elem Type) syntax.Expr {
	for _, el := range elems {
		if se, ok := el.(*syntax.SpreadExpr); ok {
			var sx operand
			check.expr(nil, &sx, se.X)
			setTyp := check.setTypeForElem(pos, elem)
			if sx.isValid() && !Identical(sx.typ(), setTyp) {
				check.errorf(se, InvalidLitField, "cannot spread %s into set literal", sx.typ())
			}
			continue
		}
		var ex operand
		check.genericExpr(&ex, el, elem)
		check.assignment(&ex, elem, "set literal")
	}
	sel := &syntax.SelectorExpr{
		X:   syntax.NewName(pos, "set"),
		Sel: syntax.NewName(pos, "Of"),
	}
	sel.SetPos(pos)
	call := &syntax.CallExpr{Fun: sel, ArgList: elems}
	call.SetPos(pos)
	return call
}

func (check *Checker) lowerSpreadShorthandLit(e *syntax.CompositeLit) syntax.Expr {
	switch e.Shorthand {
	case syntax.ShorthandArray:
		elem := check.inferShorthandElemType(e.ElemList)
		return check.lowerSpreadSliceLit(e, elem)
	case syntax.ShorthandMap:
		key, val := check.inferShorthandMapTypes(e.ElemList)
		return check.lowerSpreadMapLit(e, key, val)
	case syntax.ShorthandSet:
		elem := check.inferShorthandElemType(e.ElemList)
		return check.lowerSpreadSetLit(e, elem)
	default:
		if st, ok := syntax.Unparen(e.Type).(*syntax.SetType); ok {
			elem := check.varType(st.Elem)
			return check.lowerSpreadSetLit(e, elem)
		}
		return e
	}
}

func (check *Checker) lowerSpreadSliceLit(e *syntax.CompositeLit, elem Type) syntax.Expr {
	var result syntax.Expr
	flush := func(batch []syntax.Expr) {
		if len(batch) == 0 {
			return
		}
		cl := &syntax.CompositeLit{ElemList: batch, Shorthand: syntax.ShorthandArray}
		cl.SetPos(e.Pos())
		if result == nil {
			result = cl
		} else {
			result = check.makeAppendCall(result, cl, false)
		}
	}
	var batch []syntax.Expr
	for _, el := range e.ElemList {
		if se, ok := el.(*syntax.SpreadExpr); ok {
			flush(batch)
			batch = nil
			if result == nil {
				result = se.X
			} else {
				result = check.makeAppendCall(result, se.X, true)
			}
			continue
		}
		batch = append(batch, el)
	}
	flush(batch)
	if result == nil {
		cl := &syntax.CompositeLit{Shorthand: syntax.ShorthandArray}
		cl.SetPos(e.Pos())
		return cl
	}
	return result
}

func (check *Checker) lowerSpreadMapLit(e *syntax.CompositeLit, key, val Type) syntax.Expr {
	mapTyp := &Map{key: key, elem: val}
	var static []syntax.Expr
	var spreads []syntax.Expr
	for _, el := range e.ElemList {
		if se, ok := el.(*syntax.SpreadExpr); ok {
			spreads = append(spreads, se.X)
			continue
		}
		static = append(static, el)
	}
	init := &syntax.CompositeLit{
		Type: &syntax.MapType{
			Key:   check.typeToSyntaxExpr(key),
			Value: check.typeToSyntaxExpr(val),
		},
		ElemList: static,
	}
	init.SetPos(e.Pos())

	out := syntax.NewName(e.Pos(), "out")
	stmts := []syntax.Stmt{&syntax.AssignStmt{
		Op:  syntax.Def,
		Lhs: out,
		Rhs: init,
	}}
	for _, sp := range spreads {
		var sx operand
		check.expr(nil, &sx, sp)
		if !sx.isValid() || !Identical(sx.typ(), mapTyp) {
			check.errorf(sp, InvalidLitField, "cannot spread %s into map literal", sx.typ())
			continue
		}
		k := syntax.NewName(e.Pos(), "k")
		v := syntax.NewName(e.Pos(), "v")
		rng := &syntax.RangeClause{
			Lhs: &syntax.ListExpr{ElemList: []syntax.Expr{k, v}},
			Def: true,
			X:   sp,
		}
		rng.SetPos(e.Pos())
		stmts = append(stmts, &syntax.ForStmt{
			Init: rng,
			Body: &syntax.BlockStmt{List: []syntax.Stmt{&syntax.AssignStmt{
				Op:  0,
				Lhs: &syntax.IndexExpr{X: out, Index: k},
				Rhs: v,
			}}},
		})
	}
	body := &syntax.BlockStmt{List: stmts}
	body.SetPos(e.Pos())
	body.List = append(body.List, &syntax.ReturnStmt{Results: out})
	fl := &syntax.FuncLit{
		Type: &syntax.FuncType{ResultList: []*syntax.Field{{Type: &syntax.MapType{
			Key:   check.typeToSyntaxExpr(key),
			Value: check.typeToSyntaxExpr(val),
		}}}},
		Body: body,
	}
	fl.SetPos(e.Pos())
	call := &syntax.CallExpr{Fun: fl}
	call.SetPos(e.Pos())
	return call
}

func (check *Checker) defaultingType(t Type) Type {
	if isUntyped(t) {
		return Default(t)
	}
	return t
}

func (check *Checker) typeToSyntaxExpr(t Type) syntax.Expr {
	if t == nil {
		return syntax.NewName(nopos, "invalid")
	}
	if b, ok := t.Underlying().(*Basic); ok {
		return syntax.NewName(nopos, b.name)
	}
	return syntax.NewName(nopos, "invalid")
}

func (check *Checker) lowerSpreadSetLit(e *syntax.CompositeLit, elem Type) syntax.Expr {
	setTyp := check.setTypeForElem(e.Pos(), elem)
	var parts []syntax.Expr
	var batch []syntax.Expr
	flush := func() {
		if len(batch) == 0 {
			return
		}
		parts = append(parts, check.makeSetOfCall(e.Pos(), batch, elem))
		batch = nil
	}
	for _, el := range e.ElemList {
		if se, ok := el.(*syntax.SpreadExpr); ok {
			flush()
			var sx operand
			check.expr(nil, &sx, se.X)
			if !sx.isValid() || !AssignableTo(sx.typ(), setTyp) {
				check.errorf(se, InvalidLitField, "cannot spread %s into set literal", sx.typ())
				continue
			}
			parts = append(parts, se.X)
			continue
		}
		batch = append(batch, el)
	}
	flush()
	if len(parts) == 0 {
		return check.makeSetOfCall(e.Pos(), nil, elem)
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result = check.makeSetUnionCall(e.Pos(), result, parts[i])
	}
	return result
}

func (check *Checker) makeAppendCall(slice, arg syntax.Expr, spread bool) *syntax.CallExpr {
	call := &syntax.CallExpr{
		Fun:     syntax.NewName(slice.Pos(), "append"),
		ArgList: []syntax.Expr{slice, arg},
		HasDots: spread,
	}
	call.SetPos(slice.Pos())
	return call
}

func (check *Checker) makeSetUnionCall(pos syntax.Pos, a, b syntax.Expr) syntax.Expr {
	sel := &syntax.SelectorExpr{
		X:   a,
		Sel: syntax.NewName(pos, "Union"),
	}
	sel.SetPos(pos)
	call := &syntax.CallExpr{Fun: sel, ArgList: []syntax.Expr{b}}
	call.SetPos(pos)
	return call
}
