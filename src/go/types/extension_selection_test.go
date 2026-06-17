// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/ast"
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
	if selExpr == nil {
		t.Fatal("missing s.Double() call")
	}
	sel, ok := info.Selections[selExpr]
	if !ok {
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
