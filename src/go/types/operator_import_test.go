// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	. "go/types"
)

func TestImportedPackageOperatorOverload(t *testing.T) {
	const matrixSrc = `package matrix

struct Matrix {
	rows, cols int
	data       []float64
}

func +(l, r *Matrix) *Matrix {
	return l
}
`
	const mainSrc = `package main

import "matrix"

func demo(a, b *matrix.Matrix) {
	_ = a + b
}
`
	fset := token.NewFileSet()
	matrixFile, err := parser.ParseFile(fset, "matrix/matrix.go", matrixSrc, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := &Config{GoVersion: "go1.27"}
	matrixPkg, err := conf.Check("matrix", fset, []*ast.File{matrixFile}, nil)
	if err != nil {
		t.Fatalf("matrix Check failed: %v", err)
	}

	mainFile, err := parser.ParseFile(fset, "main.go", mainSrc, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf.Importer = importerFunc(func(path string) (*Package, error) {
		if path == "matrix" {
			return matrixPkg, nil
		}
		return nil, fmt.Errorf("unknown import %q", path)
	})
	if _, err := conf.Check("main", fset, []*ast.File{mainFile}, nil); err != nil {
		t.Fatalf("main Check failed: %v", err)
	}
}

type importerFunc func(path string) (*Package, error)

func (f importerFunc) Import(path string) (*Package, error) { return f(path) }
