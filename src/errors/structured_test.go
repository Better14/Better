// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// appError embeds errors.Base for NewCustom tests.
type appError struct {
	errors.Base
}

// myError carries a structured error plus a Code field.
type myError struct {
	err  errors.Base
	Code int
}

func (e *myError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.err.Message
}

func TestStructuredErrorError(t *testing.T) {
	err := errors.New("abc")
	if got := err.Error(); got != "abc" {
		t.Fatalf("Error() = %q, want abc", got)
	}
	var nilErr *errors.Error
	if got := nilErr.Error(); got != "<nil>" {
		t.Fatalf("nil Error() = %q, want <nil>", got)
	}
}

func TestStructuredErrorString(t *testing.T) {
	root := errors.New("root")
	rootErr := root
	rootErr.StackTrace = nil // deterministic test output

	outer := rootErr.Wrap("outer")
	outer.StackTrace = errors.StackTrace{
		{Function: "main.outer", File: "outer.go", Line: 10},
	}

	got := outer.String()
	want := "root\nouter\nmain.outer (outer.go:10)"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestStackTraceString(t *testing.T) {
	st := errors.StackTrace{
		{Function: "main.a", File: "a.go", Line: 1},
		{Function: "main.b", File: "b.go", Line: 2},
	}
	want := "main.a (a.go:1)\nmain.b (b.go:2)"
	if got := st.String(); got != want {
		t.Fatalf("StackTrace.String() = %q, want %q", got, want)
	}
	if got := errors.StackTrace(nil).String(); got != "" {
		t.Fatalf("nil StackTrace.String() = %q, want empty", got)
	}
}

func TestWrap(t *testing.T) {
	root := errors.New("root")
	wrapped := errors.Wrap(root, "wrap")
	if wrapped.Error() != "wrap" {
		t.Fatalf("Wrap Error() = %q, want wrap", wrapped.Error())
	}
	if !errors.Is(wrapped, root) {
		t.Fatal("errors.Is(wrapped, root) = false, want true")
	}
	if len(wrapped.StackTrace) == 0 {
		t.Fatal("Wrap StackTrace is empty, want frames")
	}

	foreign := errStr("foreign")
	fw := errors.Wrap(foreign, "bridge")
	if !errors.Is(fw, foreign) {
		t.Fatal("errors.Is through bridge = false, want true")
	}
	if got := errors.Unwrap(errors.Unwrap(fw)); got != foreign {
		t.Fatalf("Unwrap bridge = %v, want %v", got, foreign)
	}
}

func TestWrapMethodOnNil(t *testing.T) {
	var e *errors.Error
	got := e.Wrap("msg")
	if got == nil || got.Error() != "msg" {
		t.Fatalf("nil Wrap = %v, want msg error", got)
	}
}

func TestNewWrappedFromFmt(t *testing.T) {
	inner := errors.New("inner")
	outer := fmt.Errorf("ctx: %w", inner)
	if got := outer.Error(); got != "ctx: inner" {
		t.Fatalf("Error() = %q, want ctx: inner", got)
	}
	if errors.Unwrap(outer) != inner {
		t.Fatalf("Unwrap = %v, want %v", errors.Unwrap(outer), inner)
	}
	e, ok := outer.(*errors.Error)
	if !ok {
		t.Fatalf("type = %T, want *errors.Error", outer)
	}
	if len(e.StackTrace) == 0 {
		t.Fatal("StackTrace empty, want frames")
	}
}

func TestNewCustomInPlace(t *testing.T) {
	var err myError
	errors.InitCustom(&err.err, "not found")
	err.Code = 404

	if err.Error() != "not found" {
		t.Fatalf("Error() = %q, want not found", err.Error())
	}
	if len(err.err.StackTrace) == 0 {
		t.Fatal("StackTrace empty, want frames")
	}
	if err.err.InnerError != nil {
		t.Fatalf("InnerError = %v, want nil", err.err.InnerError)
	}
	if err.Code != 404 {
		t.Fatalf("Code = %d, want 404", err.Code)
	}

	var target *myError
	if !errors.As(&err, &target) {
		t.Fatal("errors.As failed")
	}
}

func TestNewFormat(t *testing.T) {
	err := errors.New("open file: %s", "/etc/app.conf")
	if got := err.Error(); got != "open file: /etc/app.conf" {
		t.Fatalf("Error() = %q, want open file: /etc/app.conf", got)
	}
	if got := errors.New("100% complete").Error(); got != "100% complete" {
		t.Fatalf("literal percent = %q, want 100%% complete", got)
	}
}

func TestNewCustomFormat(t *testing.T) {
	err := errors.NewCustom[appError]("invalid id: %s", "abc")
	if err.Message != "invalid id: abc" {
		t.Fatalf("Message = %q, want invalid id: abc", err.Message)
	}

	var myErr myError
	errors.InitCustom(&myErr.err, "not found: %d", 404)
	myErr.Code = 404
	if myErr.Error() != "not found: 404" {
		t.Fatalf("Error() = %q, want not found: 404", myErr.Error())
	}
}

func TestNewCustom(t *testing.T) {
	err := errors.NewCustom[appError]("invalid id")
	if err.Message != "invalid id" {
		t.Fatalf("Message = %q, want invalid id", err.Message)
	}
	if len(err.StackTrace) == 0 {
		t.Fatal("StackTrace empty, want frames")
	}
	if err.InnerError != nil {
		t.Fatalf("InnerError = %v, want nil", err.InnerError)
	}
}

func TestStructuredErrorAs(t *testing.T) {
	root := errors.New("root")
	wrapped := errors.Wrap(root, "wrap")
	var target *errors.Error
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As failed")
	}
	if target != wrapped {
		t.Fatalf("As target = %p, want %p", target, wrapped)
	}
}

func TestStructuredErrorStringContainsTrace(t *testing.T) {
	err := errors.New("boom")
	e := err
	s := e.String()
	if !strings.Contains(s, "boom") {
		t.Fatalf("String() = %q, missing message", s)
	}
	if len(e.StackTrace) > 0 && !strings.Contains(s, e.StackTrace[0].Function) {
		t.Fatalf("String() = %q, missing stack frame %q", s, e.StackTrace[0].Function)
	}
}

func TestFmtPlusVStructuredError(t *testing.T) {
	err := errors.New("msg")
	got := fmt.Sprintf("%+v", err)
	if got == err.Error() {
		t.Fatalf("Sprintf(%%+v) = %q, want full String() output", got)
	}
	if !strings.Contains(got, "msg") {
		t.Fatalf("Sprintf(%%+v) = %q, missing message", got)
	}
}

type errStr string

func (e errStr) Error() string { return string(e) }
