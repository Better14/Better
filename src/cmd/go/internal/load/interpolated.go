// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package load

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// usesInterpolatedStrings reports whether any of the package's Go source
// files contain double-quoted string literals with {expr} interpolation holes.
func usesInterpolatedStrings(goFiles []string) bool {
	fset := token.NewFileSet()
	for _, path := range goFiles {
		if fileHasInterpolatedStrings(fset, path) {
			return true
		}
	}
	return false
}

func fileHasInterpolatedStrings(fset *token.FileSet, path string) bool {
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return false
	}
	skip := byteSliceStringLits(f)
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || !isDoubleQuoted(lit.Value) || skip[lit] {
			return true
		}
		body, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		for i := 0; i < len(body); i++ {
			if body[i] != '{' {
				continue
			}
			if _, isHole := scanInterpHole(body, i); isHole {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func isDoubleQuoted(val string) bool {
	return len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"'
}

func byteSliceStringLits(f *ast.File) map[*ast.BasicLit]bool {
	skip := map[*ast.BasicLit]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		arr, ok := call.Fun.(*ast.ArrayType)
		if !ok || arr.Len != nil {
			return true
		}
		id, ok := arr.Elt.(*ast.Ident)
		if !ok || id.Name != "byte" {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				skip[lit] = true
			}
		}
		return true
	})
	return skip
}

func scanInterpHole(body string, start int) (end int, ok bool) {
	if start >= len(body) || body[start] != '{' {
		return 0, false
	}
	close := strings.IndexByte(body[start+1:], '}')
	if close < 0 {
		return 0, false
	}
	end = start + 1 + close
	inside := body[start+1 : end]
	exprPart, _ := splitInterpInside(inside)
	if strings.TrimSpace(exprPart) == "" {
		return 0, false
	}
	test := "package p; func _() { _ = " + exprPart + " }"
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "hole.go", test, 0)
	return end, err == nil
}

func splitInterpInside(inside string) (expr, format string) {
	colon := strings.IndexByte(inside, ':')
	if colon < 0 {
		return strings.TrimSpace(inside), ""
	}
	return strings.TrimSpace(inside[:colon]), strings.TrimSpace(inside[colon+1:])
}
