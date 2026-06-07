// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	"fmt"
	"go/constant"
	. "internal/types/errors"
	"strconv"
	"strings"
)

// An Enum represents an enum type lowered to a struct with _tag and _pN payload fields.
type Enum struct {
	obj        *TypeName
	tparams    *TypeParamList
	variants   []*EnumVariant
	structType *Struct
	scope      *Scope
}

// An EnumVariant describes one variant of an enum.
type EnumVariant struct {
	name   string
	tag    int64
	obj    Object // *Const for unit variants, *Func for tuple/struct constructors
	tuple  []Type
	fields []*Var
}

// NewEnum returns a new enum type. Called only from buildEnum.
func NewEnum(obj *TypeName, tparams *TypeParamList, variants []*EnumVariant, structType *Struct, scope *Scope) *Enum {
	return &Enum{obj: obj, tparams: tparams, variants: variants, structType: structType, scope: scope}
}

func (e *Enum) Obj() *TypeName              { return e.obj }
func (e *Enum) TypeParams() *TypeParamList  { return e.tparams }
func (e *Enum) Variants() []*EnumVariant    { return e.variants }
func (e *Enum) StructType() *Struct         { return e.structType }
func (e *Enum) Scope() *Scope               { return e.scope }

func (e *Enum) Underlying() Type { return e.structType }
func (e *Enum) String() string   { return TypeString(e, nil) }

// AsEnum reports whether t is (or names) an enum type.
func AsEnum(t Type) (*Enum, bool) {
	t = Unalias(t)
	if n, ok := t.(*Named); ok {
		if n.enumType != nil {
			return n.enumType, true
		}
		if n.check != nil && n.obj != nil {
			if info := n.check.objMap[n.obj]; info != nil && info.enum != nil {
				return info.enum, true
			}
		}
		if e, ok := n.rhs().(*Enum); ok {
			return e, true
		}
	}
	if e, ok := t.(*Enum); ok {
		return e, true
	}
	return nil, false
}

func (v *EnumVariant) Name() string     { return v.name }
func (v *EnumVariant) Tag() int64       { return v.tag }
func (v *EnumVariant) Obj() Object      { return v.obj }
func (v *EnumVariant) Tuple() []Type    { return v.tuple }
func (v *EnumVariant) Fields() []*Var   { return v.fields }

// IsEnumVariant reports whether obj is an enum variant constructor or unit value.
func IsEnumVariant(obj Object) bool {
	switch obj := obj.(type) {
	case *Const:
		_, ok := AsEnum(obj.typ)
		return ok
	case *Func:
		if p := obj.parent; p != nil {
			return strings.HasPrefix(p.comment, "enum ")
		}
	}
	return false
}

func (check *Checker) enumDecl(obj *TypeName, edecl *syntax.EnumDecl) {
	assert(obj.typ == nil)

	if !check.verifyVersionf(edecl, go1_27, "enum") {
		obj.setType(Typ[Invalid])
		return
	}

	named := check.newNamed(obj, nil, nil)
	if len(edecl.TParamList) > 0 {
		if !check.verifyVersionf(edecl.TParamList[0], go1_18, "type parameter") {
			obj.setType(Typ[Invalid])
			return
		}
		check.openScope(edecl, "type parameters")
		defer check.closeScope()
		check.collectTypeParams(&named.tparams, edecl.TParamList)
	}

	enumTyp := check.buildEnum(named, obj, edecl)
	if info := check.objMap[obj]; info != nil {
		info.enum = enumTyp
	}
	named.enumType = enumTyp
	named.fromRHS = enumTyp
	named.SetUnderlying(enumTyp.structType)
}

