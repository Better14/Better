package shorthandliterals

import "set"

func sliceLit() {
	_ = []string{"a", "b"} // want "slice composite literal can use array literal syntax"
	_ = []int{1, 2, 3}     // want "slice composite literal can use array literal syntax"
}

func keepNonDefaultSlice() {
	_ = []int64{1, 2, 3}
	_ = []float32{1.0, 2.0}
}

type User struct{ Name string }

func (User) GetName() string { return "" }

type IUser interface{ GetName() string }

func keepInterfaceSlice() {
	_ = []IUser{User{"bob"}}
	_ = []any{User{"bob"}}
}

func rewriteStructSlice() {
	_ = []User{User{"bob"}} // want "slice composite literal can use array literal syntax"
}

func mapLit() {
	_ = map[string]string{"a": "b"} // want "map composite literal can use dict literal syntax"
}

func keepNonDefaultMap() {
	_ = map[string]int64{"a": 1}
}

func setLit() {
	_ = set.Of("a", "b") // want "set.Of call can use set literal syntax"
}

func keepTypedSet() {
	_ = set.Of[int64](1, 2, 3)
}
