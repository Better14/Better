
package modernize

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

var SpreadCallAnalyzer = &analysis.Analyzer{
	Name: "spreadcall",
	Doc:  analyzerutil.MustExtractDoc(doc, "spreadcall"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#spreadcall",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: spreadCall,
}

func spreadCall(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !call.Ellipsis.IsValid() || len(call.Args) == 0 {
			return
		}
		last := call.Args[len(call.Args)-1]
		if !call.Ellipsis.IsValid() {
			return
		}
		var buf bytes.Buffer
		buf.WriteString("...")
		if err := format.Node(&buf, pass.Fset, last); err != nil {
			return
		}
		ellipsisEnd := call.Ellipsis + token.Pos(len("..."))
		pass.Report(analysis.Diagnostic{
			Pos:     last.Pos(),
			End:     ellipsisEnd,
			Message: "variadic call can use prefix spread syntax",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "Use prefix spread syntax",
				TextEdits: []analysis.TextEdit{{
					Pos:     last.Pos(),
					End:     ellipsisEnd,
					NewText: buf.Bytes(),
				}},
			}},
		})
	})
	return nil, nil
}
