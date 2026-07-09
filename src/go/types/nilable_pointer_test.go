// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func checkNilablePointers(t *testing.T, src, modMode string) error {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := types.NewPackage("p", "p")
	pkg.SetNilablePointers(modMode)
	return types.NewChecker(&types.Config{}, fset, pkg, nil).Files([]*ast.File{f})
}

func TestNilablePointerNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a *int?) {
	if a != nil {
		var b *int = a
		_ = b
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var p *int = nil
	_ = p
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to *int")
	}
}

func TestNilablePointerDisabledEquivalence(t *testing.T) {
	const src = `package p
func f(a *int?, b *int) {
	var x *int = a
	var y *int? = b
	_ = x
	_ = y
}`
	if err := checkNilablePointers(t, src, "disable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerRegionEnd(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func enabled() {
	var p *int = nil
	_ = p
}
//go:nilable_pointers end
func disabled(a *int?, b *int) {
	var x *int = a
	var y *int? = b
	_ = x
	_ = y
}`
	if err := checkNilablePointers(t, src, "disable"); err == nil {
		t.Fatal("expected error in enabled region")
	}
}

func TestNilablePointerModDefault(t *testing.T) {
	const src = `package p
func f() {
	var p *int = nil
	_ = p
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error with go.mod enable default")
	}
}
