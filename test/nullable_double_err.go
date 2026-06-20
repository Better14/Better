// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Applying ? twice is invalid.

package main

func main() {
	var _ (int?)? = nil // ERROR "invalid nullable type"
}
