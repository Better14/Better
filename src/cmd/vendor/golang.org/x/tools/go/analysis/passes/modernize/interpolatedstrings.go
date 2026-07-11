// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/internal/analysis/analyzerutil"
)

var InterpolatedStringsAnalyzer = &analysis.Analyzer{
	Name: "interpolatedstrings",
	Doc:  analyzerutil.MustExtractDoc(doc, "interpolatedstrings"),
	URL:  "https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize#interpolatedstrings",
	Run:  interpolatedStrings,
}

func interpolatedStrings(pass *analysis.Pass) (any, error) {
	if pkgInGOROOT(pass) {
		return nil, nil
	}
	for _, f := range pass.Files {
		var edits []textEdit
		var n int
		edits, n = rewriteSprintfToInterp(pass.Fset, f, edits)
		reportInterpEdits(pass, f, edits[len(edits)-n:], "fmt.Sprintf can use interpolated string syntax", "Use interpolated string")
		edits, n = rewriteConcatToInterp(pass.Fset, f, edits)
		reportInterpEdits(pass, f, edits[len(edits)-n:], "string concatenation can use interpolated string syntax", "Use interpolated string")
		edits, n = escapeLiteralBraces(pass.Fset, f, edits)
		reportInterpEdits(pass, f, edits[len(edits)-n:], "literal braces must be escaped in interpolated strings", "Escape braces")
	}
	return nil, nil
}

func reportInterpEdits(pass *analysis.Pass, f *ast.File, edits []textEdit, message, fix string) {
	if len(f.Decls) == 0 {
		return
	}
	file := pass.Fset.File(f.Pos())
	for _, e := range edits {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Pos(e.start),
			End:     file.Pos(e.end),
			Message: message,
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: fix,
				TextEdits: []analysis.TextEdit{{
					Pos:     file.Pos(e.start),
					End:     file.Pos(e.end),
					NewText: e.text,
				}},
			}},
		})
	}
}
