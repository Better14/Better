// Copyright authors of this Go fork

package types_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	. "go/types"
)

func loadLinqPackage(t *testing.T, fset *token.FileSet, ignoreBodies bool) *Package {
	t.Helper()
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		t.Fatal("GOROOT not set")
	}
	linqDir := filepath.Join(goroot, "src", "linq")
	entries, err := os.ReadDir(linqDir)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(linqDir, e.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	stdImp := importer.ForCompiler(fset, "source", nil)
	conf := &Config{GoVersion: "go1.27", Importer: stdImp, IgnoreFuncBodies: ignoreBodies}
	pkg, err := conf.Check("linq", fset, files, nil)
	if err != nil {
		t.Fatalf("linq Check: %v", err)
	}
	return pkg
}

func checkLinqMain(t *testing.T, fset *token.FileSet, f *ast.File, linqPkg *Package) {
	t.Helper()
	stdImp := importer.ForCompiler(fset, "source", nil)
	var errs []string
	conf := &Config{
		GoVersion: "go1.27",
		Importer: linqImportOnlyImporter{
			linq:     linqPkg,
			fallback: stdImp,
		},
		Error: func(err error) { errs = append(errs, err.Error()) },
	}
	_, err := conf.Check("main", fset, []*ast.File{f}, nil)
	for _, e := range errs {
		t.Log(e)
	}
	for _, e := range errs {
		if strings.Contains(e, "lambda") || strings.Contains(e, "no matching overload") || strings.Contains(e, "ambiguous") {
			t.Fatalf("unexpected: %s", e)
		}
	}
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestLinqImportOnlyWhereSelectChain(t *testing.T) {
	const src = `package main

import "linq"

type P struct { ID int; InStock bool }

func f(products []P) {
	_ = products.Where(p => p.InStock).Select(p => p.ID).ToList()
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	linqPkg := loadLinqPackage(t, fset, true)
	checkLinqMain(t, fset, f, linqPkg)
}

// Re-type-checking the same parsed AST must remain valid. gopls reuses syntax
// trees from its parse cache across diagnostics passes.
func TestLinqExtensionCallReusesAST(t *testing.T) {
	const src = `package main

import "linq"

type P struct { ID int; InStock bool }

func f(products []P) {
	_ = products.Where(p => p.InStock).Select(p => p.ID).ToList()
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	linqPkg := loadLinqPackage(t, fset, true)
	checkLinqMain(t, fset, f, linqPkg)
	checkLinqMain(t, fset, f, linqPkg)
}

type linqImportOnlyImporter struct {
	linq     *Package
	fallback interface{ Import(string) (*Package, error) }
}

func (i linqImportOnlyImporter) Import(path string) (*Package, error) {
	if path == "linq" {
		return i.linq, nil
	}
	return i.fallback.Import(path)
}
