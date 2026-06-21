// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

enum Section {
	MatrixOps { size int, operators []string }
}

func main() {
	_ = MatrixOps{size: 2, operators: nil} // ERROR "undefined: MatrixOps"
}
