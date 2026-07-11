
package types_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestEnumDefaultArgs(t *testing.T) {
	const src = `package p

enum Mode {
	Read
	Write
}

func open(path string, mode Mode = Read) Mode {
	return mode
}

func openQualified(path string, mode Mode = Mode.Read) Mode {
	return mode
}

func use() {
	_ = open("x")
	_ = openQualified("x")
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	cfg := types.Config{Importer: importer.Default()}
	_, err = cfg.Check("p", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestEnumDefaultArgsRejectTupleVariant(t *testing.T) {
	const src = `package p

enum Mode {
	Read
	ByName(string)
}

func bad(mode Mode = ByName("x")) {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	cfg := types.Config{Importer: importer.Default()}
	_, err = cfg.Check("p", fset, []*ast.File{f}, nil)
	if err == nil {
		t.Fatal("expected type error for tuple enum default")
	}
}

func TestEnumCompositeLitRequiresQualification(t *testing.T) {
	const src = `package p

enum Section {
	MatrixOps { size int, operators []string }
}

func bad() {
	_ = MatrixOps{size: 2, operators: nil}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	cfg := types.Config{Importer: importer.Default()}
	_, err = cfg.Check("p", fset, []*ast.File{f}, nil)
	if err == nil {
		t.Fatal("expected type error for unqualified enum struct composite literal")
	}
}
