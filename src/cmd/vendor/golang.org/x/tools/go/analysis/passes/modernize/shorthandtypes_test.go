
package modernize_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/modernize"
)

func TestShorthandTypes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, modernize.ShorthandTypesAnalyzer, "shorthandtypes")
}
