// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"fmt"
	"runtime"
)

// Error is a structured error with message, stack trace, and optional inner error.
type Error struct {
	Message    string
	StackTrace StackTrace
	InnerError *Error

	link error // non-*Error target for Unwrap when InnerError is nil
}

// StackTrace is a captured call stack (innermost frame first).
type StackTrace []StackFrame

// StackFrame identifies one stack frame.
type StackFrame struct {
	Function string // qualified name, e.g. "main.readFile"
	File     string
	Line     int
}

// Error returns the layer message only.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

// String serializes the full error chain and stack traces.
func (e *Error) String() string {
	if e == nil {
		return "<nil>"
	}
	var b []byte
	if e.InnerError != nil {
		b = append(b, e.InnerError.String()...)
		if len(b) > 0 {
			b = append(b, '\n')
		}
	}
	b = append(b, e.Message...)
	if len(e.StackTrace) > 0 {
		b = append(b, '\n')
		b = append(b, e.StackTrace.String()...)
	}
	return string(b)
}

// String formats every frame, one per line.
func (st StackTrace) String() string {
	if len(st) == 0 {
		return ""
	}
	n := 0
	for i, f := range st {
		if i > 0 {
			n++
		}
		n += len(f.Function) + len(f.File) + len(itoa(f.Line)) + 4
	}
	b := make([]byte, 0, n)
	for i, f := range st {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, f.Function...)
		b = append(b, " ("...)
		b = append(b, f.File...)
		b = append(b, ':')
		b = append(b, itoa(f.Line)...)
		b = append(b, ')')
	}
	return string(b)
}

// Unwrap returns the next error in the chain.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.InnerError != nil {
		return e.InnerError
	}
	return e.link
}

// Wrap prepends a layer with message and a fresh stack trace.
func (e *Error) Wrap(message string) *Error {
	if e == nil {
		return newError(message)
	}
	return &Error{
		Message:    message,
		StackTrace: captureStackTrace(),
		InnerError: e,
	}
}

// Wrap adds context and a stack trace layer around err.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e.Wrap(message)
	}
	return &Error{
		Message:    message,
		StackTrace: captureStackTrace(),
		InnerError: bridgeError(err),
	}
}

// NewWrapped returns a structured error for fmt.Errorf with a single %w verb.
// message is the full formatted string; wrapped is the wrapped operand.
func NewWrapped(message string, wrapped error) error {
	e := &Error{
		Message:    message,
		StackTrace: captureStackTrace(),
	}
	setLink(e, wrapped)
	return e
}

func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func newError(message string) *Error {
	return &Error{
		Message:    message,
		StackTrace: captureStackTrace(),
	}
}

// NewCustom returns a root error of type T with Message, StackTrace, and InnerError set
// on the embedded Error field. T must be a named type whose underlying type is
// struct{ Error } — that is, it embeds errors.Error and adds no other fields.
// When args are provided, format is interpreted like fmt.Sprintf.
func NewCustom[T ~struct{ Error }](format string, args ...any) *T {
	return &T{Error: *newError(formatMessage(format, args...))}
}

// NewCustom assigns Message, StackTrace, and InnerError on e from a new root error
// captured at the call site. Use this to initialize the embedded errors.Error field
// of a custom type that has extra domain fields, e.g. NewCustom(&myErr.Error, msg).
// When args are provided, format is interpreted like fmt.Sprintf.
// If e is nil, NewCustom does nothing.
func NewCustom(e *Error, format string, args ...any) {
	if e == nil {
		return
	}
	*e = *newError(formatMessage(format, args...))
}

func setLink(e *Error, err error) {
	if err == nil {
		return
	}
	if inner, ok := err.(*Error); ok {
		e.InnerError = inner
		return
	}
	e.link = err
}

// bridgeError wraps a non-*Error value for InnerError chains while
// preserving the original error for Unwrap/Is/As via link.
func bridgeError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{
		Message: err.Error(),
		link:    err,
	}
}

const stackTraceDepth = 32

func captureStackTrace() StackTrace {
	var pcs [stackTraceDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}
	frames := runtime.CallersFrames(pcs[:n])
	var trace StackTrace
	for {
		frame, more := frames.Next()
		if !skipStackFrame(frame.Function) {
			trace = append(trace, StackFrame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			})
		}
		if !more {
			break
		}
	}
	return trace
}

func skipStackFrame(fn string) bool {
	if fn == "" {
		return true
	}
	if hasPrefix(fn, "runtime.") {
		return true
	}
	// Skip frames inside the errors and fmt packages.
	if i := lastIndexByte(fn, '/'); i >= 0 {
		fn = fn[i+1:]
	}
	if hasPrefix(fn, "errors.") || hasPrefix(fn, "fmt.") {
		return true
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var b [20]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
