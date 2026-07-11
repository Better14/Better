
package types_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"runtime"
	"strings"
	"testing"

	. "go/types"
)

func TestExtensionImportMarkedUsed(t *testing.T) {
	const src = `package main

import "linq"

func f(xs []int) {
	_ = xs.Where(n => n > 0)
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := &Info{
		Types:           make(map[ast.Expr]TypeAndValue),
		Defs:            make(map[*ast.Ident]Object),
		Uses:            make(map[*ast.Ident]Object),
		Implicits:       make(map[ast.Node]Object),
		Instances:       make(map[*ast.Ident]Instance),
		Selections:      make(map[*ast.SelectorExpr]*Selection),
		Scopes:          make(map[ast.Node]*Scope),
		UsedImportNames: make(map[string]bool),
	}
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  importer.ForCompiler(fset, runtime.Compiler, nil),
		Error: func(err error) {
			errs = append(errs, err.Error())
		},
	}
	if _, err := conf.Check(f.Name.Name, fset, []*ast.File{f}, info); err != nil {
		t.Fatalf("Check failed: %v (%v)", err, errs)
	}
	for _, e := range errs {
		if strings.Contains(e, "imported and not used") && strings.Contains(e, "linq") {
			t.Fatalf("linq reported unused: %s", e)
		}
	}
	if !info.UsedImportNames["linq"] {
		t.Fatalf("UsedImportNames missing linq: %v", info.UsedImportNames)
	}
}
