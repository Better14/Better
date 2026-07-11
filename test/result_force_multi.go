// run

// Two x := call()! assignments in func () error must compile (no ICE).

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

type stamp struct{ n int }

func parseStamp(s string) (stamp, error) {
	if s == "fail" {
		return stamp{}, errBoom
	}
	return stamp{len(s)}, nil
}

func inferTwo() error {
	s1 := "a"
	t := parseStamp(s1)!
	s2 := "bb"
	t2 := parseStamp(s2)!
	if t.n != 1 || t2.n != 2 {
		return fmt.Errorf("got %d and %d", t.n, t2.n)
	}
	return nil
}

func inferTwoFail() error {
	parseStamp("fail")!
	return nil
}

func main() {
	if err := inferTwo(); err != nil {
		panic(err)
	}
	if err := inferTwoFail(); err != errBoom {
		panic(fmt.Sprintf("inferTwoFail: err=%v", err))
	}
	fmt.Println("ok")
}
