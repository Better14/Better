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

func TestLinqJoinSliceInnerArg(t *testing.T) {
	src := `package main

import "linq"

type Sale struct { ProductID int; Region string; Units int }
type Product struct { ID int; Price float64 }

var Sales []Sale

func f(products []Product) {
	_ = Sales.Join(
		products,
		s => s.ProductID,
		p => p.ID,
		(s, p) => struct{ Region string; Amount int }{s.Region, s.Units * int(p.Price)},
	).
		GroupBy(x => x.Region).
		ToList()
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, "source", nil),
		Error:     func(err error) { errs = append(errs, err.Error()) },
	}
	info := &Info{
		Uses: make(map[*ast.Ident]Object),
	}
	_, err = conf.Check("main", fset, []*ast.File{f}, info)
	for _, e := range errs {
		t.Log(e)
	}
	for _, e := range errs {
		if strings.Contains(e, "undefined: slices") || strings.Contains(e, "ambiguous extension") {
			t.Fatalf("unexpected: %s", e)
		}
	}
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	// Verify compiler-inserted slices import is resolved.
	for id, obj := range info.Uses {
		if id.Name == "slices" {
			if _, ok := obj.(*PkgName); !ok {
				t.Fatalf("slices use %v is not a package name", obj)
			}
			return
		}
	}
	t.Fatal("expected slices import use from Join slice adaptation")
}
