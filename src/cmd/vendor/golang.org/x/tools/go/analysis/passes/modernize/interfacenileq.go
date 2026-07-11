
package modernize

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

const interfaceNilEqFixme = "//FIXME: Make sure still works after interface == nil change."

var InterfaceNilEqAnalyzer = &analysis.Analyzer{
	Name: "interfacenileq",
	Doc:  analyzerutil.MustExtractDoc(doc, "interfacenileq"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#interfacenileq",
	Run:  interfaceNilEq,
}

func interfaceNilEq(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	for _, f := range pass.Files {
		seen := map[string]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			be, ok := n.(*ast.BinaryExpr)
			if !ok || (be.Op != token.EQL && be.Op != token.NEQ) {
				return true
			}
			var iface ast.Expr
			switch {
			case isNilExpr(pass, be.Y):
				iface = be.X
			case isNilExpr(pass, be.X):
				iface = be.Y
			default:
				return true
			}
			tv, ok := pass.TypesInfo.Types[iface]
			if !ok || tv.Type == nil {
				return true
			}
			if _, ok := tv.Type.Underlying().(*types.Interface); !ok {
				return true
			}
			pos := pass.Fset.Position(be.Pos())
			if seen[pos.String()] {
				return true
			}
			seen[pos.String()] = true
			lineStart := pass.Fset.File(f.Pos()).Offset(pass.Fset.File(f.Pos()).LineStart(pos.Line))
			pass.Report(analysis.Diagnostic{
				Pos:     pass.Fset.File(f.Pos()).Pos(lineStart),
				Message: "interface == nil comparison may change behavior after typed-nil fix",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Add review FIXME comment",
					TextEdits: []analysis.TextEdit{{
						Pos:     pass.Fset.File(f.Pos()).Pos(lineStart),
						NewText: []byte(interfaceNilEqFixme + "\n"),
					}},
				}},
			})
			return true
		})
	}
	return nil, nil
}

func isNilExpr(pass *analysis.Pass, e ast.Expr) bool {
	e = ast.Unparen(e)
	if id, ok := e.(*ast.Ident); ok && id.Name == "nil" {
		return true
	}
	tv, ok := pass.TypesInfo.Types[e]
	return ok && tv.IsNil()
}
