
package typecheck

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

const (
	optionalHasValueField = "hasValue"
	optionalValueField    = "value"
)

// IsOptionalStruct reports whether t is the lowered representation of T?.
func IsOptionalStruct(t *types.Type) bool {
	if t == nil || !t.IsStruct() || t.NumFields() != 2 {
		return false
	}
	f0, f1 := t.Field(0), t.Field(1)
	return f0.Sym.Name == optionalHasValueField && f0.Type.Kind() == types.TBOOL &&
		f1.Sym.Name == optionalValueField
}

func optionalField(pos src.XPos, x ir.Node, name string) ir.Node {
	typ := x.Type()
	if !IsOptionalStruct(typ) {
		base.FatalfAt(pos, "optional field access on non-optional struct %L", x)
	}
	for i := 0; i < typ.NumFields(); i++ {
		if f := typ.Field(i); f.Sym.Name == name {
			return DotField(pos, x, i)
		}
	}
	base.FatalfAt(pos, "optional struct %v has no field %q", typ, name)
	return nil
}

// OptionalHasValue returns x.hasValue for lowered T? values.
func OptionalHasValue(pos src.XPos, x ir.Node) ir.Node {
	return optionalField(pos, x, optionalHasValueField)
}

func TypedFalse(pos src.XPos) ir.Node {
	return DefaultLit(Expr(ir.NewBool(pos, false)), types.Types[types.TBOOL])
}

func TypedTrue(pos src.XPos) ir.Node {
	return DefaultLit(Expr(ir.NewBool(pos, true)), types.Types[types.TBOOL])
}

// OptionalIsNil returns true when x is a nil T? (!hasValue).
func OptionalIsNil(pos src.XPos, x ir.Node) ir.Node {
	cmp := ir.NewBinaryExpr(pos, ir.OEQ, OptionalHasValue(pos, x), TypedFalse(pos))
	cmp.SetType(types.Types[types.TBOOL])
	cmp.SetTypecheck(1)
	return cmp
}

// OptionalValue returns x.value for lowered T? values.
func OptionalValue(pos src.XPos, x ir.Node) ir.Node {
	return optionalField(pos, x, optionalValueField)
}

// OptionalStructLit builds {hasValue: hv, value: val} for typ.
func OptionalStructLit(pos src.XPos, typ *types.Type, hasValue, val ir.Node) *ir.CompLitExpr {
	if !IsOptionalStruct(typ) {
		base.FatalfAt(pos, "OptionalStructLit on non-optional struct %v", typ)
	}
	lit := ir.NewCompLitExpr(pos, ir.OSTRUCTLIT, typ, []ir.Node{
		ir.NewStructKeyExpr(pos, typ.Field(0), hasValue),
		ir.NewStructKeyExpr(pos, typ.Field(1), val),
	})
	lit.SetTypecheck(1)
	return lit
}

// OptionalWrapNil returns a nil T? as {hasValue: false, value: zero}.
func OptionalWrapNil(pos src.XPos, typ *types.Type) ir.Node {
	zero := Expr(ir.NewZero(pos, typ.Field(1).Type))
	return OptionalStructLit(pos, typ, TypedFalse(pos), zero)
}

// OptionalWrapValue returns a non-nil T? as {hasValue: true, value: val}.
func OptionalWrapValue(pos src.XPos, typ *types.Type, val ir.Node) ir.Node {
	val = DefaultLit(val, typ.Field(1).Type)
	return OptionalStructLit(pos, typ, TypedTrue(pos), val)
}

// OptionalWrapFromPtr returns a T? from *T, mapping nil pointers to nil T?.
func OptionalWrapFromPtr(pos src.XPos, typ *types.Type, ptr ir.Node) ir.Node {
	ptr = Expr(ptr)
	elem := typ.Field(1).Type
	nilOpt := OptionalWrapNil(pos, typ)
	star := Expr(ir.NewStarExpr(pos, ptr))
	star.SetType(elem)
	valOpt := OptionalWrapValue(pos, typ, star)
	niln := ir.NewNilExpr(pos, ptr.Type())
	cmp := ir.NewBinaryExpr(pos, ir.OEQ, ptr, niln)
	cmp.SetType(types.Types[types.TBOOL])
	return Expr(ir.NewIfExpr(pos, typ, cmp, nilOpt, valOpt))
}
