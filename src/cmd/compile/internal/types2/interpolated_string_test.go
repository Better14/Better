package types2

import "testing"

func TestParseQuotedInterpolationMuxTemplatesLiteral(t *testing.T) {
	quoted := `"/remote-buckets/{source-bucket-name}/{arn}"`
	format, holes, ok := parseQuotedInterpolation(quoted)
	if !ok {
		t.Fatal("parse failed")
	}
	if len(holes) != 0 {
		t.Fatalf("holes = %v, want none", holes)
	}
	want := "/remote-buckets/{source-bucket-name}/{arn}"
	if format != want {
		t.Fatalf("format = %q, want %q", format, want)
	}
}

func TestParseQuotedInterpolationWithFormatSpec(t *testing.T) {
	quoted := `"hello {name:s}"`
	format, holes, ok := parseQuotedInterpolation(quoted)
	if !ok {
		t.Fatal("parse failed")
	}
	if len(holes) != 1 || holes[0].exprSrc != "name" || holes[0].format != "s" {
		t.Fatalf("holes = %v, want [{name s}]", holes)
	}
	if format != "hello %s" {
		t.Fatalf("format = %q, want %q", format, "hello %s")
	}
}
