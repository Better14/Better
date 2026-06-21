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

func TestImportedPackageFunctionOverload(t *testing.T) {
	const overloadSrc = `package overload

func Format(v int) string { return "" }
func Format(v int64) string { return "" }
func Format(v float64, decimals int = 2) string { return "" }
func Format(v string, upper bool = false) string { return "" }

func Sum(a int) int { return a }
func Sum(a, b int) int { return a + b }
func Sum(a, b, c int) int { return a + b + c }

func Describe(label string, value int) string { return "" }
func Describe(label string, value string, quoted bool = true) string { return "" }
`
	const mainSrc = `package main

import "overload"

func demo() {
	_ = overload.Format(42)
	_ = overload.Format(int64(9007199254740991))
	_ = overload.Format(3.14159)
	_ = overload.Format("hello", true)
	_ = overload.Sum(1)
	_ = overload.Sum(1, 2)
	_ = overload.Sum(1, 2, 3)
	_ = overload.Describe("count", 7)
	_ = overload.Describe("name", "Ada")
}
`
	fset := token.NewFileSet()
	overloadFile, err := parser.ParseFile(fset, "overload/overload.go", overloadSrc, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := &Config{GoVersion: "go1.27"}
	overloadPkg, err := conf.Check("overload", fset, []*ast.File{overloadFile}, nil)
	if err != nil {
		t.Fatalf("overload Check failed: %v", err)
	}

	mainFile, err := parser.ParseFile(fset, "main.go", mainSrc, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf.Importer = importerFunc(func(path string) (*Package, error) {
		if path == "overload" {
			return overloadPkg, nil
		}
		return nil, fmt.Errorf("unknown import %q", path)
	})
	if _, err := conf.Check("main", fset, []*ast.File{mainFile}, nil); err != nil {
		t.Fatalf("main Check failed: %v", err)
	}
}