func (check *Checker) buildEnum(named *Named, obj *TypeName, edecl *syntax.EnumDecl) *Enum {
	pos := syntax.StartPos(edecl)
	scope := NewScope(check.pkg.scope, pos, syntax.EndPos(edecl), "enum "+obj.Name())

	var variants []*EnumVariant
	seen := make(map[string]bool)
	var nextTag int64

	maxPayload := 0
	type payloadSlot struct {
		types []Type
	}
	var slots []payloadSlot

	for _, sv := range edecl.Variants {
		if sv == nil || sv.Name == nil {
			continue
		}
		name := sv.Name.Value
		if name == "" || name == "_" {
			check.errorf(sv.Name, InvalidSyntaxTree, "invalid enum variant name %s", name)
			continue
		}
		if seen[name] {
			check.errorf(sv.Name, DuplicateDecl, "%s redeclared", name)
			continue
		}
		seen[name] = true

		var tag int64
		if sv.Tag != nil {
			tv := check.enumTagValue(sv.Tag)
			if !tv.isValid() {
				continue
			}
			if tv.mode() != constant_ {
				check.errorf(sv.Tag, InvalidSyntaxTree, "enum variant tag must be an integer constant")
				continue
			}
			tag, _ = constant.Int64Val(tv.val)
			if tag < 0 {
				check.errorf(sv.Tag, InvalidSyntaxTree, "enum variant tag must be non-negative")
				continue
			}
			nextTag = tag + 1
		} else {
			tag = nextTag
			nextTag++
		}

		var tuple []Type
		for _, te := range sv.Types {
			t := check.typ(te)
			if !isValid(t) {
				t = Typ[Invalid]
			}
			tuple = append(tuple, t)
		}

		var fields []*Var
		for _, f := range sv.Fields {
			if f == nil || f.Name == nil {
				continue
			}
			ft := check.varType(f.Type)
			fields = append(fields, NewField(f.Pos(), check.pkg, f.Name.Value, ft, false))
		}

		payload := len(tuple)
		if len(fields) > payload {
			payload = len(fields)
		}
		if payload > maxPayload {
			for i := maxPayload; i < payload; i++ {
				slots = append(slots, payloadSlot{})
			}
			maxPayload = payload
		}
		for i, t := range tuple {
			slots[i].types = append(slots[i].types, t)
		}
		for i, f := range fields {
			slots[i].types = append(slots[i].types, f.typ)
		}

		variant := &EnumVariant{name: name, tag: tag, tuple: tuple, fields: fields}
		variants = append(variants, variant)
	}

	structFields := []*Var{NewField(pos, nil, "_tag", Typ[Int], false)}
	for i := 0; i < maxPayload; i++ {
		typ := check.enumPayloadType(slots[i].types)
		structFields = append(structFields, NewField(pos, nil, fmt.Sprintf("_p%d", i), typ, false))
	}
	structType := NewStruct(structFields, nil)

	enumType := named
	for _, variant := range variants {
		switch {
		case len(variant.tuple) == 0 && len(variant.fields) == 0:
			variant.obj = check.enumUnitVariant(variant, enumType, scope, pos)
		case len(variant.tuple) > 0:
			variant.obj = check.enumTupleVariant(variant, enumType, scope, pos)
		default:
			variant.obj = check.enumStructVariant(variant, enumType, scope, pos)
		}
		scope.insert(variant.name, variant.obj)
	}

	return NewEnum(obj, named.tparams, variants, structType, scope)
}

func (check *Checker) enumTagValue(e syntax.Expr) operand {
	var x operand
	check.expr(nil, &x, e)
	return x
}

func (check *Checker) enumPayloadType(types []Type) Type {
	if len(types) == 0 {
		return Typ[Invalid]
	}
	first := types[0]
	for _, t := range types[1:] {
		if !Identical(first, t) {
			return universeAny.Type()
		}
	}
	return first
}

func (check *Checker) enumUnitVariant(v *EnumVariant, enumType Type, scope *Scope, pos syntax.Pos) *Const {
	c := NewConst(pos, check.pkg, v.name, enumType, constant.MakeInt64(v.tag))
	c.parent = scope
	return c
}

