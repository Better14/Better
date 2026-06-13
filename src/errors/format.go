// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

func formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
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
		buf = appendVerb(buf, verb, arg)
	}
	return string(buf)
}

func appendVerb(buf []byte, verb byte, arg any) []byte {
	switch verb {
	case 's':
		return append(buf, appendString(arg)...)
	case 'd':
		return appendInt(buf, arg)
	case 'v':
		return append(buf, appendValue(arg)...)
	case 'q':
		return appendQuoted(buf, arg)
	default:
		return append(buf, '%', verb)
	}
}

func appendString(arg any) []byte {
	switch v := arg.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case error:
		if v == nil {
			return []byte("<nil>")
		}
		return []byte(v.Error())
	default:
		return appendValue(arg)
	}
}

func appendInt(buf []byte, arg any) []byte {
	switch v := arg.(type) {
	case int:
		return append(buf, itoa(v)...)
	case int8:
		return append(buf, itoa(int(v))...)
	case int16:
		return append(buf, itoa(int(v))...)
	case int32:
		return append(buf, itoa(int(v))...)
	case int64:
		return append(buf, itoa64(v)...)
	case uint:
		return append(buf, uitoa(uint64(v))...)
	case uint8:
		return append(buf, uitoa(uint64(v))...)
	case uint16:
		return append(buf, uitoa(uint64(v))...)
	case uint32:
		return append(buf, uitoa(uint64(v))...)
	case uint64:
		return append(buf, uitoa(v)...)
	case uintptr:
		return append(buf, uitoa(uint64(v))...)
	default:
		return append(buf, appendValue(arg)...)
	}
}

func appendValue(arg any) []byte {
	switch v := arg.(type) {
	case nil:
		return []byte("<nil>")
	case string:
		return []byte(v)
	case bool:
		if v {
			return []byte("true")
		}
		return []byte("false")
	case int:
		return []byte(itoa(v))
	case int8:
		return []byte(itoa(int(v)))
	case int16:
		return []byte(itoa(int(v)))
	case int32:
		return []byte(itoa(int(v)))
	case int64:
		return []byte(itoa64(v))
	case uint:
		return []byte(uitoa(uint64(v)))
	case uint8:
		return []byte(uitoa(uint64(v)))
	case uint16:
		return []byte(uitoa(uint64(v)))
	case uint32:
		return []byte(uitoa(uint64(v)))
	case uint64:
		return []byte(uitoa(v))
	case uintptr:
		return []byte(uitoa(uint64(v)))
	case error:
		if v == nil {
			return []byte("<nil>")
		}
		return []byte(v.Error())
	default:
		return []byte("?")
	}
}

func appendQuoted(buf []byte, arg any) []byte {
	s := appendString(arg)
	buf = append(buf, '"')
	for _, c := range s {
		if c == '"' || c == '\\' {
			buf = append(buf, '\\')
		}
		buf = append(buf, c)
	}
	return append(buf, '"')
}

func itoa64(i int64) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + uitoa(uint64(-i))
	}
	return uitoa(uint64(i))
}

func uitoa(u uint64) string {
	if u == 0 {
		return "0"
	}
	var b [20]byte
	n := len(b)
	for u > 0 {
		n--
		b[n] = byte('0' + u%10)
		u /= 10
	}
	return string(b[n:])
}
