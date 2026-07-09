// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt

import "unicode/utf8"

// buffer is a lightweight byte buffer used by scan.go.
type buffer []byte

func (b *buffer) writeByte(c byte) {
	*b = append(*b, c)
}

func (b *buffer) writeRune(r rune) {
	*b = utf8.AppendRune(*b, r)
}
