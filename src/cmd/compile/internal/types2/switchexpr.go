// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	. "internal/types/errors"
)

func (check *Checker) switchExpr(x *operand, e *syntax.SwitchExpr) {
	var tag operand
	if e.Tag != nil {
		check.expr(nil, &tag, e.Tag)
		check.assignment(&tag, nil, "switch expression")
		if tag.isValid() {
			if _, ok := AsEnum(tag.typ()); !ok && !Comparable(tag.typ()) && !hasNil(tag.typ()) {
				check.errorf(&tag, InvalidExprSwitch, "cannot switch on %s (%s is not comparable)", &tag, tag.typ())
				tag.invalidate()
			}
		}
	} else {
		tag.mode_ = constant_
		tag.typ_ = Typ[Bool]
		tag.val = constant.MakeBool(true)
		pos := e.Rbrace
		if len(e.Body) > 0 {
			pos = e.Body[0].Pos()
		}
		tag.expr = syntax.NewName(pos, "true")
	}
	if !tag.isValid() {
		x.invalidate()
		return
	}

	if enumTyp, ok := AsEnum(tag.typ()); ok {
		check.enumSwitchExpr(x, e, tag.typ(), enumTyp)
		return
	}

	check.multipleSwitchDefaults(switchExprClausesToCaseClauses(e.Body))

	seen := make(valueMap)
	var arms []*operand
	hasDefault := false
	for _, clause := range e.Body {
		if clause == nil {
			continue
		}
		if clause.Cases == nil {
			hasDefault = true
		} else {
			check.caseValues(&tag, syntax.UnpackListExpr(clause.Cases), seen)
		}
		var arm operand
		check.expr(nil, &arm, clause.Body)
		if !arm.isValid() {
			x.invalidate()
			return
		}
		arms = append(arms, &arm)
	}

	if len(arms) == 0 {
		check.errorf(e, InvalidSyntaxTree, "switch expression must have at least one case")
		x.invalidate()
		return
	}
	if !hasDefault {
		check.errorf(e, InvalidSyntaxTree, "switch expression requires default case")
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

// switchExprClausesToCaseClauses adapts switch expression clauses for default checking.
func switchExprClausesToCaseClauses(body []*syntax.SwitchExprClause) []*syntax.CaseClause {
	out := make([]*syntax.CaseClause, len(body))
	for i, c := range body {
		if c == nil {
			continue
		}
		out[i] = &syntax.CaseClause{Cases: c.Cases, Colon: c.Colon}
	}
	return out
}
