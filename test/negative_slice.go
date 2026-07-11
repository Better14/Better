// run

package main

func main() {
	list := []int{0, 1, 2, 3, 4, 5, 6}

	s := list[-2:]
	if len(s) != 2 || s[0] != 5 || s[1] != 6 {
		panic("negative low bound")
	}

	t := list[0:-2]
	if len(t) != 5 || t[4] != 4 {
		panic("negative high bound")
	}

	u := list[-3:-1]
	if len(u) != 2 || u[0] != 4 || u[1] != 5 {
		panic("negative both bounds")
	}

	arr := [7]int{0, 1, 2, 3, 4, 5, 6}
	v := arr[-1:]
	if len(v) != 1 || v[0] != 6 {
		panic("array negative slice")
	}
}
