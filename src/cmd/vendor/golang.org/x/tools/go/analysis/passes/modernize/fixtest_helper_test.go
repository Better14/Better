
package modernize_test

import (
	"bytes"
	"os"
	"testing"

	"golang.org/x/tools/go/analysis"
	. "golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/internal/diff"
)

func testFixWithoutFormat(t *testing.T, a *analysis.Analyzer, pkg string) {
	t.Helper()
	dir := TestData()
	results := Run(t, dir, a, pkg)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	result := results[0]
	filename := result.Pass.Fset.File(result.Pass.Files[0].Pos()).Name()
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filename + ".golden"
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}

	var edits []diff.Edit
	for _, d := range result.Diagnostics {
		if len(d.SuggestedFixes) == 0 {
			continue
		}
		fix := d.SuggestedFixes[0]
		for _, te := range fix.TextEdits {
			start := result.Pass.Fset.Position(te.Pos).Offset
			end := result.Pass.Fset.Position(te.End).Offset
			edits = append(edits, diff.Edit{Start: start, End: end, New: string(te.NewText)})
		}
	}
	if len(edits) == 0 {
		t.Fatal("no edits")
	}
	fixed, err := diff.ApplyBytes(content, edits)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(fixed), bytes.TrimSpace(want)) {
		t.Fatalf("got:\n%s\nwant:\n%s", fixed, want)
	}
}
