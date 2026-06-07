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

enum Message {
	Quit
	Write { text string, bytes int }
	ChangeColor { r, g, b uint8 }
}

func main() {
	var a SomeEnum = Value1
	var msg SomeEnum = Value2("hello")
	var c SomeEnum = Value3(42)

	if a == msg {
		panic("unexpected equality")
	}

	switch v := msg; v {
	case Value1:
		panic("wrong variant")
	case Value2(s):
		if s != "hello" {
			panic(s)
		}
	case Value3(x):
		panic(x)
	case Value4:
		panic("value4")
	}

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
	if num != 42 {
		panic(num)
	}

	m := Message.Write{ text: "hi", bytes: 5 }
	desc := switch m {
	case Quit:
		"quit"
	case Write { text }:
		text
	case ChangeColor { r, g, b }:
		string([]byte{r, g, b})
	}
	if desc != "hi" {
		panic(desc)
	}

	color := Message.ChangeColor { r: 1, g: 2, b: 3 }
	switch color {
	case ChangeColor { r, g, b }:
		if r != 1 || g != 2 || b != 3 {
			panic("struct pattern binding failed")
		}
	default:
		panic("unexpected message variant")
	}

	switch a {
	case Value1:
	default:
		panic("default should not run")
	}

	_ = a
	println("ok")
}
