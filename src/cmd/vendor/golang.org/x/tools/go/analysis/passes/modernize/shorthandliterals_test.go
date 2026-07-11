
package modernize_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/passes/modernize"
)

func TestShorthandLiterals(t *testing.T) {
	testFixWithoutFormat(t, modernize.ShorthandLiteralsAnalyzer, "shorthandliterals")
}
