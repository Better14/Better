// errorcheck

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func bad(a int = 1, b int) {} // ERROR "missing default"

func bad2(a int, b int = 2, c int) {} // ERROR "missing default"

func bad3(x int = nonexistent) {} // ERROR "undefined"

func bad5(x int = f()) {} // ERROR "compile-time constant"

func f() int { return 1 }
