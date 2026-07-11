
package modernize_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/passes/modernize"
)

func TestSpreadCall(t *testing.T) {
	testFixWithoutFormat(t, modernize.SpreadCallAnalyzer, "spreadcall")
}
