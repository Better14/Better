// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func take(p *string?) {}

func main() {
	s := "id"
	take(&s)
}
