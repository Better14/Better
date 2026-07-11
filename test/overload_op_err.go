// errorcheck

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T int

func ==(a, b T) bool { return a == b } // ERROR "operator == requires paired operator !="
