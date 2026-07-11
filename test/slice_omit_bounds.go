// run

package main

func main() {
	list := []int{0, 1, 2, 3, 4, 5}

	head := list[:5]
	if len(head) != 5 || head[4] != 4 {
		panic("list[:5]")
	}

	tail := list[3:]
	if len(tail) != 3 || tail[0] != 3 || tail[2] != 5 {
		panic("list[3:]")
	}

	full := list[:]
	if len(full) != 6 || full[0] != 0 || full[5] != 5 {
		panic("list[:]")
	}

	// negative indices with omitted low bound
	trim := list[:-1]
	if len(trim) != 5 || trim[4] != 4 {
		panic("list[:-1]")
	}

	mid := list[1:-1]
	if len(mid) != 4 || mid[0] != 1 || mid[3] != 4 {
		panic("list[1:-1]")
	}
}
