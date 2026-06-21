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

func TestLinqSelectChain(t *testing.T) {
	src := `package main

import "linq"

type P struct { Category string; InStock bool; Price float64 }
type categorySummary struct {
	Category string
	Count int
	AvgPrice float64
}

func f(products []P) {
	_ = products.
		Where(p => p.InStock).
		GroupBy(p => p.Category).
		Select(g => categorySummary{
			Category: g.Key,
			Count: g.Count(),
			AvgPrice: g.Values().Select(p => p.Price).Average(),
		}).
		ToList()
	_ = products.
		Where(p => p.InStock).
		GroupBy(p => p.Category).
		Select(g => g.Key).
		ToList()
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil { t.Fatal(err) }
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer: importer.ForCompiler(fset, "source", nil),
		Error: func(err error) { errs = append(errs, err.Error()) },
	}
	_, err = conf.Check("main", fset, []*ast.File{f}, nil)
	for _, e := range errs { t.Log(e) }
	for _, e := range errs {
		if strings.Contains(e, "lambda") || strings.Contains(e, "no matching overload") || strings.Contains(e, "ambiguous overloaded") {
			t.Fatalf("unexpected: %s", e)
		}
	}
	if err != nil { t.Fatalf("failed: %v", err) }
}
