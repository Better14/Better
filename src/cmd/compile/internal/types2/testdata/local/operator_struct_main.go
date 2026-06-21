package main

struct S { x int }

func +(a, b S) S { return a }

func main() {
	var a, b S
	_ = a + b
}
