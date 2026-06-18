// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs_test

import (
	"errors"
	"io/fs"
	"testing"
)

func TestNewPathErrorStackTrace(t *testing.T) {
	err := fs.NewPathError("open", "/etc/passwd", fs.ErrNotExist)
	if len(err.StackTrace) == 0 {
		t.Fatal("StackTrace empty, want frames")
	}
	if got := err.Error(); got != "open /etc/passwd: file does not exist" {
		t.Fatalf("Error() = %q", got)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("errors.Is failed")
	}
}
