
package types2_test

import (
	"strings"
	"testing"
)

func TestTypeSwitchDuplicateNamedType(t *testing.T) {
	t.Parallel()
	src := `
package p
func f(x any) {
	switch x.(type) {
	case int:
	case string:
	case int:
	}
}`
	_, err := typecheck(src, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "duplicate case int in type switch") {
		t.Fatalf("got %v; want duplicate int error", err)
	}
}

func TestTypeSwitchDuplicateNil(t *testing.T) {
	t.Parallel()
	src := `
package p
func f(x any) {
	switch x.(type) {
	case nil:
	case nil:
	}
}`
	_, err := typecheck(src, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "duplicate case nil in type switch") {
		t.Fatalf("got %v; want duplicate nil error", err)
	}
}

func TestTypeSwitchDuplicateAnonymousStruct(t *testing.T) {
	t.Parallel()
	src := `
package p
type S struct{ a int }
func f(x any) {
	switch x.(type) {
	case struct{ a int }:
	case S:
	case struct{ a int }:
	}
}`
	_, err := typecheck(src, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "duplicate case struct{a int} in type switch") {
		t.Fatalf("got %v; want duplicate struct error", err)
	}
}

func TestTypeSwitchDistinctCases(t *testing.T) {
	t.Parallel()
	src := `
package p
func f(x any) {
	switch x.(type) {
	case int, int8, int16, int32, int64:
	case uint, uint8, uint16, uint32, uint64:
	case float32, float64:
	case string, bool:
	case complex64, complex128:
	}
}`
	if _, err := typecheck(src, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
