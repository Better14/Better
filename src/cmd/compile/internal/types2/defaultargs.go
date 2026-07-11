
package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	. "internal/types/errors"
)

func (check *Checker) paramDefault(v *Var, expr syntax.Expr) {
	var x operand
	check.rawExpr(nil, &x, expr, v.typ, false)
	if !x.isValid() {
		return
	}
	if x.mode() != constant_ && !check.isEnumUnitVariantDefault(&x, v.typ) {
		check.error(expr, InvalidSyntaxTree, "default argument must be compile-time constant")
		return
	}
	check.assignment(&x, v.typ, "default argument")
	if !x.isValid() {
		return
	}
	v.defExpr = expr
	if x.mode() == constant_ {
		v.defVal = x.val
	} else {
		v.defVal = nil
	}
}

func (check *Checker) validateParamDefaults(params []*Var) {
	seen := false
	for _, p := range params {
		if p.defExpr != nil {
			seen = true
		} else if seen {
			check.errorf(p.pos, InvalidSyntaxTree, "missing default value for parameter %q", p.name)
		}
	}
}

func sigMinParams(sig *Signature) int {
	n := sig.params.Len()
	min := 0
	for i := 0; i < n; i++ {
		p := sig.params.vars[i]
		if p.defExpr == nil && (p.defVal == nil || p.defVal.Kind() == constant.Unknown) {
			min++
		} else {
			break
		}
	}
	return min
}

func (check *Checker) appendDefaultArgs(call *syntax.CallExpr, sig *Signature, args []*operand) []*operand {
	if sig.variadic {
		return args
	}
	npars := sig.params.Len()
	nargs := len(args)
	if nargs >= npars {
		return args
	}
	out := append([]*operand(nil), args...)
	var extras []syntax.Expr
	for i := nargs; i < npars; i++ {
		v := sig.params.vars[i]
		if v.defExpr == nil {
			if v.defVal == nil || v.defVal.Kind() == constant.Unknown {
				break
			}
			var d operand
			d.mode_ = constant_
			d.typ_ = v.typ
			d.val = v.defVal
			out = append(out, &d)
			if lit := check.syntaxForDefaultVal(call.Pos(), v); lit != nil {
				check.recordTypeAndValue(lit, constant_, v.typ, v.defVal)
				extras = append(extras, lit)
			}
			continue
		}
		var d operand
		check.rawExpr(nil, &d, v.defExpr, v.typ, false)
		if !d.isValid() {
			break
		}
		out = append(out, &d)
		extras = append(extras, v.defExpr)
	}
	if len(out) == npars {
		call.ArgList = append(call.ArgList, extras...)
		return out
	}
	return args
}

func (check *Checker) syntaxForDefaultVal(pos syntax.Pos, v *Var) syntax.Expr {
	if v.defExpr != nil {
		return v.defExpr
	}
	if v.defVal == nil || v.defVal.Kind() == constant.Unknown {
		return nil
	}
	lit := new(syntax.BasicLit)
	lit.SetPos(pos)
	switch v.defVal.Kind() {
	case constant.String:
		lit.Kind = syntax.StringLit
		lit.Value = constant.StringVal(v.defVal)
	case constant.Int:
		lit.Kind = syntax.IntLit
		lit.Value = v.defVal.String()
	case constant.Float:
		lit.Kind = syntax.FloatLit
		lit.Value = v.defVal.String()
	case constant.Bool:
		lit.Kind = syntax.IntLit
		if constant.BoolVal(v.defVal) {
			lit.Value = "true"
		} else {
			lit.Value = "false"
		}
	default:
		return nil
	}
	return lit
}

func (check *Checker) overloadArgOperand(args []*operand, nargs, i int, v *Var) (operand, bool) {
	if i < nargs {
		return *args[i], true
	}
	if v.defExpr == nil {
		if v.defVal != nil && v.defVal.Kind() != constant.Unknown {
			var d operand
			d.mode_ = constant_
			d.typ_ = v.typ
			d.val = v.defVal
			return d, true
		}
		return operand{}, false
	}
	var d operand
	check.rawExpr(nil, &d, v.defExpr, v.typ, false)
	if !d.isValid() {
		return operand{}, false
	}
	return d, true
}
