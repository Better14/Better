// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"strings"
	"testing"

	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
)

func TestEnumDefaultArgs(t *testing.T) {
	const src = `package p

enum Mode {
	Read
	Write
}

func open(path string, mode Mode = Read) Mode {
	return mode
}

func openQualified(path string, mode Mode = Mode.Read) Mode {
	return mode
}

func use() {
	_ = open("x")
	_ = openQualified("x")
}
`
	f, err := syntax.Parse(syntax.NewFileBase("p"), strings.NewReader(src), nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := &types2.Config{}
	_, err = conf.Check("p", []*syntax.File{f}, nil)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestEnumDefaultArgsRejectTupleVariant(t *testing.T) {
	const src = `package p

enum Mode {
	Read
	ByName(string)
}

func bad(mode Mode = ByName("x")) {}
`
	f, err := syntax.Parse(syntax.NewFileBase("p"), strings.NewReader(src), nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := &types2.Config{}
	_, err = conf.Check("p", []*syntax.File{f}, nil)
	if err == nil {
		t.Fatal("expected type error for tuple enum default")
	}
}

func TestEnumCompositeLitRequiresQualification(t *testing.T) {
	const src = `package p

enum Section {
	MatrixOps { size int, operators []string }
}

func bad() {
	_ = MatrixOps{size: 2, operators: nil}
}
`
	f, err := syntax.Parse(syntax.NewFileBase("p"), strings.NewReader(src), nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := &types2.Config{}
	_, err = conf.Check("p", []*syntax.File{f}, nil)
	if err == nil {
		t.Fatal("expected type error for unqualified enum struct composite literal")
	}
}
