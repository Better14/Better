// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"go/constant"
	"internal/pkgbits"
)

// enumCaseBinds maps pattern binding names to payload slot indices.
func enumCaseBinds(p *pkgWriter, enumTyp *types2.Enum, cas syntax.Expr) map[string]int {
	binds := make(map[string]int)
	switch cas := cas.(type) {
	case *syntax.CallExpr:
		name := cas.Fun.(*syntax.Name)
		obj := p.info.Uses[name]
		v := enumVariantObj(enumTyp, obj)
		for i, arg := range cas.ArgList {
			n, ok := arg.(*syntax.Name)
			if !ok || n.Value == "_" {
				continue
			}
			binds[n.Value] = i
		}
		_ = v

	case *syntax.EnumPattern:
		obj := p.info.Uses[cas.Variant]
		v := enumVariantObj(enumTyp, obj)
		for i, f := range v.Fields() {
			if f == nil {
				continue
			}
			for _, pf := range cas.Fields {
				if pf != nil && pf.Name != nil && pf.Name.Value == f.Name() && pf.Name.Value != "_" {
					binds[pf.Name.Value] = i
					break
				}
			}
		}
	}
	return binds
}

// enumVariantObj returns the enum variant described by obj, if any.
func enumVariantObj(enumTyp *types2.Enum, obj types2.Object) *types2.EnumVariant {
	if enumTyp == nil || obj == nil {
		return nil
	}
	for _, v := range enumTyp.Variants() {
		if v.Obj() == obj {
			return v
		}
	}
	return nil
}

// enumTypeOf returns the enum type for typ, if any.
func enumTypeOf(typ types2.Type) (*types2.Enum, bool) {
	return types2.AsEnum(typ)
}

// enumVariantFromCase returns the variant for an enum switch case label.
func enumVariantFromCase(p *pkgWriter, enumTyp *types2.Enum, cas syntax.Expr) (*types2.EnumVariant, bool) {
	switch cas := cas.(type) {
	case *syntax.Name:
		obj := p.info.Uses[cas]
		if obj == nil {
			return nil, false
		}
		v := enumVariantObj(enumTyp, obj)
		if v == nil || len(v.Tuple()) > 0 || len(v.Fields()) > 0 {
			return nil, false
		}
		return v, true

	case *syntax.CallExpr:
		name, ok := cas.Fun.(*syntax.Name)
		if !ok {
			return nil, false
		}
		obj := p.info.Uses[name]
		if obj == nil {
			return nil, false
		}
		v := enumVariantObj(enumTyp, obj)
		if v == nil || len(v.Tuple()) == 0 {
			return nil, false
		}
		return v, true

	case *syntax.EnumPattern:
		if cas.Variant == nil {
			return nil, false
		}
		obj := p.info.Uses[cas.Variant]
		if obj == nil {
			return nil, false
		}
		v := enumVariantObj(enumTyp, obj)
		if v == nil || len(v.Fields()) == 0 {
			return nil, false
		}
		return v, true

	default:
		return nil, false
	}
}

// writeConstExpr writes a typed constant expression.
func (w *writer) writeConstExpr(pos syntax.Pos, typ types2.Type, val constant.Value) {
	w.Code(exprConst)
	w.pos(pos)
	w.typ(typ)
	w.Value(val)
}

type enumFieldInit struct {
	index int
	expr  syntax.Expr
	val   constant.Value
}

// writeEnumStructLit writes the payload of an enum composite literal.
// The caller must emit exprCompLit first, except when invoked from compLit
// (which already wrote the opcode).
func (w *writer) writeEnumStructLit(expr syntax.Expr, enumType types2.Type, inits []enumFieldInit) {
	w.Sync(pkgbits.SyncCompLit)
	w.pos(expr)
	w.typ(enumType)

	structTyp := types2.CoreType(enumType).(*types2.Struct)
	if w.Version().Has(pkgbits.CompactCompLiterals) {
		w.Int(-len(inits))
		for _, init := range inits {
			w.pos(expr)
			w.Int(init.index)
			if init.val != nil {
				w.writeConstExpr(expr.Pos(), structTyp.Field(init.index).Type(), init.val)
			} else {
				w.implicitConvExpr(structTyp.Field(init.index).Type(), init.expr)
			}
		}
		return
	}

	w.p.fatalf(expr, "enum composite literal requires compact format")
}

