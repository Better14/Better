
package modernize

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

var ForInAnalyzer = &analysis.Analyzer{
	Name: "forin",
	Doc:  analyzerutil.MustExtractDoc(doc, "forin"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#forin",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: forIn,
}

func forIn(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.RangeStmt)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		rs := n.(*ast.RangeStmt)
		if rs.InPos.IsValid() || rs.Tok != token.DEFINE {
			return
		}
		rangePos := rs.Range
		if !rangePos.IsValid() {
			return
		}
		rangeEnd := rangePos + token.Pos(len("range"))
		if rs.Value == nil {
			name, ok := forInIdentName(rs.Key)
			if !ok {
				return
			}
			pass.Report(analysis.Diagnostic{
				Pos:     rs.For,
				End:     rs.Body.Pos(),
				Message: "range loop can use for-in syntax",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Use for-in syntax",
					TextEdits: []analysis.TextEdit{{
						Pos:     rs.Key.Pos(),
						End:     rangeEnd,
						NewText: []byte(name + ", _ in"),
					}},
				}},
			})
			return
		}
		if forInBlankValue(rs) {
			name, ok := forInIdentName(rs.Value)
			if !ok {
				return
			}
			pass.Report(analysis.Diagnostic{
				Pos:     rs.For,
				End:     rs.Body.Pos(),
				Message: "range loop can use for-in syntax",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Use for-in syntax",
					TextEdits: []analysis.TextEdit{{
						Pos:     rs.Key.Pos(),
						End:     rangeEnd,
						NewText: []byte(name + " in"),
					}},
				}},
			})
			return
		}
		if rs.Key != nil && !forInBlankIdent(rs.Key) {
			pass.Report(analysis.Diagnostic{
				Pos:     rs.For,
				End:     rs.Body.Pos(),
				Message: "range loop can use for-in syntax",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Use for-in syntax",
					TextEdits: []analysis.TextEdit{{
						Pos:     rs.Value.End(),
						End:     rangeEnd,
						NewText: []byte(" in"),
					}},
				}},
			})
		}
	})
	return nil, nil
}

func forInBlankValue(rs *ast.RangeStmt) bool {
	return rs.Value != nil && forInBlankIdent(rs.Key)
}

func forInBlankIdent(e ast.Expr) bool {
	id, ok := ast.Unparen(e).(*ast.Ident)
	return ok && id.Name == "_"
}

func forInIdentName(e ast.Expr) (string, bool) {
	id, ok := ast.Unparen(e).(*ast.Ident)
	if !ok {
		return "", false
	}
	return id.Name, true
}
