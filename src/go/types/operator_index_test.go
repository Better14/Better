// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestIndexOperatorOverloadMultiIndex(t *testing.T) {
	const src = `
package matrix

type Matrix struct {
	rows, cols int
	data       []float64
}

func [](m *Matrix, i, j int) float64 {
	return m.data[i*m.cols+j]
}

func []=(m *Matrix, i, j int, v float64) {
	m.data[i*m.cols+j] = v
}

func use(m *Matrix) {
	_ = m[1, 2]
	m[0, 1] = 3.14
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "matrix.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{Importer: importer.Default()}
	_, err = conf.Check("matrix", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNullableNilInStructLiteral(t *testing.T) {
	const src = `
package data

type Employee struct {
	ID        int
	ManagerID int?
}

var Employees = []Employee{
	{1, nil},
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "data.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{Importer: importer.Default()}
	_, err = conf.Check("data", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}