func (w *writer) writeEnumUnitVariant(expr syntax.Expr, enumType types2.Type, tag int64) {
	w.Code(exprCompLit)
	w.writeEnumStructLit(expr, enumType, []enumFieldInit{{
		index: 0,
		val:   constant.MakeInt64(tag),
	}})
}

func enumTypeFromSelectorX(p *pkgWriter, x syntax.Expr) types2.Type {
	x = syntax.Unparen(x)
	if tv := x.GetTypeInfo(); tv.IsType() {
		return tv.Type
	}
	if name, ok := x.(*syntax.Name); ok {
		if obj, ok := p.info.Defs[name].(*types2.TypeName); ok {
			return obj.Type()
		}
		if obj, ok := p.info.Uses[name].(*types2.TypeName); ok {
			return obj.Type()
		}
	}
	if tv, ok := p.maybeTypeAndValue(x); ok && tv.IsValue() {
		if _, ok := enumTypeOf(tv.Type); ok {
			return tv.Type
		}
	}
	return nil
}

func enumObjFromSel(p *pkgWriter, enumTyp *types2.Enum, sel *syntax.Name) types2.Object {
	if obj := p.info.Uses[sel]; obj != nil {
		return obj
	}
	if enumTyp != nil && enumTyp.Scope() != nil {
		return enumTyp.Scope().Lookup(sel.Value)
	}
	return nil
}

func (w *writer) enumVariantFromExpr(expr syntax.Expr) (*types2.Enum, *types2.EnumVariant, types2.Object) {
	switch expr := syntax.Unparen(expr).(type) {
	case *syntax.Name:
		obj := w.p.info.Uses[expr]
		if obj == nil {
			return nil, nil, nil
		}
		if typ := obj.Type(); typ != nil {
			if enumTyp, ok := enumTypeOf(typ); ok {
				if v := enumVariantObj(enumTyp, obj); v != nil {
					return enumTyp, v, obj
				}
			}
		}
		return nil, nil, obj

	case *syntax.SelectorExpr:
		if tv, ok := w.p.maybeTypeAndValue(expr); ok {
			if enumTyp, ok := enumTypeOf(tv.Type); ok {
				obj := enumObjFromSel(w.p, enumTyp, expr.Sel)
				if obj == nil {
					return nil, nil, nil
				}
				if v := enumVariantObj(enumTyp, obj); v != nil {
					return enumTyp, v, obj
				}
			}
		}
		if enumTyp, ok := enumTypeOf(enumTypeFromSelectorX(w.p, expr.X)); ok {
			obj := enumObjFromSel(w.p, enumTyp, expr.Sel)
			if obj == nil {
				return nil, nil, nil
			}
			if v := enumVariantObj(enumTyp, obj); v != nil {
				return enumTyp, v, obj
			}
		}
		return nil, nil, nil

	default:
		return nil, nil, nil
	}
}

func (w *writer) tryWriteEnumName(expr *syntax.Name) bool {
	enumTyp, variant, obj := w.enumVariantFromExpr(expr)
	if variant == nil || obj == nil || enumTyp == nil {
		return false
	}
	if len(variant.Tuple()) > 0 || len(variant.Fields()) > 0 {
		return false
	}
	w.writeEnumUnitVariant(expr, obj.Type(), variant.Tag())
	return true
}

func (w *writer) tryWriteEnumSelector(expr *syntax.SelectorExpr) bool {
	enumType := enumTypeFromSelectorX(w.p, expr.X)
	enumTyp, ok := enumTypeOf(enumType)
	if !ok {
		if tv, ok := w.p.maybeTypeAndValue(expr); ok {
			enumTyp, ok = enumTypeOf(tv.Type)
			enumType = tv.Type
		}
		if !ok {
			return false
		}
	}

	selName := expr.Sel.Value
	for _, v := range enumTyp.Variants() {
		if v.Name() != selName {
			continue
		}
		switch v.Obj().(type) {
		case *types2.Const:
			if len(v.Tuple()) > 0 || len(v.Fields()) > 0 {
				return false
			}
			w.writeEnumUnitVariant(expr, enumType, v.Tag())
			return true

		case *types2.Func:
			w.Code(exprGlobal)
			w.obj(v.Obj(), nil)
			return true
		}
		return false
	}
	return false
}

