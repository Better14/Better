// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !compiler_bootstrap

package test

// The racecompile builder only builds packages, but does not build
// or run tests. This is a non-test file to hold cases that (used
// to) trigger compiler data races, so they will be exercised on
// the racecompile builder.
//
// This package is not imported so functions here are not included
// in the actual compiler.

// Issue 55357: data race when building multiple instantiations of
// generic closures with _ parameters.
func Issue55357() {
	var q t55357U
	q.Count()
	q.List()

	var q2 t55357M
	q2.Count()
	q2.List()
}

type issue55357U struct {
	A int
	B string
	C string
}

type issue55357M struct {
	A int64
	B uint32
	C uint32
}

type t55357U struct{}

//go:noinline
func (q *t55357U) do(w, v bool, fn func(bk []byte, v issue55357U) error) error {
	return nil
}

func (q *t55357U) Count() (n int, rerr error) {
	err := q.do(false, false, func(kb []byte, _ issue55357U) error {
		n++
		return nil
	})
	return n, err
}

func (q *t55357U) List() (list []issue55357U, rerr error) {
	var l []issue55357U
	err := q.do(false, true, func(_ []byte, v issue55357U) error {
		l = append(l, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

type t55357M struct{}

//go:noinline
func (q *t55357M) do(w, v bool, fn func(bk []byte, v issue55357M) error) error {
	return nil
}

func (q *t55357M) Count() (n int, rerr error) {
	err := q.do(false, false, func(kb []byte, _ issue55357M) error {
		n++
		return nil
	})
	return n, err
}

func (q *t55357M) List() (list []issue55357M, rerr error) {
	var l []issue55357M
	err := q.do(false, true, func(_ []byte, v issue55357M) error {
		l = append(l, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}
