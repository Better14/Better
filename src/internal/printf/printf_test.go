package printf_test

import (
	"internal/printf"
	"testing"
)

func TestMissingWrapArgs(t *testing.T) {
	p := printf.New()
	p.SetWrapErrs(true)
	p.Printf("too many verbs: %w %v")
	if got := len(p.WrappedErrs()); got != 0 {
		t.Fatalf("WrappedErrs len = %d, want 0; %v", got, p.WrappedErrs())
	}
	want := "too many verbs: %!w(MISSING) %!v(MISSING)"
	if got := p.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	p.Free()
}

func TestMissingWrapArgsViaSlice(t *testing.T) {
	var a []any
	p := printf.New()
	p.SetWrapErrs(true)
	p.Printf("too many verbs: %w %v", a...)
	we := p.WrappedErrs()
	if len(we) != 0 {
		t.Fatalf("WrappedErrs = %v, want empty", we)
	}
	p.Free()
}
