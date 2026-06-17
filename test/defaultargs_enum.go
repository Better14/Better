// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

enum Mode {
	Read
	Write
	Both
}

func open(path string, mode Mode = Read) Mode {
	return mode
}

func openQualified(path string, mode Mode = Mode.Read) Mode {
	return mode
}

func main() {
	if got := open("x"); got != Read {
		panic(got)
	}
	if got := open("x", Write); got != Write {
		panic(got)
	}
	if got := openQualified("x"); got != Read {
		panic(got)
	}
}