func (check *Checker) enumTupleVariant(v *EnumVariant, enumType Type, scope *Scope, pos syntax.Pos) *Func {
	var params []*Var
	for i, t := range v.tuple {
		params = append(params, NewParam(pos, check.pkg, "arg"+strconv.Itoa(i), t))
	}
	result := NewParam(pos, check.pkg, "", enumType)
	result.SetKind(ResultVar)
	sig := NewSignatureType(nil, nil, nil, NewTuple(params...), NewTuple(result), false)
	fn := NewFunc(pos, check.pkg, v.name, sig)
	fn.parent = scope
	return fn
}

func (check *Checker) enumStructVariant(v *EnumVariant, enumType Type, scope *Scope, pos syntax.Pos) *Func {
	var params []*Var
	for _, f := range v.fields {
		params = append(params, NewParam(pos, check.pkg, f.name, f.typ))
	}
	result := NewParam(pos, check.pkg, "", enumType)
	result.SetKind(ResultVar)
	sig := NewSignatureType(nil, nil, nil, NewTuple(params...), NewTuple(result), false)
	fn := NewFunc(pos, check.pkg, v.name, sig)
	fn.parent = scope
	return fn
}

func (check *Checker) lookupEnumVariant(hint Type, name string) Object {
	enumTyp, ok := AsEnum(hint)
	if !ok || enumTyp.scope == nil {
		return nil
	}
	return enumTyp.scope.Lookup(name)
}

func (check *Checker) enumVariantOperand(x *operand, obj Object, e syntax.Expr) {
	switch obj := obj.(type) {
	case *Const:
		x.mode_ = value
		x.typ_ = obj.typ
		// Unit variants store the tag in Const.val, but the operand type is
		// the enum, not int. Leave val unset so noder lowers to a struct lit.
		x.expr = e
	case *Func:
		x.mode_ = value
		x.typ_ = obj.typ
		x.expr = e
	default:
		x.invalidate()
	}
	if x.isValid() {
		check.recordTypeAndValue(e, x.mode(), x.typ(), x.val)
	}
}

func (check *Checker) enumSelector(x *operand, e *syntax.SelectorExpr, typ Type, isTypeExpr bool) bool {
	if !isTypeExpr {
		return false
	}
	enumTyp, ok := AsEnum(typ)
	if !ok {
		return false
	}
	sel := e.Sel.Value
	obj := enumTyp.scope.Lookup(sel)
	if obj == nil {
		return false
	}
	check.recordUse(e.Sel, obj)
	check.enumVariantOperand(x, obj, e)
	return true
}

func (check *Checker) enumTypeExpr(e syntax.Expr) Type {
	switch e := e.(type) {
	case *syntax.Name:
		_, obj := check.lookupScope(e.Value)
		tn, ok := obj.(*TypeName)
		if !ok {
			return nil
		}
		if tn.typ == nil || tn.Pkg() == check.pkg {
			check.objDecl(tn)
		}
		typ := tn.Type()
		if !isValid(typ) {
			return nil
		}
		return typ

	case *syntax.SelectorExpr:
		if ident, ok := e.X.(*syntax.Name); ok {
			_, obj := check.lookupScope(ident.Value)
			pname, ok := obj.(*PkgName)
			if !ok {
				return nil
			}
			exp := pname.imported.scope.Lookup(e.Sel.Value)
			tn, ok := exp.(*TypeName)
			if !ok {
				return nil
			}
			typ := tn.Type()
			if !isValid(typ) {
				return nil
			}
			return typ
		}
		return nil

	case *syntax.IndexExpr:
		typ := check.typ(e)
		if !isValid(typ) {
			return nil
		}
		return typ

	default:
		return nil
	}
}

