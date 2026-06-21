package local

struct S { x int }

func +(a, b S) S { return a }

func Use() {
	var a, b S
	_ = a + b
}
