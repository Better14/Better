// go1.27

package p

func (s []T) Where(pred func(T) bool) []T {
	var out []T
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

func _() {
	nums := []int{1, 2, 3}
	_ = nums.Where(func(n int) bool { return n > 1 })
}
