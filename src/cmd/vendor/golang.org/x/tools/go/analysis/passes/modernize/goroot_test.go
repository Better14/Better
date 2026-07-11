// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modernize

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestPkgInGOROOT(t *testing.T) {
	goroot := t.TempDir()
	gorootSrc := filepath.Join(goroot, "src", "go", "parser")
	if err := os.MkdirAll(gorootSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	gorootFile := filepath.Join(gorootSrc, "parser.go")
	if err := os.WriteFile(gorootFile, []byte("package parser\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "main.go")
	if err := os.WriteFile(outsideFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldGOROOT := build.Default.GOROOT
	build.Default.GOROOT = goroot
	t.Cleanup(func() { build.Default.GOROOT = oldGOROOT })

	check := func(t *testing.T, filename string, want bool) {
		t.Helper()
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		pass := &analysis.Pass{Fset: fset, Files: []*ast.File{f}}
		if got := pkgInGOROOT(pass); got != want {
			t.Fatalf("pkgInGOROOT(%q) = %v, want %v", filename, got, want)
		}
	}

	t.Run("inside", func(t *testing.T) {
		check(t, gorootFile, true)
	})
	t.Run("outside", func(t *testing.T) {
		check(t, outsideFile, false)
	})
}
