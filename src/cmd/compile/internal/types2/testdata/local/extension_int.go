// go1.27

package p

func (i int) Double() int { return i + i }

func _() {
	_ = 21.Double()
}
