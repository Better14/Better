// compile

package main

type ArrayDimension struct {
	Length     int32
	LowerBound int32
}

type FlatArray[T any] []T

func (a FlatArray[T]) Dimensions() []ArrayDimension {
	if a == nil {
		return nil
	}
	return []ArrayDimension{{Length: int32(len(a)), LowerBound: 1}}
}

func main() {
	var a FlatArray[int]
	_ = a.Dimensions()
}
