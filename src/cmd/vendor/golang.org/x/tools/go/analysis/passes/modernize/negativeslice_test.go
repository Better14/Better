
package modernize_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/passes/modernize"
)

func TestNegativeSlice(t *testing.T) {
	testFixWithoutFormat(t, modernize.NegativeSliceAnalyzer, "negativeslice")
}