func enumVariantByObj(enumTyp *types2.Enum, obj types2.Object) *types2.EnumVariant {
	if enumTyp == nil || obj == nil {
		return nil
	}
	for _, v := range enumTyp.Variants() {
		if v.Obj() == obj {
			return v
		}
	}
	return nil
}

func (w *writer) tryWriteEnumVariantCall(expr *syntax.CallExpr) bool {
	tv := w.p.typeAndValue(expr)
	enumTyp, ok := enumTypeOf(tv.Type)
	if !ok {
		return false
	}

	var variant *types2.EnumVariant
	switch f := syntax.Unparen(expr.Fun).(type) {
	case *syntax.Name:
		if obj := w.p.info.Uses[f]; obj != nil {
			variant = enumVariantByObj(enumTyp, obj)
		}
	case *syntax.SelectorExpr:
		if _, v, obj := w.enumVariantFromExpr(f); v != nil {
			variant = v
			_ = obj
		}
	}
	if variant == nil {
		return false
	}

	var inits []enumFieldInit
	inits = append(inits, enumFieldInit{index: 0, val: constant.MakeInt64(variant.Tag())})

	switch {
	case len(variant.Tuple()) > 0:
		for i, arg := range expr.ArgList {
			inits = append(inits, enumFieldInit{index: i + 1, expr: arg})
		}

	case len(variant.Fields()) == 0 && len(expr.ArgList) == 0:
		// unit variant call

	default:
		return false
	}

	w.Code(exprCompLit)
	w.writeEnumStructLit(expr, tv.Type, inits)
	return true
}

func (w *writer) tryWriteEnumCompositeLit(lit *syntax.CompositeLit) bool {
	if lit.Type == nil {
		return false
	}
	sel, ok := syntax.Unparen(lit.Type).(*syntax.SelectorExpr)
	if !ok {
		return false
	}

	tv := w.p.typeAndValue(lit)
	enumTyp, ok := enumTypeOf(tv.Type)
	if !ok {
		return false
	}

	obj := w.p.info.Uses[sel.Sel]
	if obj == nil {
		return false
	}
	variant := enumVariantObj(enumTyp, obj)
	if variant == nil || len(variant.Fields()) == 0 {
		return false
	}

	var inits []enumFieldInit
	inits = append(inits, enumFieldInit{index: 0, val: constant.MakeInt64(variant.Tag())})

	visited := make(map[string]bool)
	for _, el := range lit.ElemList {
		kv, ok := el.(*syntax.KeyValueExpr)
		if !ok {
			continue
		}
		key := kv.Key.(*syntax.Name)
		if visited[key.Value] {
			continue
		}
		visited[key.Value] = true
		for i, f := range variant.Fields() {
			if f.Name() == key.Value {
				inits = append(inits, enumFieldInit{index: i + 1, expr: kv.Value})
				break
			}
		}
	}

	w.writeEnumStructLit(lit, tv.Type, inits)
	return true
}

func (w *writer) writeEnumFieldRef(base syntax.Expr, fieldIndex int) {
	typ := w.p.typeOf(base)
	structTyp := types2.CoreType(typ).(*types2.Struct)

	w.Code(exprFieldVal)
	w.expr(base)
	w.pos(base)
	w.selector(structTyp.Field(fieldIndex))
}

