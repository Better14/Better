// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:nilable_pointers enable

package main

func main() {
	var p *int
	_ = p
	p = nil       // ERROR "cannot use nil"
	var q *int? = nil
	var r *int = q // ERROR "cannot use"
	_ = r
	_ = q
	if q != nil {
		var s *int = q
		_ = s
	}
}
