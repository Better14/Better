// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

enum Mode {
	Read
	ByName(string)
}

func bad(mode Mode = ByName("x")) {} // ERROR "compile-time constant"
