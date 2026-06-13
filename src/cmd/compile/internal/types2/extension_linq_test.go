// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"fmt"
	"strings"
	"testing"

	. "cmd/compile/internal/types2"
)

func TestJoinTypeParamConstraint(t *testing.T) {
	imp := defaultImporter().(*gcimports)
	pkg, err := imp.Import("linq")
	if err != nil {
		t.Fatal(err)
	}
	join := pkg.Scope().Lookup("Join").(*Func)
	sig := join.Type().(*Signature)
	T := sig.TypeParams().At(0)
	U := sig.TypeParams().At(1)
	fmt.Printf("Join T constraint=%s\n", T.Constraint())
	fmt.Printf("Join U constraint=%s\n", U.Constraint())
	p0 := sig.Params().At(0).Type()
	fmt.Printf("Join param0=%s\n", p0)
}

func TestLinqExtensionCalls(t *testing.T) {
	src := `package main
import (
	"linq"
	"slices"
)

func f() {
	nums := []int{1, 2, 3}
	_ = linq.From(nums).Where(func(n int) bool { return n < 5 })
	_ = linq.From(nums).Sum()
	_ = linq.From(nums).Max()
	inner := linq.From([]int{2})
	_ = linq.From(nums).Join(inner, func(int) int { return 0 }, func(int) int { return 0 }, func(int, int) int { return 0 })
	_, _ = nums.TryGetSeqLen()
	_ = slices.Values(nums)
}
`
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  defaultImporter(),
		Error: func(err error) {
			t.Error(err)
		},
	}
	_, err := typecheck(src, conf, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
}

func TestLinqExtensionElemMatch(t *testing.T) {
	// Regression: constrained type params (Number, cmp.Ordered) must match concrete elems.
	src := `package main
import ("cmp"; "linq")

func f() {
	var _ float64 = linq.From([]int{1}).Average()
	var _ int = linq.From([]int{1}).Min()
	_ = linq.From([]string{"a"}).Max()
	_ = cmp.Compare(0, 0)
}
`
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  defaultImporter(),
		Error: func(err error) {
			errs = append(errs, err.Error())
		},
	}
	_, _ = typecheck(src, conf, nil)
	for _, e := range errs {
		if strings.Contains(e, "undefined") && strings.Contains(e, "iter.Seq") {
			t.Errorf("unexpected extension error: %s", e)
		}
	}
}
