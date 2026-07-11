
package types_test

import (
	"go/ast"
	"go/importer"
	"go/token"
	"strings"
	"testing"

	. "go/types"
)

func linqTestImporter(fset *token.FileSet) Importer {
	return importer.ForCompiler(fset, "source", nil)
}

func TestLinqSliceLambdaHintSource(t *testing.T) {
	t.Skip("source importer does not yet index linq extension methods for type-checking")
	src := `package main

import "linq"

type Post struct {
	ID string
}

func f(posts []Post) {
	_ = posts.Select(p => p.ID).ToList()
}
`
	fset := token.NewFileSet()
	f := mustParse(fset, src)
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer:  linqTestImporter(fset),
		Error: func(err error) {
			errs = append(errs, err.Error())
		},
	}
	_, err := conf.Check(f.Name.Name, fset, []*ast.File{f}, nil)
	for _, e := range errs {
		if strings.Contains(e, "lambda expression requires type context") {
			t.Fatalf("unexpected lambda hint error: %s", e)
		}
	}
	if err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}
}
