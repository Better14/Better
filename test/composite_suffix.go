// run

// Test postfix ! and ? binding with composite types:
//   []T!     -> ([]T, error)
//   [](T!)   -> [] of T! value type
//   []T?     -> ([]T)? explicit optional slice
//   [](T?)   -> [] of T? elements

package main

import "fmt"

func sliceResult() []string! {
	return nil
}

func sliceOfResults() [](string!) {
	var s string! = "ok"
	return [](string!){s}
}

func optionalSlice() []string? {
	var absent []string? = nil
	if absent != nil {
		panic("absent optional slice")
	}
	data := []string{"a"}
	return &data
}

func sliceOfOptionals() [](string?) {
	hi := "a"
	return [](string?){nil, &hi, nil}
}

func exercise() int! {
	rows := sliceResult()!
	if rows != nil {
		panic("expected nil slice")
	}
	got := sliceOfResults()
	if len(got) != 1 || got[0].value != "ok" || got[0].err != nil {
		panic(fmt.Sprintf("sliceOfResults: %#v", got))
	}
	present := optionalSlice()
	if present == nil {
		panic("optionalSlice nil")
	} else if len(present) != 1 || present[0] != "a" {
		panic(fmt.Sprintf("optionalSlice: %#v", present))
	}
	opts := sliceOfOptionals()
	if len(opts) != 3 {
		panic(fmt.Sprintf("sliceOfOptionals len: %d", len(opts)))
	}
	if opts[1] == nil {
		panic(fmt.Sprintf("sliceOfOptionals nil elem: %#v", opts))
	} else if (opts[1] ?? "") != "a" {
		panic(fmt.Sprintf("sliceOfOptionals: %#v", opts))
	}
	return 0
}

func main() {
	if _, err := exercise(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
}
