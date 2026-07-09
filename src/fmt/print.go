// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt

import (
	"errors"
	"internal/printf"
	"io"
	"os"
)

// State represents the printer state passed to custom formatters.
type State = printf.State

// Formatter is implemented by any value that has a Format method.
type Formatter = printf.Formatter

// Stringer is implemented by any value that has a String method.
type Stringer = printf.Stringer

// GoStringer is implemented by any value that has a GoString method.
type GoStringer = printf.GoStringer

func init() {
	printf.ErrorPlusV = func(err error) (string, bool) {
		if e, ok := err.(*errors.Error); ok {
			return e.String(), true
		}
		return "", false
	}
	errors.RegisterFormatter(printf.Sprintf)
}

// FormatString returns a string representing the fully qualified formatting
// directive captured by the [State], followed by the argument verb.
func FormatString(state State, verb rune) string {
	return printf.FormatString(state, verb)
}

// Fprintf formats according to a format specifier and writes to w.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return printf.Fprintf(w, format, a...)
}

// Printf formats according to a format specifier and writes to standard output.
func Printf(format string, a ...any) (n int, err error) {
	return Fprintf(os.Stdout, format, a...)
}

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	return printf.Sprintf(format, a...)
}

// Appendf formats according to a format specifier, appends the result to the byte
// slice, and returns the updated slice.
func Appendf(b []byte, format string, a ...any) []byte {
	return printf.Appendf(b, format, a...)
}

// Fprint formats using the default formats for its operands and writes to w.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	return printf.Fprint(w, a...)
}

// Print formats using the default formats for its operands and writes to standard output.
func Print(a ...any) (n int, err error) {
	return Fprint(os.Stdout, a...)
}

// Sprint formats using the default formats for its operands and returns the resulting string.
func Sprint(a ...any) string {
	return printf.Sprint(a...)
}

// Append formats using the default formats for its operands, appends the result to
// the byte slice, and returns the updated slice.
func Append(b []byte, a ...any) []byte {
	return printf.Append(b, a...)
}

// Fprintln formats using the default formats for its operands and writes to w.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	return printf.Fprintln(w, a...)
}

// Println formats using the default formats for its operands and writes to standard output.
func Println(a ...any) (n int, err error) {
	return Fprintln(os.Stdout, a...)
}

// Sprintln formats using the default formats for its operands and returns the resulting string.
func Sprintln(a ...any) string {
	return printf.Sprintln(a...)
}

// Appendln formats using the default formats for its operands, appends the result
// to the byte slice, and returns the updated slice.
func Appendln(b []byte, a ...any) []byte {
	return printf.Appendln(b, a...)
}

// parsenum is used by scan.go in this package.
func parsenum(s string, start, end int) (num int, isnum bool, newi int) {
	return printf.Parsenum(s, start, end)
}
