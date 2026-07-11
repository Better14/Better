// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

enum SomeEnum {
	Value1
	Value2(string)
	Value3(int)
	Value4 = 3
}

func main() {
	var c SomeEnum = Value3(42)
	num := switch c {
	case Value1:
		0
	case Value2(s):
		len(s)
	case Value3(x):
		x
	case Value4:
		3
	}
	println(num)
}