func (check *Checker) tryEnumVariantCall(x *operand, call *syntax.CallExpr, hint Type) bool {
	var enumType Type
	var variant *EnumVariant
	var fun syntax.Expr

	switch f := call.Fun.(type) {
	case *syntax.Name:
		obj := check.lookupEnumVariant(hint, f.Value)
		if obj == nil {
			return false
		}
		enumType = hint
		variant = enumVariantByName(enumFromHint(hint), f.Value)
		if variant == nil {
			return false
		}
		fun = f
		check.recordUse(f, obj)
		check.recordTypeAndValue(f, value, obj.Type(), nil)

	case *syntax.SelectorExpr:
		typ := check.enumTypeExpr(f.X)
		enumTyp, ok := AsEnum(typ)
		if !ok {
			return false
		}
		obj := enumTyp.scope.Lookup(f.Sel.Value)
		if obj == nil {
			return false
		}
		enumType = typ
		variant = enumVariantByName(enumTyp, f.Sel.Value)
		if variant == nil {
			return false
		}
		fun = f
		check.recordUse(f.Sel, obj)
		check.recordTypeAndValue(f, value, obj.Type(), nil)

	default:
		return false
	}

	if len(variant.tuple) == 0 && len(variant.fields) == 0 {
		if len(call.ArgList) != 0 {
			check.errorf(call, WrongArgCount, "%s expects no arguments", variant.name)
		}
		check.enumVariantOperand(x, variant.obj, call)
		x.expr = call
		return true
	}

	fn, ok := variant.obj.(*Func)
	if !ok {
		return false
	}
	sig := fn.typ.(*Signature)
	args := call.ArgList
	if hasDots(call) {
		check.errorf(call, BadDotDotDotSyntax, "invalid use of ... in enum variant call")
		x.invalidate()
		x.expr = call
		return true
	}
	if len(args) != sig.params.Len() {
		check.errorf(call, WrongArgCount, "wrong argument count for %s", variant.name)
		check.use(args...)
		x.invalidate()
		x.expr = call
		return true
	}
	for i, arg := range args {
		var a operand
		check.expr(nil, &a, arg)
		if !a.isValid() {
			x.invalidate()
			x.expr = call
			return true
		}
		check.assignment(&a, sig.params.At(i).typ, "argument to "+variant.name)
	}
	_ = fun
	x.mode_ = value
	x.typ_ = enumType
	x.expr = call
	check.recordTypeAndValue(call, value, enumType, nil)
	return true
}

func enumFromHint(hint Type) *Enum {
	enumTyp, _ := AsEnum(hint)
	return enumTyp
}

func enumVariantByName(enumTyp *Enum, name string) *EnumVariant {
	if enumTyp == nil {
		return nil
	}
	for _, v := range enumTyp.variants {
		if v.name == name {
			return v
		}
	}
	return nil
}

func enumVariantByObj(enumTyp *Enum, obj Object) *EnumVariant {
	if enumTyp == nil || obj == nil {
		return nil
	}
	for _, v := range enumTyp.variants {
		if v.obj == obj {
			return v
		}
	}
	return nil
}

