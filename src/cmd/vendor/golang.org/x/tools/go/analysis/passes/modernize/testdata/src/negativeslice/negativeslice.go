package negativeslice

func f(list []int) {
	_ = list[len(list)-2:len(list)] // want "slice bounds can use negative index syntax"
	_ = list[0:len(list)-1]         // want "slice bounds can use negative index syntax"
}