func (w *writer) writeEnumPayloadExpr(pos syntax.Pos, tag syntax.Expr, payloadIndex int, dstTyp types2.Type) {
	structTyp := types2.CoreType(w.p.typeOf(tag)).(*types2.Struct)
	srcTyp := structTyp.Field(payloadIndex + 1).Type()

	writeField := func() {
		w.Code(exprFieldVal)
		w.expr(tag)
		w.pos(pos)
		w.selector(structTyp.Field(payloadIndex + 1))
	}

	if types2.IsInterface(srcTyp) && !types2.IsInterface(dstTyp) {
		w.Code(exprAssert)
		writeField()
		w.pos(pos)
		w.Sync(pkgbits.SyncExprType)
		w.pos(pos)
		w.Bool(false)
		w.rtype(dstTyp)
		info := w.p.typIdx(dstTyp, w.dict)
		w.Bool(info.derived)
		w.rtype(srcTyp)
		return
	}

	if types2.Identical(srcTyp, dstTyp) {
		writeField()
		return
	}
	if !types2.AssignableTo(srcTyp, dstTyp) {
		w.p.fatalf(pos, "%v is not assignable to %v", srcTyp, dstTyp)
	}

	w.Code(exprConvert)
	w.Bool(true)
	w.typ(dstTyp)
	w.pos(pos)
	w.convRTTI(srcTyp, dstTyp)
	w.Bool(false)
	w.Bool(false)
	writeField()
}

func enumBindNames(cas syntax.Expr) []*syntax.Name {
	var names []*syntax.Name
	switch cas := cas.(type) {
	case *syntax.CallExpr:
		for _, arg := range cas.ArgList {
			if n, ok := arg.(*syntax.Name); ok && n.Value != "_" {
				names = append(names, n)
			}
		}
	case *syntax.EnumPattern:
		for _, f := range cas.Fields {
			if f != nil && f.Name != nil && f.Name.Value != "_" {
				names = append(names, f.Name)
			}
		}
	}
	return names
}

func (w *writer) declareEnumBindNames(cases []syntax.Expr) {
	for _, cas := range cases {
		for _, n := range enumBindNames(cas) {
			v, ok := w.p.info.Defs[n].(*types2.Var)
			if !ok {
				continue
			}
			if w.localsIdx != nil {
				if _, ok := w.localsIdx[v]; ok {
					continue
				}
			}
			w.Code(stmtAssign)
			w.pos(n)
			w.Len(1)
			w.assign(n)
			w.Sync(pkgbits.SyncMultiExpr)
			w.Bool(false)
			w.Len(0)
		}
	}
}

func (w *writer) writeEnumPayloadBind(pos syntax.Pos, lhs *syntax.Name, tag syntax.Expr, payloadIndex int) {
	obj := w.p.info.Defs[lhs]
	if obj == nil {
		obj = w.p.info.Uses[lhs]
	}
	if obj == nil {
		w.p.fatalf(lhs, "missing object for enum binding %v", lhs)
	}
	dstTyp := obj.(*types2.Var).Type()

	w.Code(stmtAssign)
	w.pos(pos)
	w.Len(1)
	w.Code(assignExpr)
	w.Code(exprLocal)
	w.useLocal(lhs.Pos(), obj.(*types2.Var))

	w.Sync(pkgbits.SyncMultiExpr)
	w.Bool(false)
	w.Len(1)
	w.writeEnumPayloadExpr(pos, tag, payloadIndex, dstTyp)
}

func (w *writer) writeEnumSwitchBindings(tag syntax.Expr, enumTyp *types2.Enum, cas syntax.Expr) {
	switch cas := cas.(type) {
	case *syntax.CallExpr:
		name := cas.Fun.(*syntax.Name)
		obj := w.p.info.Uses[name]
		v := enumVariantObj(enumTyp, obj)
		for i, arg := range cas.ArgList {
			n, ok := arg.(*syntax.Name)
			if !ok || n.Value == "_" {
				continue
			}
			w.writeEnumPayloadBind(cas.Pos(), n, tag, i)
		}
		_ = v

	case *syntax.EnumPattern:
		obj := w.p.info.Uses[cas.Variant]
		v := enumVariantObj(enumTyp, obj)
		for i, f := range v.Fields() {
			if f == nil {
				continue
			}
			var bind *syntax.Name
			for _, pf := range cas.Fields {
				if pf != nil && pf.Name != nil && pf.Name.Value == f.Name() {
					bind = pf.Name
					break
				}
			}
			if bind == nil || bind.Value == "_" {
				continue
			}
			w.writeEnumPayloadBind(cas.Pos(), bind, tag, i)
		}
	}
}

