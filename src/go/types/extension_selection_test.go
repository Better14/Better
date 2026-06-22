// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/ast"
	"go/importer"
	"go/token"
	"testing"

	. "go/types"
)

func TestExtensionSelectionRecording(t *testing.T) {
	const src = `package p

func (s []int) Double() []int {
	return s
}

func use(s []int) {
	_ = s.Double()
}
`
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	info := &Info{
		Uses: make(map[*ast.Ident]Object),
		Defs: make(map[*ast.Ident]Object),
	}
	conf := &Config{GoVersion: "go1.27"}
	if _, err := conf.Check("p", fset, []*ast.File{f}, info); err != nil {
		t.Fatal(err)
	}

	var doubleIdent *ast.Ident
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "Double" {
				doubleIdent = id
			}
		}
		return true
	})
	if doubleIdent == nil {
		t.Fatal("missing s.Double() call")
	}
	if obj := info.Uses[doubleIdent]; obj == nil || obj.Name() != "Double" {
		t.Fatalf("Uses[Double ident] = %v, want Double", obj)
	}
}

func TestExtensionSelectionRecordingImported(t *testing.T) {
	fset := token.NewFileSet()
	imports := make(testImporter)
	conf := Config{GoVersion: "go1.27", Importer: imports}

	const libSrc = `
package lib

func (s []int) Double() []int {
	return s
}
`
	const mainSrc = `
package main

import "lib"

func use(s []int) {
	_ = s.Double()
}
`
	libPkg, err := conf.Check("lib", fset, []*ast.File{mustParse(fset, libSrc)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	imports["lib"] = libPkg

	info := &Info{
		Uses:       make(map[*ast.Ident]Object),
		Defs:       make(map[*ast.Ident]Object),
		Selections: make(map[*ast.SelectorExpr]*Selection),
	}
	mainFile := mustParse(fset, mainSrc)
	if _, err := conf.Check("main", fset, []*ast.File{mainFile}, info); err != nil {
		t.Fatal(err)
	}

	var selExpr *ast.SelectorExpr
	ast.Inspect(mainFile, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Double" {
				selExpr = sel
			}
		}
		return true
	})
	var sel *Selection
	for expr, s := range info.Selections {
		if expr.Sel.Name == "Double" {
			sel = s
			if selExpr == nil {
				selExpr = expr
			}
			break
		}
	}
	if sel == nil {
		t.Fatal("extension call missing Selection")
	}
	if sel.Kind() != MethodVal {
		t.Fatalf("Selection kind = %v, want MethodVal", sel.Kind())
	}
	if sel.Obj().Name() != "Double" {
		t.Fatalf("Selection obj = %s, want Double", sel.Obj().Name())
	}
	if obj := info.Uses[selExpr.Sel]; obj == nil || obj.Name() != "Double" {
		t.Fatalf("Uses[Double] = %v, want Double", obj)
	}
}

func TestLinqExtensionSelectionRecording(t *testing.T) {
	const src = `package main

import "linq"

type P struct { Price float64; Tags []string }

func f(products []P) {
	_ = products.
		Where(p => p.Price >= 100).
		SelectMany(p => p.Tags).
		ToList()
}
`
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	info := &Info{
		Uses:       make(map[*ast.Ident]Object),
		Defs:       make(map[*ast.Ident]Object),
		Selections: make(map[*ast.SelectorExpr]*Selection),
	}
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, "source", nil),
	}
	if _, err := conf.Check("main", fset, []*ast.File{f}, info); err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{"Where": false, "SelectMany": false}
	for expr, s := range info.Selections {
		name := expr.Sel.Name
		if _, ok := want[name]; !ok {
			continue
		}
		if s.Kind() != MethodVal || s.Obj().Name() != name {
			t.Errorf("Selection for .%s() = %v %s, want method %s", name, s.Kind(), s.Obj().Name(), name)
		}
		if obj := info.Uses[expr.Sel]; obj == nil || obj.Name() != name {
			t.Errorf("Uses[%s] = %v, want %s", name, obj, name)
		}
		want[name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("missing Selection for .%s()", name)
		}
	}
}
