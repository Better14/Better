// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

// formatFunc formats messages when args are provided. fmt.init registers the
// shared printf engine; until then a minimal fallback is used.
var formatFunc func(format string, args ...any) string

// RegisterFormatter installs the shared message formatter (printf.Sprintf).
// Called from fmt.init so errors does not import fmt or internal/printf.
func RegisterFormatter(fn func(format string, args ...any) string) {
	formatFunc = fn
}

func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	if formatFunc != nil {
		return formatFunc(format, args...)
	}
	return fallbackFormatMessage(format, args...)
}

func fallbackFormatMessage(format string, args ...any) string {
	var buf []byte
	argi := 0
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			buf = append(buf, format[i])
			continue
		}
		if i+1 >= len(format) {
			buf = append(buf, '%')
			break
		}
		i++
		verb := format[i]
		if verb == '%' {
			buf = append(buf, '%')
			continue
		}
		if argi >= len(args) {
			buf = append(buf, '%', verb)
			continue
		}
		arg := args[argi]
		argi++
		buf = appendFallbackVerb(buf, verb, arg)
	}
	return string(buf)
}

func appendFallbackVerb(buf []byte, verb byte, arg any) []byte {
	switch verb {
	case 's', 'v':
		switch v := arg.(type) {
		case string:
			return append(buf, v...)
		case error:
			if v == nil {
				return append(buf, "<nil>"...)
			}
			return append(buf, v.Error()...)
		default:
			return append(buf, '?')
		}
	case 'd':
		switch v := arg.(type) {
		case int:
			return append(buf, itoa(v)...)
		default:
			return append(buf, '?')
		}
	default:
		return append(buf, '%', verb)
	}
}
