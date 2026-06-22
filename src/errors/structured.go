// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import "runtime"

// Base is the embeddable structured-error payload (message, stack trace, inner link).
// Embed errors.Base in custom error types that also define Error() string;
// the type name Base avoids a field/method name clash with Error().
type Base struct {
	Message    string
	StackTrace StackTrace
	InnerError *Base

	link error // non-*Base target for Unwrap when InnerError is nil
}

// Error is an alias for Base. APIs such as NewError and WrapError return *Error.
type Error = Base

// StackTrace is a captured call stack (innermost frame first).
type StackTrace []StackFrame

// StackFrame identifies one stack frame.
type StackFrame struct {
	Function string // qualified name, e.g. "main.readFile"
	File     string
	Line     int
}

// Error returns the layer message only.
func (e *Base) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

// String serializes the full error chain and stack traces.
func (e *Base) String() string {
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
func (e *Base) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.InnerError != nil {
		return e.InnerError
	}
	return e.link
}

// Wrap prepends a layer with message and a fresh stack trace.
func (e *Base) Wrap(message string) error {
	return e.wrap(message)
}

// WrapError is like [Wrap] but returns [*Error].
func (e *Base) WrapError(message string) *Error {
	return e.wrap(message)
}

func (e *Base) wrap(message string) *Error {
	if e == nil {
		return newError(message)
	}
	return &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: e,
	}
}

// Wrap adds context and a stack trace layer around err.
func Wrap(err error, message string) error {
	return wrapError(err, message)
}

// WrapError is like [Wrap] but returns [*Error].
func WrapError(err error, message string) *Error {
	return wrapError(err, message)
}

func wrapError(err error, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Base); ok {
		return e.wrap(message)
	}
	return &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: bridgeError(err),
	}
}

// NewWrapped returns a structured error for fmt.Errorf with a single %w verb.
// message is the full formatted string; wrapped is the wrapped operand.
func NewWrapped(message string, wrapped error) error {
	return newWrapped(message, wrapped)
}

// NewWrappedError is like [NewWrapped] but returns [*Error].
func NewWrappedError(message string, wrapped error) *Error {
	return newWrapped(message, wrapped)
}

func newWrapped(message string, wrapped error) *Error {
	e := &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
	}
	setLink(e, wrapped)
	return e
}

func newError(message string) *Error {
	return &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
	}
}

// NewCustom returns a root error of type T with Message, StackTrace, and InnerError set
// on the embedded Error field. T must be a named type whose underlying type is
// struct{ Base } — that is, it embeds errors.Base and adds no other fields.
// When args are provided, format is interpreted like fmt.Sprintf.
func NewCustom[T ~struct{ Base }](format string, args ...any) *T {
	return &T{Base: *newError(formatMessage(format, args...))}
}

// InitCustom assigns Message, StackTrace, and InnerError on e from a new root error
// captured at the call site. Use this to initialize the embedded errors.Base field
// of a custom type that has extra domain fields, e.g. InitCustom(&myErr.Base, msg).
// When args are provided, format is interpreted like fmt.Sprintf.
// If e is nil, InitCustom does nothing.
func InitCustom(e *Base, format string, args ...any) {
	if e == nil {
		return
	}
	*e = *newError(formatMessage(format, args...))
}

func setLink(e *Base, err error) {
	if err == nil {
		return
	}
	if inner, ok := err.(*Base); ok {
		e.InnerError = inner
		return
	}
	e.link = err
}

// bridgeError wraps a non-*Base value for InnerError chains while
// preserving the original error for Unwrap/Is/As via link.
func bridgeError(err error) *Base {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Base); ok {
		return e
	}
	return &Base{
		Message: err.Error(),
		link:    err,
	}
}

const stackTraceDepth = 32

// CaptureStackTrace records the current goroutine stack, skipping internal
// frames in the errors, fmt, and log packages. Frames are ordered innermost
// caller first. The trace starts at the call site of the function that invoked
// CaptureStackTrace.
func CaptureStackTrace() StackTrace {
	var pcs [stackTraceDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}
	frames := runtime.CallersFrames(pcs[:])
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
	// Skip frames inside the errors, fmt, and log packages.
	if i := lastIndexByte(fn, '/'); i >= 0 {
		fn = fn[i+1:]
	}
	if hasPrefix(fn, "errors.") || hasPrefix(fn, "fmt.") || hasPrefix(fn, "log.") {
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
