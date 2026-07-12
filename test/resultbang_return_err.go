// run

package main

import "errors"

func f(s string) []string! {
	var out []string
	x, err := parse(s)
	if err != nil {
		return errors.New("fail: %v", err)
	}
	out = append(out, x)
	return out
}

func parse(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty")
	}
	return s, nil
}

func main() {
	got, err := f("x")
	if err != nil {
		panic(err)
	}
	if len(got) != 1 || got[0] != "x" {
		panic(got)
	}
}
