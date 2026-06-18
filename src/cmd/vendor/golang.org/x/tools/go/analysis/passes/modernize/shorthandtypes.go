// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

var ShorthandTypesAnalyzer = &analysis.Analyzer{
	Name: "shorthandtypes",
	Doc:  analyzerutil.MustExtractDoc(doc, "shorthandtypes"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#shorthandtypes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: shorthandTypes,
}

func shorthandTypes(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	type candidate struct {
		pos     token.Pos
		end     token.Pos
		newText string
	}

	nodeFilter := []ast.Node{(*ast.GenDecl)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		gdecl := n.(*ast.GenDecl)
		if gdecl.Tok != token.TYPE || gdecl.Lparen.IsValid() {
			return
		}
		var cands []candidate
		for _, spec := range gdecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Assign.IsValid() || ts.TypeParams != nil {
				continue
			}
			var (
				keyword token.Token
				opening token.Pos
			)
			switch t := ts.Type.(type) {
			case *ast.StructType:
				keyword = token.STRUCT
				if t.Fields != nil {
					opening = t.Fields.Opening
				}
			case *ast.InterfaceType:
				keyword = token.INTERFACE
				if t.Methods != nil {
					opening = t.Methods.Opening
				}
			default:
				continue
			}
			if !opening.IsValid() {
				continue
			}
			cands = append(cands, candidate{
				pos:     gdecl.TokPos,
				end:     opening,
				newText: keyword.String() + " " + ts.Name.Name + " ",
			})
		}
		for _, c := range cands {
			pass.Report(analysis.Diagnostic{
				Pos:     c.pos,
				End:     c.end,
				Message: "type declaration can use struct/interface shorthand syntax",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Use shorthand syntax",
					TextEdits: []analysis.TextEdit{{
						Pos:     c.pos,
						End:     c.end,
						NewText: []byte(c.newText),
					}},
				}},
			})
		}
	})
	return nil, nil
}
