package spreadcall

func spread(xs ...int) int { return len(xs) }

func call() {
	s := []int{1, 2, 3}
	_ = spread(s...) // want "variadic call can use prefix spread syntax"
}