func (w *writer) writeEnumCaseTags(pos syntax.Pos, tagType types2.Type, enumTyp *types2.Enum, cases []syntax.Expr) {
	w.Sync(pkgbits.SyncExprList)
	w.Sync(pkgbits.SyncExprs)
	w.Len(len(cases))
	for _, cas := range cases {
		v, ok := enumVariantFromCase(w.p, enumTyp, cas)
		if !ok {
			w.p.fatalf(cas, "invalid enum switch case")
		}
		w.writeConstExpr(pos, tagType, constant.MakeInt64(v.Tag()))
	}
}

type enumPayloadSubst struct {
	tag   syntax.Expr
	index int
	typ   types2.Type
}

func (w *writer) enumSubstMap(tag syntax.Expr, enumTyp *types2.Enum, cases []syntax.Expr) map[string]enumPayloadSubst {
	m := make(map[string]enumPayloadSubst)
	for _, cas := range cases {
		binds := enumCaseBinds(w.p, enumTyp, cas)
		for name, idx := range binds {
			var obj types2.Object
			for _, n := range enumBindNames(cas) {
				if n.Value == name {
					obj = w.p.info.Defs[n]
					if obj == nil {
						obj = w.p.info.Uses[n]
					}
					break
				}
			}
			if obj == nil {
				continue
			}
			m[name] = enumPayloadSubst{tag: tag, index: idx, typ: obj.Type()}
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func (w *writer) writeEnumSwitchTag(tag syntax.Expr) {
	w.writeEnumFieldRef(tag, 0)
}

func (w *writer) writeEnumSwitchBody(stmt *syntax.SwitchStmt, enumTyp *types2.Enum, tag syntax.Expr) {
	structTyp := enumTyp.StructType()
	tagFieldType := structTyp.Field(0).Type()

	w.Bool(true)
	w.writeEnumSwitchTag(tag)

	w.Len(len(stmt.Body))
	for i, clause := range stmt.Body {
		if i > 0 {
			w.closeScope(clause.Pos())
		}
		w.openScope(clause.Pos())

		w.pos(clause)

		cases := syntax.UnpackListExpr(clause.Cases)
		w.writeEnumCaseTags(clause.Pos(), tagFieldType, enumTyp, cases)

		w.declareEnumBindNames(cases)
		for _, cas := range cases {
			w.writeEnumSwitchBindings(tag, enumTyp, cas)
		}

		w.stmts(clause.Body)
	}
	if len(stmt.Body) > 0 {
		w.closeScope(stmt.Rbrace)
	}
}

func (w *writer) enumSwitchExpr(expr *syntax.SwitchExpr, enumTyp *types2.Enum) {
	tv := w.p.typeAndValue(expr)
	w.Code(exprSwitchExpr)
	w.pos(expr)
	w.typ(tv.Type)

	structTyp := enumTyp.StructType()
	tagFieldType := structTyp.Field(0).Type()

	w.Bool(expr.Tag != nil)
	if expr.Tag != nil {
		w.writeEnumSwitchTag(expr.Tag)
	}

	w.Len(len(expr.Body))
	for _, c := range expr.Body {
		w.Bool(c.Cases != nil)
		if c.Cases != nil {
			cases := syntax.UnpackListExpr(c.Cases)
			w.Len(len(cases))
			for _, cas := range cases {
				v, ok := enumVariantFromCase(w.p, enumTyp, cas)
				if !ok {
					w.p.fatalf(cas, "invalid enum switch case")
				}
				w.writeConstExpr(cas.Pos(), tagFieldType, constant.MakeInt64(v.Tag()))
			}
		}
		if c.Cases != nil {
			cases := syntax.UnpackListExpr(c.Cases)
			w.enumSubst = w.enumSubstMap(expr.Tag, enumTyp, cases)
		}
		w.implicitConvExpr(tv.Type, c.Body)
		w.enumSubst = nil
	}
}
