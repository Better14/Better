// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package printf

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	p := newPrinter()
	p.doPrintf(format, a)
	s := string(p.buf)
	p.free()
	return s
}

// Appendf formats according to a format specifier, appends the result to b,
// and returns the updated slice.
func Appendf(b []byte, format string, a ...any) []byte {
	p := newPrinter()
	p.doPrintf(format, a)
	b = append(b, p.buf...)
	p.free()
	return b
}

// Printer formats values into a buffer. It is pooled; call Free when done.
type Printer struct {
	*pp
}

// New returns a pooled printer.
func New() *Printer {
	return &Printer{pp: newPrinter()}
}

// Free returns the printer to the pool.
func (p *Printer) Free() {
	p.pp.free()
}

// SetWrapErrs enables recording of %w operand indices.
func (p *Printer) SetWrapErrs(v bool) {
	p.wrapErrs = v
}

// Printf formats into the printer buffer.
func (p *Printer) Printf(format string, a ...any) {
	p.doPrintf(format, a)
}

// String returns the formatted output.
func (p *Printer) String() string {
	return string(p.buf)
}

// Bytes returns the formatted output.
func (p *Printer) Bytes() []byte {
	return p.buf
}

// WrappedErrs returns indices of %w operands from the last Printf.
func (p *Printer) WrappedErrs() []int {
	return p.wrappedErrs
}

// Reordered reports whether the last Printf used argument reordering.
func (p *Printer) Reordered() bool {
	return p.reordered
}

// Parsenum converts ASCII to integer. num is 0 (and isnum is false) if no number present.
func Parsenum(s string, start, end int) (num int, isnum bool, newi int) {
	return parsenum(s, start, end)
}