func (check *Checker) tryEnumCompositeLit(x *operand, e *syntax.CompositeLit, hint Type) bool {
	if e.Type == nil {
		return false
	}
	sel, ok := syntax.Unparen(e.Type).(*syntax.SelectorExpr)
	if !ok {
		return false
	}
	typ := check.enumTypeExpr(sel.X)
	enumTyp, ok := AsEnum(typ)
	if !ok {
		return false
	}
	obj := enumTyp.scope.Lookup(sel.Sel.Value)
	if obj == nil {
		return false
	}
	variant := enumVariantByObj(enumTyp, obj)
	if variant == nil || len(variant.fields) == 0 {
		return false
	}
	check.recordUse(sel.Sel, obj)

	if len(e.ElemList) == 0 {
		check.error(e, InvalidLitField, "enum struct variant requires fields")
		x.invalidate()
		return true
	}

	visited := make(map[string]bool)
	for _, el := range e.ElemList {
		kv, ok := el.(*syntax.KeyValueExpr)
		if !ok {
			check.error(el, MixedStructLit, "enum struct variant requires keyed fields")
			continue
		}
		key, ok := kv.Key.(*syntax.Name)
		if !ok {
			check.errorf(kv, InvalidLitField, "invalid field name %s in enum literal", kv.Key)
			continue
		}
		if visited[key.Value] {
			check.errorf(key, DuplicateDecl, "%s repeated in enum literal", key.Value)
			continue
		}
		visited[key.Value] = true
		var ftyp Type
		found := false
		for _, f := range variant.fields {
			if f.name == key.Value {
				ftyp = f.typ
				found = true
				break
			}
		}
		if !found {
			check.errorf(key, MissingFieldOrMethod, "unknown field %s in %s", key.Value, variant.name)
			continue
		}
		var v operand
		check.expr(nil, &v, kv.Value)
		if v.isValid() {
			check.assignment(&v, ftyp, "field value")
		}
	}
	for _, f := range variant.fields {
		if !visited[f.name] {
			check.errorf(e, MissingLitField, "missing field %s in %s literal", f.name, variant.name)
		}
	}

	x.mode_ = value
	x.typ_ = typ
	x.expr = e
	return true
}

func (check *Checker) enumSwitchStmt(inner stmtContext, s *syntax.SwitchStmt, tag Type, enumTyp *Enum) {
	check.multipleSwitchDefaults(s.Body)

	hasDefault := false
	for _, clause := range s.Body {
		if clause != nil && clause.Cases == nil {
			hasDefault = true
			break
		}
	}

	covered := make(map[string]bool)
	for i, clause := range s.Body {
		if clause == nil {
			continue
		}
		inner := inner
		if i+1 < len(s.Body) {
			inner |= fallthroughOk
		} else {
			inner |= finalSwitchCase
		}
		for _, c := range syntax.UnpackListExpr(clause.Cases) {
			check.enumCasePattern(tag, enumTyp, c, covered)
		}
		check.openScope(clause, "case")
		check.stmtList(inner, clause.Body)
		check.closeScope()
	}

	if !hasDefault {
		for _, v := range enumTyp.variants {
			if !covered[v.name] {
				check.errorf(s, InvalidSyntaxTree, "switch on %s is not exhaustive: missing case %s", tag, v.name)
			}
		}
	}
}

func (check *Checker) enumSwitchExpr(x *operand, e *syntax.SwitchExpr, tag Type, enumTyp *Enum) {
	check.multipleSwitchDefaults(switchExprClausesToCaseClauses(e.Body))

	hasDefault := false
	for _, clause := range e.Body {
		if clause != nil && clause.Cases == nil {
			hasDefault = true
			break
		}
	}

	covered := make(map[string]bool)
	var arms []*operand
	for _, clause := range e.Body {
		if clause == nil {
			continue
		}
		if clause.Cases != nil {
			for _, c := range syntax.UnpackListExpr(clause.Cases) {
				check.enumCasePattern(tag, enumTyp, c, covered)
			}
		}
		var arm operand
		check.expr(nil, &arm, clause.Body)
		if !arm.isValid() {
			x.invalidate()
			return
		}
		arms = append(arms, &arm)
	}

	if !hasDefault {
		for _, v := range enumTyp.variants {
			if !covered[v.name] {
				check.errorf(e, InvalidSyntaxTree, "switch on %s is not exhaustive: missing case %s", tag, v.name)
			}
		}
	}

	if len(arms) == 0 {
		check.errorf(e, InvalidSyntaxTree, "switch expression must have at least one case")
		x.invalidate()
		return
	}

	t := check.mergeBranchTypes(e, arms)
	if !isValid(t) {
		x.invalidate()
		return
	}
	x.mode_ = value
	x.typ_ = t
	x.expr = e
}

