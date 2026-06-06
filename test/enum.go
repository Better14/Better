// run

// Copyright 2026 The Go Authors. All rights reserved.
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
	var a SomeEnum = SomeEnum.Value1
	var b SomeEnum = SomeEnum.Value2("hello")
	c := SomeEnum.Value3(42)

	switch v := b; v {
	case Value1:
		panic("wrong variant")
	case Value2(s):
		if s != "hello" {
			panic(s)
		}
	case Value3(n):
		panic(n)
	case Value4:
		panic("value4")
	}

	n := switch c {
	case Value1:
		0
	case Value2(s):
		len(s)
	case Value3(n):
		n
	case Value4:
		3
	}
	if n != 42 {
		panic(n)
	}

	_ = a
	println("ok")
}
