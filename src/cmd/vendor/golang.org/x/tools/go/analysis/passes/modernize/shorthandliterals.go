// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"bytes"
	"go/ast"
	"go/format"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

var ShorthandLiteralsAnalyzer = &analysis.Analyzer{
	Name: "shorthandliterals",
	Doc:  analyzerutil.MustExtractDoc(doc, "shorthandliterals"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#shorthandliterals",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: shorthandLiterals,
}

func shorthandLiterals(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CompositeLit)(nil), (*ast.CallExpr)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.CompositeLit:
			shorthandLitComposite(pass, n)
		case *ast.CallExpr:
			shorthandLitSetOf(pass, n)
		}
	})
	return nil, nil
}

func shorthandLitComposite(pass *analysis.Pass, lit *ast.CompositeLit) {
	if lit.Type == nil {
		return
	}
	switch t := lit.Type.(type) {
	case *ast.ArrayType:
		if t.Len != nil {
			return
		}
		if !lit.Lbrace.IsValid() || !lit.Rbrace.IsValid() {
			return
		}
		if !shorthandSliceRewriteOK(lit) {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:     lit.Type.Pos(),
			End:     lit.Lbrace + 1,
			Message: "slice composite literal can use array literal syntax",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "Use array literal syntax",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     lit.Type.Pos(),
						End:     lit.Lbrace + 1,
						NewText: []byte("["),
					},
					{
						Pos:     lit.Rbrace,
						End:     lit.Rbrace + 1,
						NewText: []byte("]"),
					},
				},
			}},
		})
	case *ast.MapType:
		if !lit.Lbrace.IsValid() {
			return
		}
		if !shorthandMapRewriteOK(lit) {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:     lit.Type.Pos(),
			End:     lit.Lbrace + 1,
			Message: "map composite literal can use dict literal syntax",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "Use dict literal syntax",
				TextEdits: []analysis.TextEdit{{
					Pos:     lit.Type.Pos(),
					End:     lit.Lbrace + 1,
					NewText: []byte("{"),
				}},
			}},
		})
	}
}

func shorthandLitSetOf(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !isSetPackageIdent(sel.X) || !isSetOfIdent(sel.Sel) {
		return
	}
	if !shorthandSetOfRewriteOK(call) {
		return
	}
	if len(call.Args) == 0 {
		pass.Report(analysis.Diagnostic{
			Pos:     call.Pos(),
			End:     call.End(),
			Message: "set.Of call can use set literal syntax",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "Use set literal syntax",
				TextEdits: []analysis.TextEdit{{
					Pos:     call.Pos(),
					End:     call.End(),
					NewText: []byte("{}"),
				}},
			}},
		})
		return
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, arg := range call.Args {
		if i > 0 {
			buf.WriteString(", ")
		}
		if err := format.Node(&buf, pass.Fset, arg); err != nil {
			return
		}
	}
	buf.WriteByte('}')
	pass.Report(analysis.Diagnostic{
		Pos:     call.Pos(),
		End:     call.End(),
		Message: "set.Of call can use set literal syntax",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "Use set literal syntax",
			TextEdits: []analysis.TextEdit{{
				Pos:     call.Pos(),
				End:     call.End(),
				NewText: buf.Bytes(),
			}},
		}},
	})
}

func isSetPackageIdent(x ast.Expr) bool {
	id, ok := ast.Unparen(x).(*ast.Ident)
	return ok && id.Name == "set"
}

func isSetOfIdent(id *ast.Ident) bool {
	return id != nil && id.Name == "Of"
}
