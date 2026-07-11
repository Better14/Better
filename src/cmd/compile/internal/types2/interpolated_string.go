
package types2

import (
	"cmd/compile/internal/syntax"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isQuotedStringLit(lit *syntax.BasicLit) bool {
	return lit != nil && lit.Kind == syntax.StringLit && len(lit.Value) >= 2 &&
		lit.Value[0] == '"' && lit.Value[len(lit.Value)-1] == '"'
}

func isRawStringLit(lit *syntax.BasicLit) bool {
	return lit != nil && lit.Kind == syntax.StringLit && len(lit.Value) >= 2 &&
		lit.Value[0] == '`' && lit.Value[len(lit.Value)-1] == '`'
}

func quotedStringHasInterpolation(quoted string) bool {
	_, holes, ok := parseQuotedInterpolation(quoted)
	return ok && len(holes) > 0
}

type interpHole struct {
	exprSrc string
	format  string // printf verb without leading %, empty means v
}

func parseQuotedInterpolation(quoted string) (format string, holes []interpHole, ok bool) {
	if len(quoted) < 2 || quoted[0] != '"' || quoted[len(quoted)-1] != '"' {
		return "", nil, false
	}
	body := quoted[1 : len(quoted)-1]
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		ch := body[i]
		if ch == '\\' && i+1 < len(body) {
			next := body[i+1]
			if next == '{' || next == '}' {
				b.WriteByte(next)
				i++
				continue
			}
			b.WriteByte('\\')
			b.WriteByte(next)
			i++
			continue
		}
		if ch != '{' {
			b.WriteByte(ch)
			continue
		}
		// ${VAR} shell-style expansions are literal braces, not holes.
		if i > 0 && body[i-1] == '$' {
			b.WriteByte(ch)
			continue
		}
		// {{template}} patterns are literal braces, not holes.
		if i > 0 && body[i-1] == '{' {
			b.WriteByte(ch)
			continue
		}
		close := strings.IndexByte(body[i+1:], '}')
		if close < 0 {
			b.WriteByte(ch)
			continue
		}
		inside := body[i+1 : i+1+close]
		// fmt.Sprintf patterns like "{PkgName:%v, DeclList:%v}" and interface
		// type sets like "{int|string}" use braces literally.
		if strings.ContainsAny(inside, ",;|{") {
			b.WriteByte(ch)
			continue
		}
		exprSrc, fmtSpec := splitInterpSpec(inside)
		trimmed := strings.TrimSpace(exprSrc)
		if trimmed == "" {
			b.WriteByte(ch)
			continue
		}
		// Documentation patterns like v8.{0-9} are literal braces, not holes.
		if r, _ := utf8.DecodeRuneInString(trimmed); r != '_' && r != '(' && !unicode.IsLetter(r) {
			b.WriteByte(ch)
			continue
		}
		if exprSrc == "" {
			return "", nil, false
		}
		b.WriteString("%")
		if fmtSpec == "" {
			b.WriteString("v")
		} else if strings.HasPrefix(fmtSpec, "%") {
			b.WriteString(fmtSpec[1:])
		} else if strings.HasPrefix(fmtSpec, ".") || strings.HasPrefix(fmtSpec, "#") || strings.HasPrefix(fmtSpec, "+") {
			b.WriteString(fmtSpec)
		} else {
			b.WriteString(fmtSpec)
		}
		holes = append(holes, interpHole{exprSrc: strings.TrimSpace(exprSrc), format: fmtSpec})
		i += close + 1
	}
	return b.String(), holes, true
}

func splitInterpSpec(inside string) (expr, format string) {
	colon := strings.IndexByte(inside, ':')
	if colon < 0 {
		return strings.TrimSpace(inside), ""
	}
	return strings.TrimSpace(inside[:colon]), strings.TrimSpace(inside[colon+1:])
}

func isInterpFormatSpec(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, ",;%") {
		return false
	}
	return true
}

