// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"testing"

	"cmd/compile/internal/types2"
)

func TestResultShortVarDecl(t *testing.T) {
	const src = `
package p
func f() {
	var a int! = 0
	val, err := a
	_, _ = val, err
}
`
	_, err := typecheck(src, nil, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
}

func TestResultReturnVarType(t *testing.T) {
	const src = `
package p
func f() int! {
	var a int! = 0
	return a
}
`
	pkg, err := typecheck(src, nil, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
	// Find return statement's 'a' identifier type via syntax walk would be heavy;
	// instead verify multiExpr would see Result on the variable object.
	a := pkg.Scope().Lookup("f").(*types2.Func).Scope().Lookup("a")
	if a == nil {
		t.Fatal("missing var a")
	}
	if _, ok := types2.AsResult(a.Type()); !ok {
		t.Fatalf("var a type = %v, want Result", a.Type())
	}
}

func TestResultReturnQuery(t *testing.T) {
	const src = `
package p
func f() int! {
	var a int! = 0
	return a
}
`
	pkg, err := typecheck(src, nil, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
	obj := pkg.Scope().Lookup("f")
	sig := obj.Type().(*types2.Signature)
	if !sig.ResultQuery() {
		t.Fatal("expected ResultQuery for int! function")
	}
	if sig.Results().Len() != 2 {
		t.Fatalf("expected 2 results, got %d", sig.Results().Len())
	}
}

func TestResultAssign(t *testing.T) {
	const src = `
package p
func f() {
	var a int! = 0
	var val int
	var err error
	val, err = a
	_, _ = val, err
}
`
	_, err := typecheck(src, nil, nil)
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
}
