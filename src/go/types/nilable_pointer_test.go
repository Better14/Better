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

func TestNilablePointerNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a *int?) {
	if a != nil {
		var b *int = a
		_ = b
	}
}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{NilablePointers: "enable"}
	pkg, err := conf.Check("p", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	_ = pkg
}

func TestNilablePointerNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var p *int = nil
	_ = p
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	conf := types.Config{NilablePointers: "enable"}
	_, err := conf.Check("p", fset, []*ast.File{f}, nil)
	if err == nil {
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
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "p.go", src, 0)
	conf := types.Config{NilablePointers: "disable"}
	_, err := conf.Check("p", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}
