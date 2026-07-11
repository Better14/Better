
package modernize_test

import (
	"testing"

	. "golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/modernize"
)

func TestForIn(t *testing.T) {
	RunWithSuggestedFixes(t, TestData(), modernize.ForInAnalyzer, "forin")
}