func (check *Checker) holeLooksLikeInterpolation(hole interpHole, lit *syntax.BasicLit) bool {
	if isInterpFormatSpec(hole.format) {
		return true
	}
	exprSrc := hole.exprSrc
	if strings.ContainsAny(exprSrc, ".(,)[}]\"'`+-*/%&|^<>=!") {
		return true
	}
	expr, err := check.parseInterpHoleExpr(lit.Pos(), exprSrc)
	if err != nil {
		return false
	}
	name, ok := expr.(*syntax.Name)
	if !ok {
		return true
	}
	if !isValidName(name.Value) {
		return false
	}
	_, obj := check.lookupScope(name.Value)
	if obj == nil {
		obj = check.lookupPkgEnumVariant(name.Value)
	}
	if obj == nil {
		return false
	}
	switch obj.(type) {
	case *Var, *Const:
		return true
	default:
		return false
	}
}

func (check *Checker) lowerInterpolatedString(lit *syntax.BasicLit) syntax.Expr {
	format, holes, ok := parseQuotedInterpolation(lit.Value)
	if !ok || len(holes) == 0 {
		return lit
	}
	for _, hole := range holes {
		if !check.holeLooksLikeInterpolation(hole, lit) {
			return lit
		}
	}
	args := make([]syntax.Expr, 0, len(holes)+1)
	args = append(args, check.stringLit(lit.Pos(), format))
	for _, hole := range holes {
		expr, err := check.parseInterpHoleExpr(lit.Pos(), hole.exprSrc)
		if err != nil {
			return lit
		}
		var x operand
		check.exprOrType(&x, expr, false)
		if !x.isValid() || x.mode() == typexpr {
			return lit
		}
		args = append(args, expr)
	}
	return check.makeFmtSprintfCall(lit.Pos(), args)
}

func (check *Checker) stringLit(pos syntax.Pos, val string) *syntax.BasicLit {
	lit := new(syntax.BasicLit)
	lit.SetPos(pos)
	lit.Kind = syntax.StringLit
	lit.Value = strconv.Quote(val)
	return lit
}

func (check *Checker) makeFmtSprintfCall(pos syntax.Pos, args []syntax.Expr) *syntax.CallExpr {
	check.importPackage(pos, "fmt", "")
	sel := &syntax.SelectorExpr{
		X:   syntax.NewName(pos, "fmt"),
		Sel: syntax.NewName(pos, "Sprintf"),
	}
	sel.SetPos(pos)
	call := &syntax.CallExpr{Fun: sel, ArgList: args}
	call.SetPos(pos)
	return call
}

func (check *Checker) parseInterpHoleExpr(pos syntax.Pos, exprSrc string) (syntax.Expr, error) {
	src := fmt.Sprintf("package p; func _() { _ = %s }", exprSrc)
	base := syntax.NewFileBase("<interpolation>")
	file, err := syntax.Parse(base, strings.NewReader(src), nil, nil, 0)
	if err != nil {
		return nil, err
	}
	if len(file.DeclList) == 0 {
		return nil, fmt.Errorf("missing declaration")
	}
	fn, ok := file.DeclList[0].(*syntax.FuncDecl)
	if !ok || fn.Body == nil || len(fn.Body.List) == 0 {
		return nil, fmt.Errorf("missing body")
	}
	stmt, ok := fn.Body.List[0].(*syntax.AssignStmt)
	if !ok || stmt.Rhs == nil {
		return nil, fmt.Errorf("missing assignment")
	}
	expr := stmt.Rhs
	expr.SetPos(pos)
	return expr, nil
}

func (check *Checker) maybeLowerInterpolatedString(x *operand, e *syntax.BasicLit) bool {
	if e.Bad || !isQuotedStringLit(e) || isRawStringLit(e) || !quotedStringHasInterpolation(e.Value) {
		return false
	}
	lowered := check.lowerInterpolatedString(e)
	if lowered == e {
		return false
	}
	check.expr(nil, x, lowered)
	return true
}
