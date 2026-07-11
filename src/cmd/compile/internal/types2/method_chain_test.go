
package types2_test

import (
	"strings"
	"testing"
	"time"
	"cmd/compile/internal/syntax"
	. "cmd/compile/internal/types2"
)

// TestLongMethodChain checks that nested selector calls type-check without
// re-evaluating receiver chains (see callExpr typedOperand).
func TestLongMethodChain(t *testing.T) {
	const n = 40
	var b strings.Builder
	b.WriteString("package p\n\ntype U struct { Children []J }\n")
	b.WriteString("func (u *U) Child(n int) J { return u.Children[n] }\n")
	b.WriteString("type J interface { Child(n int) J }\n")
	b.WriteString("func F(u *U) *U { return u")
	for i := 0; i < n; i++ {
		b.WriteString(".Child(0)")
	}
	b.WriteString(".(*U) }\n")

	src := b.String()
	file, err := syntax.Parse(syntax.NewFileBase("chain.go"), strings.NewReader(src), nil, nil, syntax.CheckBranches)
	if err != nil {
		t.Fatal(err)
	}

	conf := Config{Importer: defaultImporter()}
	info := &Info{Types: make(map[syntax.Expr]TypeAndValue)}
	start := time.Now()
	_, err = conf.Check("p", []*syntax.File{file}, info)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("type-check took %v; expected well under 5s (receiver chain re-eval bug?)", d)
	}
}
