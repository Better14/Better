package parser

import (
	"go/ast"
	"go/token"
)

func (p *parser) atTypeStart() bool {
	switch p.tok {
	case token.IDENT, token.MUL, token.LBRACK, token.CHAN, token.MAP, token.STRUCT, token.INTERFACE, token.FUNC, token.LPAREN, token.ARROW:
		return true
	default:
		return false
	}
}

func (p *parser) parseLbrackOperand() ast.Expr {
	if p.trace {
		defer un(trace(p, "LbrackOperand"))
	}

	lbrack := p.expect(token.LBRACK)
	if p.tok == token.ELLIPSIS {
		ellipsis := p.pos
		p.next()
		p.expect(token.RBRACK)
		elem := p.tryIdentOrType()
		if elem == nil {
			pos := p.pos
			p.errorExpected(pos, "type")
			elem = &ast.BadExpr{From: pos, To: pos}
		}
		return &ast.ArrayType{Lbrack: lbrack, Len: &ast.Ellipsis{Ellipsis: ellipsis}, Elt: elem}
	}
	if p.tok == token.RBRACK {
		rbrack := p.pos
		p.next()
		if p.atTypeStart() {
			elem := p.tryIdentOrType()
			if elem == nil {
				pos := p.pos
				p.errorExpected(pos, "type")
				elem = &ast.BadExpr{From: pos, To: pos}
			}
			return &ast.ArrayType{Lbrack: lbrack, Elt: elem}
		}
		return &ast.CompositeLit{Lbrace: lbrack, Rbrace: rbrack, Shorthand: ast.ShorthandArray}
	}

	p.exprLev++
	first := p.parseShorthandLitElem()
	if p.tok == token.COMMA {
		p.next()
		elts := []ast.Expr{first}
		for p.tok != token.RBRACK && p.tok != token.EOF {
			elts = append(elts, p.parseShorthandLitElem())
			if !p.atComma("array literal", token.RBRACK) {
				break
			}
			p.next()
		}
		p.exprLev--
		rbrack := p.expectClosing(token.RBRACK, "array literal")
		return &ast.CompositeLit{Lbrace: lbrack, Elts: elts, Rbrace: rbrack, Shorthand: ast.ShorthandArray}
	}
	if p.tok == token.RBRACK {
		rbrack := p.pos
		p.next()
		p.exprLev--
		if p.atTypeStart() {
			elem := p.tryIdentOrType()
			if elem == nil {
				pos := p.pos
				p.errorExpected(pos, "type")
				elem = &ast.BadExpr{From: pos, To: pos}
			}
			return &ast.ArrayType{Lbrack: lbrack, Len: first, Elt: elem}
		}
		return &ast.CompositeLit{Lbrace: lbrack, Elts: []ast.Expr{first}, Rbrace: rbrack, Shorthand: ast.ShorthandArray}
	}
	p.exprLev--
	pos := p.pos
	p.errorExpected(pos, "]")
	p.advance(exprEnd)
	return &ast.BadExpr{From: lbrack, To: p.pos}
}

func (p *parser) parseLbraceOperand() ast.Expr {
	if p.trace {
		defer un(trace(p, "LbraceOperand"))
	}

	lbrace := p.pos
	p.next()
	if p.tok == token.RBRACE {
		rbrace := p.pos
		p.next()
		if p.atTypeStart() {
			elem := p.tryIdentOrType()
			if elem == nil {
				pos := p.pos
				p.errorExpected(pos, "type")
				elem = &ast.BadExpr{From: pos, To: pos}
			}
			return &ast.SetType{Lbrace: lbrace, Rbrace: rbrace, Elem: elem}
		}
		return &ast.CompositeLit{Lbrace: lbrace, Rbrace: rbrace, Shorthand: ast.ShorthandMap}
	}
	return p.parseShorthandComplitBody(lbrace)
}

func (p *parser) parseShorthandComplitBody(lbrace token.Pos) *ast.CompositeLit {
	if p.trace {
		defer un(trace(p, "ShorthandComplitBody"))
	}

	var elts []ast.Expr
	nkeys := 0
	p.exprLev++
	for p.tok != token.RBRACE && p.tok != token.EOF {
		e := p.parseShorthandLitElem()
		if p.tok == token.COLON {
			colon := p.pos
			p.next()
			e = &ast.KeyValueExpr{Key: e, Colon: colon, Value: p.parseShorthandLitElem()}
			nkeys++
		}
		elts = append(elts, e)
		if !p.atComma("composite literal", token.RBRACE) {
			break
		}
		p.next()
	}
	p.exprLev--
	rbrace := p.expectClosing(token.RBRACE, "composite literal")
	shorthand := ast.ShorthandSet
	if nkeys > 0 {
		shorthand = ast.ShorthandMap
	}
	return &ast.CompositeLit{Lbrace: lbrace, Elts: elts, Rbrace: rbrace, Shorthand: shorthand}
}

func (p *parser) parseShorthandLitElem() ast.Expr {
	if p.tok == token.ELLIPSIS {
		ellipsis := p.pos
		p.next()
		return &ast.SpreadExpr{Ellipsis: ellipsis, X: p.parseBareComplitValue()}
	}
	return p.parseBareComplitValue()
}

func (p *parser) parseBareComplitValue() ast.Expr {
	switch p.tok {
	case token.LBRACE:
		return p.parseLbraceOperand()
	case token.LBRACK:
		return p.parseLbrackOperand()
	default:
		return p.parseRhs()
	}
}
