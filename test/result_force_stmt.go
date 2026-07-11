// run

// Test (T, error)! as a standalone statement without _ =.

package main

import (
	"errors"
	"fmt"
)

var errBoom = errors.New("boom")

type widget struct{ n int }

func fakeExec(query string) (widget, error) {
	if query == "fail" {
		return widget{}, errBoom
	}
	return widget{1}, nil
}

func withExecForceStmt() int! {
	fakeExec("ok")!
	return 42
}

func withExecForceStmtFail() int! {
	fakeExec("fail")!
	return 42
}

func withExecForceAssign() int! {
	_ = fakeExec("ok")!
	return 7
}

func main() {
	v, err := withExecForceStmt()
	if err != nil || v != 42 {
		panic(fmt.Sprintf("withExecForceStmt: v=%d err=%v", v, err))
	}

	v, err = withExecForceStmtFail()
	if err != errBoom || v != 0 {
		panic(fmt.Sprintf("withExecForceStmtFail: v=%d err=%v", v, err))
	}

	v, err = withExecForceAssign()
	if err != nil || v != 7 {
		panic(fmt.Sprintf("withExecForceAssign: v=%d err=%v", v, err))
	}

	fmt.Println("ok")
}
