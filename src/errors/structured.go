// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import "runtime"

// Layer is the embeddable structured-error payload (message, stack trace, inner link).
// Embed errors.Layer in custom error types that also define Error() string;
// the type name Layer avoids a field/method name clash with Error().
type Layer struct {
	Message    string
	StackTrace StackTrace
	InnerError *Layer

	link error // non-*Layer target for Unwrap when InnerError is nil
}

// Error is an alias for Layer. APIs such as New and Wrap return *Error.
type Error = Layer

// StackTrace is a captured call stack (innermost frame first).
type StackTrace []StackFrame

// StackFrame identifies one stack frame.
type StackFrame struct {
	Function string // qualified name, e.g. "main.readFile"
	File     string
	Line     int
}

// Error returns the layer message only.
func (e *Layer) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

// String serializes the full error chain and stack traces.
func (e *Layer) String() string {
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
func (e *Layer) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.InnerError != nil {
		return e.InnerError
	}
	return e.link
}

// Wrap prepends a layer with message and a fresh stack trace.
func (e *Layer) Wrap(message string) *Layer {
	if e == nil {
		return newError(message)
	}
	return &Layer{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: e,
	}
}

// Wrap adds context and a stack trace layer around err.
func Wrap(err error, message string) *Layer {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Layer); ok {
		return e.Wrap(message)
	}
	return &Layer{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: bridgeError(err),
	}
}

// NewWrapped returns a structured error for fmt.Errorf with a single %w verb.
// message is the full formatted string; wrapped is the wrapped operand.
func NewWrapped(message string, wrapped error) error {
	e := &Layer{
		Message:    message,
		StackTrace: CaptureStackTrace(),
	}
	setLink(e, wrapped)
	return e
}

func newError(message string) *Error {
	return &Layer{
		Message:    message,
		StackTrace: CaptureStackTrace(),
	}
}

// NewCustom returns a root error of type T with Message, StackTrace, and InnerError set
// on the embedded Error field. T must be a named type whose underlying type is
// struct{ Layer } — that is, it embeds errors.Layer and adds no other fields.
// When args are provided, format is interpreted like fmt.Sprintf.
func NewCustom[T ~struct{ Layer }](format string, args ...any) *T {
	return &T{Layer: *newError(formatMessage(format, args...))}
}

// InitCustom assigns Message, StackTrace, and InnerError on e from a new root error
// captured at the call site. Use this to initialize the embedded errors.Layer field
// of a custom type that has extra domain fields, e.g. InitCustom(&myErr.Layer, msg).
// When args are provided, format is interpreted like fmt.Sprintf.
// If e is nil, InitCustom does nothing.
func InitCustom(e *Layer, format string, args ...any) {
	if e == nil {
		return
	}
	*e = *newError(formatMessage(format, args...))
}

func setLink(e *Layer, err error) {
	if err == nil {
		return
	}
	if inner, ok := err.(*Layer); ok {
		e.InnerError = inner
		return
	}
	e.link = err
}

// bridgeError wraps a non-*Layer value for InnerError chains while
// preserving the original error for Unwrap/Is/As via link.
func bridgeError(err error) *Layer {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Layer); ok {
		return e
	}
	return &Layer{
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
