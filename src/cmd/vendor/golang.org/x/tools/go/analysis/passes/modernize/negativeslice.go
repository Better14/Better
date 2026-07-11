// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

var NegativeSliceAnalyzer = &analysis.Analyzer{
	Name: "negativeslice",
	Doc:  analyzerutil.MustExtractDoc(doc, "negativeslice"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#negativeslice",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: negativeSlice,
}

func negativeSlice(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.SliceExpr)(nil)}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		se := n.(*ast.SliceExpr)
		var edits []analysis.TextEdit
		if low, ok := lenMinusBound(pass, se.X, se.Low); ok {
			edits = append(edits, analysis.TextEdit{
				Pos:     se.Low.Pos(),
				End:     se.Low.End(),
				NewText: []byte(low),
			})
		}
		if high, omit, ok := lenHighBound(pass, se.X, se.High); ok {
			if omit {
				edits = append(edits, analysis.TextEdit{
					Pos:     se.High.Pos(),
					End:     se.High.End(),
					NewText: nil,
				})
			} else {
				edits = append(edits, analysis.TextEdit{
					Pos:     se.High.Pos(),
					End:     se.High.End(),
					NewText: []byte(high),
				})
			}
		}
		if omit, ok := omitZeroLowBound(pass, se); ok && len(edits) > 0 {
			edits = append(edits, omit)
		}
		if len(edits) == 0 {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:     se.Lbrack,
			End:     se.Rbrack + 1,
			Message: "slice bounds can use negative index syntax",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message:   "Use negative slice indices",
				TextEdits: edits,
			}},
		})
	})
	return nil, nil
}

func omitZeroLowBound(pass *analysis.Pass, se *ast.SliceExpr) (analysis.TextEdit, bool) {
	if se.Low == nil || !isZeroIntConst(pass.TypesInfo, se.Low) {
		return analysis.TextEdit{}, false
	}
	return analysis.TextEdit{
		Pos:     se.Low.Pos(),
		End:     se.Low.End(),
		NewText: nil,
	}, true
}

func lenMinusBound(pass *analysis.Pass, slice ast.Expr, bound ast.Expr) (string, bool) {
	if bound == nil {
		return "", false
	}
	be, ok := ast.Unparen(bound).(*ast.BinaryExpr)
	if !ok || be.Op != token.SUB {
		return "", false
	}
	if !isLenOf(pass, be.X, slice) {
		return "", false
	}
	return negativeLiteral(be.Y)
}

func lenHighBound(pass *analysis.Pass, slice ast.Expr, bound ast.Expr) (string, bool, bool) {
	if bound == nil {
		return "", false, false
	}
	if isLenOf(pass, bound, slice) {
		return "", true, true
	}
	be, ok := ast.Unparen(bound).(*ast.BinaryExpr)
	if !ok || be.Op != token.SUB {
		return "", false, false
	}
	if !isLenOf(pass, be.X, slice) {
		return "", false, false
	}
	neg, ok := negativeLiteral(be.Y)
	return neg, false, ok
}

func isLenOf(pass *analysis.Pass, call ast.Expr, slice ast.Expr) bool {
	ce, ok := ast.Unparen(call).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn, ok := ce.Fun.(*ast.Ident)
	if !ok || fn.Name != "len" || len(ce.Args) != 1 {
		return false
	}
	return exprEqual(pass, ce.Args[0], slice)
}

func exprEqual(pass *analysis.Pass, a, b ast.Expr) bool {
	var ba, bb bytes.Buffer
	if err := format.Node(&ba, pass.Fset, a); err != nil {
		return false
	}
	if err := format.Node(&bb, pass.Fset, b); err != nil {
		return false
	}
	return ba.String() == bb.String()
}

func negativeLiteral(e ast.Expr) (string, bool) {
	bl, ok := ast.Unparen(e).(*ast.BasicLit)
	if !ok || bl.Kind != token.INT {
		return "", false
	}
	if len(bl.Value) > 0 && bl.Value[0] == '-' {
		return "", false
	}
	return "-" + bl.Value, true
}