func (check *Checker) enumCasePattern(tag Type, enumTyp *Enum, pattern syntax.Expr, covered map[string]bool) {
	switch p := pattern.(type) {
	case *syntax.Name:
		obj := enumTyp.scope.Lookup(p.Value)
		if obj == nil {
			check.errorf(p, UndeclaredName, "undefined: %s", p.Value)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil {
			check.errorf(p, InvalidSyntaxTree, "%s is not an enum variant", p.Value)
			return
		}
		if len(v.tuple) > 0 || len(v.fields) > 0 {
			check.errorf(p, WrongArgCount, "enum variant %s requires arguments in case pattern", v.name)
			return
		}
		covered[v.name] = true
		check.recordUse(p, obj)

	case *syntax.CallExpr:
		name, ok := p.Fun.(*syntax.Name)
		if !ok {
			check.error(p, InvalidSyntaxTree, "invalid enum case pattern")
			return
		}
		obj := enumTyp.scope.Lookup(name.Value)
		if obj == nil {
			check.errorf(name, UndeclaredName, "undefined: %s", name.Value)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil {
			check.errorf(name, InvalidSyntaxTree, "%s is not an enum variant", name.Value)
			return
		}
		if len(v.tuple) == 0 {
			check.errorf(p, WrongArgCount, "enum variant %s expects no arguments", v.name)
			return
		}
		args := p.ArgList
		if len(args) != len(v.tuple) {
			check.errorf(p, WrongArgCount, "wrong argument count for %s", v.name)
			return
		}
		for i, arg := range args {
			n, ok := arg.(*syntax.Name)
			if !ok || (n.Value != "_" && !isValidName(n.Value)) {
				check.errorf(arg, InvalidSyntaxTree, "enum case pattern requires identifier bindings")
				continue
			}
			if n.Value != "_" {
				vobj := newVar(LocalVar, n.Pos(), check.pkg, n.Value, v.tuple[i])
				check.declare(check.scope, n, vobj, n.Pos())
			}
		}
		covered[v.name] = true
		check.recordUse(name, obj)

	case *syntax.EnumPattern:
		if p.Variant == nil {
			check.error(p, InvalidSyntaxTree, "invalid enum case pattern")
			return
		}
		obj := enumTyp.scope.Lookup(p.Variant.Value)
		if obj == nil {
			check.errorf(p.Variant, UndeclaredName, "undefined: %s", p.Variant.Value)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil || len(v.fields) == 0 {
			check.errorf(p.Variant, InvalidSyntaxTree, "%s is not a struct enum variant", p.Variant.Value)
			return
		}
		seen := make(map[string]bool)
		for _, f := range p.Fields {
			if f == nil || f.Name == nil {
				continue
			}
			name := f.Name.Value
			if name == "" {
				continue
			}
			if seen[name] {
				check.errorf(f.Name, DuplicateDecl, "%s repeated in enum pattern", name)
				continue
			}
			seen[name] = true
			var ftyp Type
			found := false
			for _, vf := range v.fields {
				if vf.name == name {
					ftyp = vf.typ
					found = true
					break
				}
			}
			if !found {
				check.errorf(f.Name, MissingFieldOrMethod, "unknown field %s in %s", name, v.name)
				continue
			}
			if name != "_" {
				vobj := newVar(LocalVar, f.Name.Pos(), check.pkg, name, ftyp)
				check.declare(check.scope, f.Name, vobj, f.Name.Pos())
			}
		}
		for _, arg := range p.Args {
			if arg == nil {
				continue
			}
			if arg.Value != "_" {
				vobj := newVar(LocalVar, arg.Pos(), check.pkg, arg.Value, Typ[Invalid])
				check.declare(check.scope, arg, vobj, arg.Pos())
			}
		}
		covered[v.name] = true
		check.recordUse(p.Variant, obj)

	default:
		check.error(pattern, InvalidSyntaxTree, "invalid enum case pattern")
	}
}
