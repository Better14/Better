
package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	"strconv"
	. "internal/types/errors"
)

func isNegativeIndexExpr(e syntax.Expr) bool {
	op, ok := syntax.Unparen(e).(*syntax.Operation)
	return ok && op.Y == nil && op.Op == syntax.Sub
}

func (check *Checker) intLit(pos syntax.Pos, v int64) *syntax.BasicLit {
	lit := new(syntax.BasicLit)
	lit.SetPos(pos)
	lit.Kind = syntax.IntLit
	lit.Value = strconv.FormatInt(v, 10)
	return lit
}

func (check *Checker) offsetFromLen(x, neg syntax.Expr) syntax.Expr {
	pos := neg.Pos()
	call := &syntax.CallExpr{
		Fun:     syntax.NewName(pos, "len"),
		ArgList: []syntax.Expr{x},
	}
	call.SetPos(pos)
	op := &syntax.Operation{
		Op: syntax.Add,
		X:  call,
		Y:  neg,
	}
	op.SetPos(pos)
	return op
}

func (check *Checker) resolveNegativeSliceIndices(e *syntax.SliceExpr, length int64) {
	for i := range e.Index {
		expr := e.Index[i]
		if expr == nil {
			continue
		}
		var x operand
		check.expr(nil, &x, expr)
		if !x.isValid() {
			continue
		}
		if x.mode() == constant_ {
			v, ok := constant.Int64Val(x.val)
			if !ok {
				continue
			}
			if v < 0 {
				if length < 0 {
					e.Index[i] = check.offsetFromLen(e.X, expr)
					continue
				}
				resolved := length + v
				if resolved < 0 {
					check.errorf(expr, InvalidIndex, invalidArg+"slice index %s out of bounds", x.val.String())
					continue
				}
				e.Index[i] = check.intLit(expr.Pos(), resolved)
			}
			continue
		}
		if isNegativeIndexExpr(expr) {
			e.Index[i] = check.offsetFromLen(e.X, expr)
		}
	}
}
