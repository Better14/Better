// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package length

import "testing"

func TestMapLenExtensionInternal(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	if m.MapLen() != 2 {
		t.Fatalf("map MapLen: %d", m.MapLen())
	}
	if MapLen(m) != 2 {
		t.Fatalf("MapLen func: %d", MapLen(m))
	}
}
