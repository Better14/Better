// Copyright authors of this Go fork

package types_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	. "go/types"
)

func TestLinqSliceExtensionEnsureImported(t *testing.T) {
	const src = `package main

import "linq"

type Post struct {
	Score int
}

func f(posts []Post) {
	for _, line := range posts.OrderByDescending(p => p.Score)
		.Take(5)
		.ToList() {
		_ = line
	}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var foundRange bool
	ast.Inspect(f, func(n ast.Node) bool {
		if r, ok := n.(*ast.RangeStmt); ok {
			foundRange = true
			if call, ok := r.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					t.Logf("range X = %T sel=%s", call.Fun, sel.Sel.Name)
				}
			}
		}
		return true
	})
	if !foundRange {
		t.Fatal("missing range stmt")
	}

	info := &Info{
		Uses:       make(map[*ast.Ident]Object),
		Defs:       make(map[*ast.Ident]Object),
		Scopes:     make(map[ast.Node]*Scope),
	}
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, "source", nil),
		Error: func(err error) {
			t.Logf("type error: %v", err)
		},
	}
	_, err = conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestLinqSliceExtensionSingleLine(t *testing.T) {
	const src = `package main

import "linq"

type Post struct {
	Score int
}

func f(posts []Post) {
	_ = posts.OrderByDescending(p => p.Score).Take(5).ToList()
}
`
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	var errs []string
	info := &Info{
		Uses:   make(map[*ast.Ident]Object),
		Defs:   make(map[*ast.Ident]Object),
		Scopes: make(map[ast.Node]*Scope),
	}
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, "source", nil),
		Error: func(err error) {
			errs = append(errs, err.Error())
		},
	}
	_, err := conf.Check(f.Name.Name, fset, []*ast.File{f}, info)
	for _, e := range errs {
		if strings.Contains(e, "undefined: slices") || strings.Contains(e, "slices") && strings.Contains(e, "undefined") {
			t.Fatalf("unexpected slices error: %s", e)
		}
	}
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestLinqSliceExtensionWithoutScopes(t *testing.T) {
	const src = `package main

import "linq"

type Post struct {
	Score int
}

func f(posts []Post) {
	_ = posts.OrderByDescending(p => p.Score).Take(5).ToList()
}
`
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, "source", nil),
		Error: func(err error) {
			errs = append(errs, err.Error())
		},
	}
	_, err := conf.Check(f.Name.Name, fset, []*ast.File{f}, nil)
	for _, e := range errs {
		if strings.Contains(e, "undefined: slices") {
			t.Fatalf("unexpected slices error: %s", e)
		}
	}
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}
