// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
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

	variantByName map[string]*EnumVariant
	variantByObj  map[Object]*EnumVariant
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
	byName := make(map[string]*EnumVariant, len(variants))
	byObj := make(map[Object]*EnumVariant, len(variants))
	for _, v := range variants {
		byName[v.name] = v
		if v.obj != nil {
			byObj[v.obj] = v
		}
	}
	return &Enum{
		obj:           obj,
		tparams:       tparams,
		variants:      variants,
		structType:    structType,
		scope:         scope,
		variantByName: byName,
		variantByObj:  byObj,
	}
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
			if info := n.check.objMap[n.obj]; info != nil && info.enumTyp != nil {
				return info.enumTyp, true
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

func (check *Checker) enumDecl(obj *TypeName, edecl *ast.EnumDecl) {
	assert(obj.typ == nil)

	if !check.verifyVersionf(edecl, go1_27, "enum") {
		obj.setType(Typ[Invalid])
		return
	}

	named := check.newNamed(obj, nil, nil)
	if edecl.TypeParams != nil && len(edecl.TypeParams.List) > 0 {
		if !check.verifyVersionf(edecl.TypeParams.List[0], go1_18, "type parameter") {
			obj.setType(Typ[Invalid])
			return
		}
		check.openScope(edecl, "type parameters")
		defer check.closeScope()
		check.collectTypeParams(&named.tparams, edecl.TypeParams)
	}

	enumTyp := check.buildEnum(named, obj, edecl)
	if info := check.objMap[obj]; info != nil {
		info.enumTyp = enumTyp
	}
	named.enumType = enumTyp
	named.fromRHS = enumTyp
	named.SetUnderlying(enumTyp.structType)
}

func (check *Checker) buildEnum(named *Named, obj *TypeName, edecl *ast.EnumDecl) *Enum {
	pos := edecl.Pos()
	scope := NewScope(check.pkg.scope, pos, edecl.End(), "enum "+obj.Name())

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
		name := sv.Name.Name
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
				check.error(sv.Tag, InvalidSyntaxTree, "enum variant tag must be an integer constant")
				continue
			}
			tag, _ = constant.Int64Val(tv.val)
			if tag < 0 {
				check.error(sv.Tag, InvalidSyntaxTree, "enum variant tag must be non-negative")
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
		for _, f := range enumVariantFields(sv) {
			if f == nil {
				continue
			}
			ft := check.varType(f.Type)
			if !isValid(ft) {
				ft = Typ[Invalid]
			}
			if len(f.Names) == 0 {
				continue
			}
			for _, n := range f.Names {
				if n == nil || n.Name == "" || n.Name == "_" {
					continue
				}
				fields = append(fields, NewField(n.Pos(), check.pkg, n.Name, ft, false))
			}
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

func (check *Checker) enumTagValue(e ast.Expr) operand {
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

func (check *Checker) enumUnitVariant(v *EnumVariant, enumType Type, scope *Scope, pos token.Pos) *Const {
	c := NewConst(pos, check.pkg, v.name, enumType, constant.MakeInt64(v.tag))
	c.parent = scope
	return c
}

func (check *Checker) enumTupleVariant(v *EnumVariant, enumType Type, scope *Scope, pos token.Pos) *Func {
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

func (check *Checker) enumStructVariant(v *EnumVariant, enumType Type, scope *Scope, pos token.Pos) *Func {
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

// isEnumVariant reports whether obj is an enum variant in the current package.
func isEnumVariant(obj Object) bool {
	name := obj.Name()
	var hint Type
	switch obj := obj.(type) {
	case *Const:
		hint = obj.Type()
	case *Func:
		if obj.Signature().Recv() != nil || obj.Signature().Results().Len() != 1 {
			return false
		}
		hint = obj.Signature().Results().At(0).Type()
	default:
		return false
	}
	e, ok := AsEnum(hint)
	if !ok || e.Scope() == nil || e.Scope().Lookup(name) != obj {
		return false
	}
	return true
}

func (check *Checker) lookupEnumVariant(hint Type, name string) Object {
	enumTyp, ok := AsEnum(hint)
	if !ok || enumTyp.scope == nil {
		return nil
	}
	return enumTyp.scope.Lookup(name)
}

// lookupPkgEnumVariant resolves an unqualified enum variant name at package level.
// It succeeds only when exactly one enum in the current package or its imports
// defines the variant.
func (check *Checker) lookupPkgEnumVariant(name string) Object {
	var found Object
	findInScope := func(scope *Scope) Object {
		if scope == nil {
			return nil
		}
		for _, n := range scope.Names() {
			obj := scope.Lookup(n)
			tn, ok := obj.(*TypeName)
			if !ok {
				continue
			}
			if et, ok := AsEnum(tn.Type()); ok && et.scope != nil {
				if v := et.scope.Lookup(name); v != nil {
					return v
				}
			}
		}
		return nil
	}
	if v := findInScope(check.pkg.scope); v != nil {
		found = v
	}
	for _, imp := range check.imports {
		if imp == nil || imp.imported == nil {
			continue
		}
		if v := findInScope(imp.imported.scope); v != nil {
			if found != nil {
				return nil // ambiguous
			}
			found = v
		}
	}
	return found
}

// enumTypeForVariant finds the enum type defining obj (a variant member).
func (check *Checker) enumTypeForVariant(obj Object) (Type, *Enum) {
	findInScope := func(scope *Scope) (Type, *Enum) {
		if scope == nil {
			return nil, nil
		}
		for _, n := range scope.Names() {
			tn, ok := scope.Lookup(n).(*TypeName)
			if !ok {
				continue
			}
			if et, ok := AsEnum(tn.Type()); ok && enumVariantByObj(et, obj) != nil {
				return tn.Type(), et
			}
		}
		return nil, nil
	}
	if t, e := findInScope(check.pkg.scope); e != nil {
		return t, e
	}
	for _, imp := range check.imports {
		if imp == nil || imp.imported == nil {
			continue
		}
		if t, e := findInScope(imp.imported.scope); e != nil {
			return t, e
		}
	}
	return nil, nil
}

// isEnumUnitVariantDefault reports whether x is a unit enum variant suitable
// as a default argument for typ.
func (check *Checker) isEnumUnitVariantDefault(x *operand, typ Type) bool {
	if x.mode() != value {
		return false
	}
	enumTyp, ok := AsEnum(typ)
	if !ok {
		return false
	}
	obj := check.enumVariantFromExpr(x.expr, enumTyp)
	if obj == nil {
		return false
	}
	c, ok := obj.(*Const)
	return ok && isEnumVariant(c)
}

func (check *Checker) enumVariantFromExpr(e ast.Expr, enumTyp *Enum) Object {
	switch e := e.(type) {
	case *ast.Ident:
		return check.lookupEnumVariant(enumTyp, e.Name)
	case *ast.SelectorExpr:
		t := check.enumTypeExpr(e.X)
		et, ok := AsEnum(t)
		if !ok || !Identical(et, enumTyp) {
			return nil
		}
		return enumTyp.scope.Lookup(e.Sel.Name)
	}
	return nil
}

func (check *Checker) enumVariantOperand(x *operand, obj Object, e ast.Expr) {
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

func (check *Checker) enumSelector(x *operand, e *ast.SelectorExpr, typ Type, isTypeExpr bool) bool {
	if !isTypeExpr {
		return false
	}
	enumTyp, ok := AsEnum(typ)
	if !ok {
		return false
	}
	sel := e.Sel.Name
	obj := enumTyp.scope.Lookup(sel)
	if obj == nil {
		return false
	}
	check.recordUse(e.Sel, obj)
	check.enumVariantOperand(x, obj, e)
	return true
}

func (check *Checker) enumTypeExpr(e ast.Expr) Type {
	switch e := e.(type) {
	case *ast.Ident:
		_, obj := check.lookupScope(e.Name)
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

	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			_, obj := check.lookupScope(ident.Name)
			pname, ok := obj.(*PkgName)
			if !ok {
				return nil
			}
			exp := pname.imported.scope.Lookup(e.Sel.Name)
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

	case *ast.IndexExpr, *ast.IndexListExpr:
		ix := unpackIndexedExpr(e)
		var x operand
		check.exprOrType(&x, ix.x, true)
		if x.mode() != typexpr {
			return nil
		}
		typ := check.instantiatedType(ix)
		if !isValid(typ) {
			return nil
		}
		return typ

	default:
		return nil
	}
}

func (check *Checker) tryEnumVariantCall(x *operand, call *ast.CallExpr, hint Type) bool {
	var enumType Type
	var variant *EnumVariant
	var fun ast.Expr

	switch f := call.Fun.(type) {
	case *ast.Ident:
		obj := check.lookupEnumVariant(hint, f.Name)
		if obj == nil {
			return false
		}
		enumType = hint
		variant = enumVariantByName(enumFromHint(hint), f.Name)
		if variant == nil {
			return false
		}
		fun = f
		check.recordUse(f, obj)
		check.recordTypeAndValue(f, value, obj.Type(), nil)

	case *ast.SelectorExpr:
		typ := check.enumTypeExpr(f.X)
		enumTyp, ok := AsEnum(typ)
		if !ok {
			return false
		}
		obj := enumTyp.scope.Lookup(f.Sel.Name)
		if obj == nil {
			return false
		}
		enumType = typ
		variant = enumVariantByName(enumTyp, f.Sel.Name)
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
		if len(call.Args) != 0 {
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
	args := call.Args
	if hasDots(call) {
		check.error(call, BadDotDotDotSyntax, "invalid use of ... in enum variant call")
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
	if enumTyp.variantByName != nil {
		return enumTyp.variantByName[name]
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
	if enumTyp.variantByObj != nil {
		return enumTyp.variantByObj[obj]
	}
	for _, v := range enumTyp.variants {
		if v.obj == obj {
			return v
		}
	}
	return nil
}

func (check *Checker) tryEnumCompositeLit(x *operand, e *ast.CompositeLit, hint Type) bool {
	if e.Type == nil {
		return false
	}
	var enumType Type
	var enumTyp *Enum
	var variantName string
	var use *ast.Ident
	switch t := ast.Unparen(e.Type).(type) {
	case *ast.Ident:
		variantName = t.Name
		use = t
		if et, ok := AsEnum(hint); ok {
			enumTyp = et
			enumType = hint
		} else if obj := check.lookupPkgEnumVariant(variantName); obj != nil {
			enumType, enumTyp = check.enumTypeForVariant(obj)
		}
	case *ast.SelectorExpr:
		enumType = check.enumTypeExpr(t.X)
		var ok bool
		enumTyp, ok = AsEnum(enumType)
		if !ok {
			return false
		}
		variantName = t.Sel.Name
		use = t.Sel
	default:
		return false
	}
	if enumTyp == nil {
		return false
	}
	obj := enumTyp.scope.Lookup(variantName)
	if obj == nil {
		return false
	}
	variant := enumVariantByObj(enumTyp, obj)
	if variant == nil || len(variant.fields) == 0 {
		return false
	}
	check.recordUse(use, obj)

	if len(e.Elts) == 0 {
		check.error(e, InvalidLitField, "enum struct variant requires fields")
		x.invalidate()
		return true
	}

	visited := make(map[string]bool)
	for _, el := range e.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			check.error(el, MixedStructLit, "enum struct variant requires keyed fields")
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			check.errorf(kv, InvalidLitField, "invalid field name %s in enum literal", kv.Key)
			continue
		}
		if visited[key.Name] {
			check.errorf(key, DuplicateDecl, "%s repeated in enum literal", key.Name)
			continue
		}
		visited[key.Name] = true
		var ftyp Type
		found := false
		for _, f := range variant.fields {
			if f.name == key.Name {
				ftyp = f.typ
				found = true
				break
			}
		}
		if !found {
			check.errorf(key, MissingFieldOrMethod, "unknown field %s in %s", key.Name, variant.name)
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
	x.typ_ = enumType
	x.expr = e
	return true
}

func (check *Checker) enumSwitchStmt(inner stmtContext, s *ast.SwitchStmt, tag Type, enumTyp *Enum) {
	check.multipleDefaults(s.Body.List)

	hasDefault := false
	for _, c := range s.Body.List {
		clause, _ := c.(*ast.CaseClause)
		if clause != nil && len(clause.List) == 0 {
			hasDefault = true
			break
		}
	}

	covered := make(map[string]bool)
	for i, c := range s.Body.List {
		clause, _ := c.(*ast.CaseClause)
		if clause == nil {
			continue
		}
		inner := inner
		if i+1 < len(s.Body.List) {
			inner |= fallthroughOk
		} else {
			inner |= finalSwitchCase
		}
		for _, pat := range clause.List {
			check.enumCasePattern(tag, enumTyp, pat, covered)
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

func (check *Checker) enumSwitchExpr(x *operand, e *ast.SwitchExpr, tag Type, enumTyp *Enum) {
	check.multipleSwitchExprDefaults(e.Body)

	hasDefault := false
	for _, clause := range e.Body {
		if clause != nil && len(clause.Cases) == 0 {
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
		if len(clause.Cases) > 0 {
			for _, c := range clause.Cases {
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
		check.error(e, InvalidSyntaxTree, "switch expression must have at least one case")
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

func (check *Checker) enumCasePattern(tag Type, enumTyp *Enum, pattern ast.Expr, covered map[string]bool) {
	switch p := pattern.(type) {
	case *ast.Ident:
		obj := enumTyp.scope.Lookup(p.Name)
		if obj == nil {
			check.errorf(p, UndeclaredName, "undefined: %s", p.Name)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil {
			check.errorf(p, InvalidSyntaxTree, "%s is not an enum variant", p.Name)
			return
		}
		if len(v.tuple) > 0 || len(v.fields) > 0 {
			check.errorf(p, WrongArgCount, "enum variant %s requires arguments in case pattern", v.name)
			return
		}
		covered[v.name] = true
		check.recordUse(p, obj)

	case *ast.CallExpr:
		name, ok := p.Fun.(*ast.Ident)
		if !ok {
			check.error(p, InvalidSyntaxTree, "invalid enum case pattern")
			return
		}
		obj := enumTyp.scope.Lookup(name.Name)
		if obj == nil {
			check.errorf(name, UndeclaredName, "undefined: %s", name.Name)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil {
			check.errorf(name, InvalidSyntaxTree, "%s is not an enum variant", name.Name)
			return
		}
		if len(v.tuple) == 0 {
			check.errorf(p, WrongArgCount, "enum variant %s expects no arguments", v.name)
			return
		}
		args := p.Args
		if len(args) != len(v.tuple) {
			check.errorf(p, WrongArgCount, "wrong argument count for %s", v.name)
			return
		}
		for i, arg := range args {
			n, ok := arg.(*ast.Ident)
			if !ok || (n.Name != "_" && !isValidName(n.Name)) {
				check.error(arg, InvalidSyntaxTree, "enum case pattern requires identifier bindings")
				continue
			}
			if n.Name != "_" {
				vobj := newVar(LocalVar, n.Pos(), check.pkg, n.Name, v.tuple[i])
				check.declare(check.scope, n, vobj, n.Pos())
			}
		}
		covered[v.name] = true
		check.recordUse(name, obj)

	case *ast.EnumPatternExpr:
		if p.Variant == nil {
			check.error(p, InvalidSyntaxTree, "invalid enum case pattern")
			return
		}
		obj := enumTyp.scope.Lookup(p.Variant.Name)
		if obj == nil {
			check.errorf(p.Variant, UndeclaredName, "undefined: %s", p.Variant.Name)
			return
		}
		v := enumVariantByObj(enumTyp, obj)
		if v == nil || len(v.fields) == 0 {
			check.errorf(p.Variant, InvalidSyntaxTree, "%s is not a struct enum variant", p.Variant.Name)
			return
		}
		if len(p.Fields) == 0 {
			covered[v.name] = true
			check.recordUse(p.Variant, obj)
			return
		}
		if len(p.Fields) > len(v.fields) {
			check.errorf(p.Variant, WrongArgCount, "enum pattern for %s has too many fields (got %d, variant has %d)", p.Variant.Name, len(p.Fields), len(v.fields))
			return
		}
		allWildcard := true
		for _, f := range p.Fields {
			if f == nil || fieldName(f) != "_" {
				allWildcard = false
				break
			}
		}
		if allWildcard && len(p.Fields) != len(v.fields) {
			check.errorf(p.Variant, WrongArgCount, "enum pattern for %s must list every field as _ or use an empty pattern (got %d, want %d)", p.Variant.Name, len(p.Fields), len(v.fields))
			return
		}
		if len(p.Fields) == len(v.fields) {
			for i, f := range p.Fields {
				if f == nil || fieldNameIdent(f) == nil {
					continue
				}
				name := fieldName(f)
				if name == "_" {
					continue
				}
				if v.fields[i].name != name {
					check.errorf(fieldNameIdent(f), InvalidSyntaxTree, "field %s does not match %s in %s", name, v.fields[i].name, v.name)
					continue
				}
				vobj := newVar(LocalVar, fieldNameIdent(f).Pos(), check.pkg, name, v.fields[i].typ)
				check.declare(check.scope, fieldNameIdent(f), vobj, fieldNameIdent(f).Pos())
			}
			covered[v.name] = true
			check.recordUse(p.Variant, obj)
			return
		}
		seen := make(map[string]bool)
		for _, f := range p.Fields {
			if f == nil || fieldNameIdent(f) == nil {
				continue
			}
			name := fieldName(f)
			if name == "" {
				continue
			}
			if seen[name] {
				check.errorf(fieldNameIdent(f), DuplicateDecl, "%s repeated in enum pattern", name)
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
				check.errorf(fieldNameIdent(f), MissingFieldOrMethod, "unknown field %s in %s", name, v.name)
				continue
			}
			if name != "_" {
				vobj := newVar(LocalVar, fieldNameIdent(f).Pos(), check.pkg, name, ftyp)
				check.declare(check.scope, fieldNameIdent(f), vobj, fieldNameIdent(f).Pos())
			}
		}
		for _, arg := range p.Args {
			if arg == nil {
				continue
			}
			if arg.Name != "_" {
				vobj := newVar(LocalVar, arg.Pos(), check.pkg, arg.Name, Typ[Invalid])
				check.declare(check.scope, arg, vobj, arg.Pos())
			}
		}
		covered[v.name] = true
		check.recordUse(p.Variant, obj)

	default:
		check.error(pattern, InvalidSyntaxTree, "invalid enum case pattern")
	}
}
