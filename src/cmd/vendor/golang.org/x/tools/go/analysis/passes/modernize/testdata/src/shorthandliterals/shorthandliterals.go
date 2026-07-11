package shorthandliterals

import "set"

func sliceLit() {
	_ = []string{"a", "b"} // want "slice composite literal can use array literal syntax"
}

func mapLit() {
	_ = map[string]string{"a": "b"} // want "map composite literal can use dict literal syntax"
}

func setLit() {
	_ = set.Of("a", "b") // want "set.Of call can use set literal syntax"
}
