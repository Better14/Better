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
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.GenDecl)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		gdecl := n.(*ast.GenDecl)
		if gdecl.Tok != token.TYPE || gdecl.Lparen.IsValid() {
			return
		}
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
				kwPos   token.Pos
			)
			switch t := ts.Type.(type) {
			case *ast.StructType:
				keyword = token.STRUCT
				kwPos = t.Struct
			case *ast.InterfaceType:
				keyword = token.INTERFACE
				kwPos = t.Interface
			default:
				continue
			}
			if !kwPos.IsValid() {
				continue
			}
			pass.Report(analysis.Diagnostic{
				Pos:     gdecl.TokPos,
				End:     kwPos + token.Pos(len(keyword.String())),
				Message: "type declaration can use struct/interface shorthand syntax",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Use shorthand syntax",
					TextEdits: []analysis.TextEdit{{
						Pos:     gdecl.TokPos,
						End:     kwPos + token.Pos(len(keyword.String())),
						NewText: []byte(keyword.String() + " " + ts.Name.Name),
					}},
				}},
			})
		}
	})
	return nil, nil
}
