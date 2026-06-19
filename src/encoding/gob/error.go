// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gob

import (
	stderrors "errors"
	"fmt"
)

// Errors in decoding and encoding are handled using panic and recover.
// Panics caused by user error (that is, everything except run-time panics
// such as "index out of bounds" errors) do not leave the file that caused
// them, but are instead turned into plain error returns. Encoding and
// decoding functions and methods that do not return an error either use
// panic to report an error or are guaranteed error-free.

// A gobError is used to distinguish errors (panics) generated in this package.
type gobError struct {
	stderrors.Base
	err error
}

func newGobError(err error) gobError {
	e := gobError{err: err}
	stderrors.InitCustom(&e.Base, "%s", err.Error())
	return e
}

func (e gobError) Error() string { return e.err.Error() }

func (e gobError) Unwrap() error { return e.err }

// errorf is like error_ but takes Printf-style arguments to construct an error.
// It always prefixes the message with "gob: ".
func errorf(format string, args ...any) {
	error_(fmt.Errorf("gob: "+format, args...))
}

// error_ wraps the argument error and uses it as the argument to panic.
func error_(err error) {
	panic(newGobError(err))
}

// catchError is meant to be used as a deferred function to turn a panic(gobError) into a
// plain error. It overwrites the error return of the function that deferred its call.
func catchError(err *error) {
	if e := recover(); e != nil {
		ge, ok := e.(gobError)
		if !ok {
			panic(e)
		}
		*err = ge
	}
}
