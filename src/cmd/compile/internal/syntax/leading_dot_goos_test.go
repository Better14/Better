// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"os"
	"strings"
	"testing"
)

func TestLeadingDotGoosPackage(t *testing.T) {
	data, err := os.ReadFile("../../../../internal/goos/goos.go")
	if err != nil {
		t.Fatal(err)
	}
	var first error
	_, err = Parse(NewFileBase("goos.go"), strings.NewReader(string(data)), func(e error) {
		if first == nil {
			first = e
		}
	}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if first != nil {
		t.Fatal(first)
	}
}
