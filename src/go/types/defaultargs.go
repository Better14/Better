
package types

import (
	"go/ast"
	. "internal/types/errors"
)

func (check *Checker) paramDefault(v *Var, expr ast.Expr) {
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
			check.errorf(atPos(p.pos), InvalidSyntaxTree, "missing default value for parameter %q", p.name)
		}
	}
}

func sigMinParams(sig *Signature) int {
	n := sig.params.Len()
	min := 0
	for i := 0; i < n; i++ {
		if sig.params.vars[i].defExpr == nil {
			min++
		} else {
			break
		}
	}
	return min
}

func (check *Checker) appendDefaultArgs(call *ast.CallExpr, sig *Signature, args []*operand) []*operand {
	if sig.variadic {
		return args
	}
	npars := sig.params.Len()
	nargs := len(args)
	if nargs >= npars {
		return args
	}
	out := append([]*operand(nil), args...)
	var extras []ast.Expr
	for i := nargs; i < npars; i++ {
		v := sig.params.vars[i]
		if v.defExpr == nil {
			break
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
		call.Args = append(call.Args, extras...)
		return out
	}
	return args
}

func (check *Checker) overloadArgOperand(args []*operand, nargs, i int, v *Var) (operand, bool) {
	if i < nargs {
		return *args[i], true
	}
	if v.defExpr == nil {
		return operand{}, false
	}
	var d operand
	check.rawExpr(nil, &d, v.defExpr, v.typ, false)
	if !d.isValid() {
		return operand{}, false
	}
	return d, true
}
