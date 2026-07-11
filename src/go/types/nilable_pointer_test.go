// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func checkNilablePointers(t *testing.T, src, modMode string) error {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := types.NewPackage("p", "p")
	pkg.SetNilablePointers(modMode)
	return types.NewChecker(&types.Config{}, fset, pkg, nil).Files([]*ast.File{f})
}

func TestNilablePointerNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a *int?) {
	if a != nil {
		var b *int = a
		_ = b
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var p *int = nil
	_ = p
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to *int")
	}
}

func TestNilablePointerDisabledEquivalence(t *testing.T) {
	const src = `package p
func f(a *int?, b *int) {
	var x *int = a
	var y *int? = b
	_ = x
	_ = y
}`
	if err := checkNilablePointers(t, src, "disable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerRegionEnd(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func enabled() {
	var p *int = nil
	_ = p
}
//go:nilable_pointers end
func disabled(a *int?, b *int) {
	var x *int = a
	var y *int? = b
	_ = x
	_ = y
}`
	if err := checkNilablePointers(t, src, "disable"); err == nil {
		t.Fatal("expected error in enabled region")
	}
}

func TestNilablePointerModDefault(t *testing.T) {
	const src = `package p
func f() {
	var p *int = nil
	_ = p
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error with go.mod enable default")
	}
}

func TestNilablePointerEarlyReturnNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a *int?) int {
	if a == nil {
		return 0
	}
	var b *int = a
	return *b
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerDefaultAssignNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
type Info struct {
	API string
}

func Get() *Info? {
	return nil
}

func Use() {
	req := Get()
	if req == nil {
		req = &Info{API: "SYSTEM"}
	}
	_ = req.API
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerOrNilFirstNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
type E struct { Children map[string]struct{} }
func find() *E? { return nil }
func f() *E? {
	root := find()
	if root == nil || len(root.Children) == 0 {
		return root
	}
	x := *root
	return &x
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerPackageEarlyReturn(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
var dbConn *int?

func Conn() *int {
	if dbConn == nil {
		panic("uninit")
	}
	return dbConn
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}
func TestNilablePointerCrossPackageNarrowing(t *testing.T) {
	const srcM = `package m
//go:nilable_pointers enable
func Get() *int? { return nil }`
	const srcP = `package p
import "m"
//go:nilable_pointers enable
func f() {
	auth := m.Get()
	if auth != nil {
		var x *int = auth
		_ = x
	}
}`
	fset := token.NewFileSet()
	m, err := parser.ParseFile(fset, "m.go", srcM, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.ParseFile(fset, "p.go", srcP, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	mpkg := types.NewPackage("m", "m")
	mpkg.SetNilablePointers("enable")
	if err := types.NewChecker(&types.Config{}, fset, mpkg, nil).Files([]*ast.File{m}); err != nil {
		t.Fatalf("m: %v", err)
	}
	pkg := types.NewPackage("p", "p")
	pkg.SetNilablePointers("enable")
	if err := types.NewChecker(&types.Config{Importer: fakeImporter{mpkg}}, fset, pkg, nil).Files([]*ast.File{p}); err != nil {
		t.Fatalf("p: %v", err)
	}
}

type fakeImporter struct{ pkg *types.Package }

func (f fakeImporter) Import(path string) (*types.Package, error) { return f.pkg, nil }

func TestNilablePointerArgWrap(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func take(p *string?) {}
func f() {
	s := "id"
	take(&s)
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerAndOrNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a *int?) bool {
	return a != nil && *a > 0
}
func g(a *int?) bool {
	return a == nil || *a == 0
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilablePointerSelectorNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
type row struct {
	Body *string?
}
func f(r row) string {
	if r.Body != nil {
		return *r.Body
	}
	return ""
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableSliceNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var s []string = nil
	_ = s
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to []string")
	}
}

func TestNilableSliceOptionalAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var s []string? = nil
	_ = s
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableSliceNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a []string?) {
	if a != nil {
		var b []string = a
		_ = b
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableMapNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var m map[string]int = nil
	_ = m
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to map[string]int")
	}
}

func TestNilableMapOptionalAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var m map[string]int? = nil
	_ = m
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableMapNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a map[string]int?) {
	if a != nil {
		var b map[string]int = a
		_ = b
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableChanNilAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var ch chan int = nil
	_ = ch
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to chan int")
	}
}

func TestNilableChanOptionalAssign(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var ch chan int? = nil
	_ = ch
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableChanNarrowing(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(a chan int?) {
	if a != nil {
		var b chan int = a
		_ = b
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableChanElemOptional(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f() {
	var ch chan (*int?) = nil
	_ = ch
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for nil assign to strict chan (*int?)")
	}
}

func TestNilableChanCloseRequiresNilCheck(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(ch chan int?) {
	close(ch)
}`
	if err := checkNilablePointers(t, src, "enable"); err == nil {
		t.Fatal("expected error for close without nil check")
	}
}

func TestNilableChanCloseAfterNilCheck(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(ch chan int?) {
	if ch != nil {
		close(ch)
	}
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestNilableChanCloseAfterEarlyReturn(t *testing.T) {
	const src = `package p
//go:nilable_pointers enable
func f(ch chan int?) {
	if ch == nil {
		return
	}
	close(ch)
}`
	if err := checkNilablePointers(t, src, "enable"); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}
