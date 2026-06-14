// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"errors"
)

// ErrBadFormat is returned when a custom format string is invalid.
var ErrBadFormat = errors.New("time: bad custom format string")

// ErrParseCustom is returned when ParseCustom cannot parse the input.
var ErrParseCustom = errors.New("time: FormatCustom parse failed")

// ctoken is one element of a parsed .NET custom format string.
type ctoken struct {
	spec byte   // specifier letter, or 'L' for literal
	width int   // repeat count for specifiers
	lit  string // literal text when spec == 'L'
}

// FormatCustom formats t using a .NET custom date/time format string.
func (t Time) FormatCustom(format string) string {
	return t.FormatCustomLocale(format, "")
}

// FormatCustomLocale is FormatCustom with an explicit BCP 47 locale tag.
// An empty tag uses [DefaultLocale].
func (t Time) FormatCustomLocale(format string, tag string) string {
	loc := localeForTag(tag)
	tokens, err := parseCustomFormat(format)
	if err != nil {
		return ""
	}
	return formatCustomTokens(t, tokens, loc)
}

// ParseCustom parses value according to format in the given location.
func ParseCustom(format string, value string, loc *Location) (Time, error) {
	return ParseCustomLocale(format, value, "", loc)
}

// ParseCustomLocale parses value with an explicit locale tag.
func ParseCustomLocale(format string, value string, tag string, defaultLoc *Location) (Time, error) {
	cult := localeForTag(tag)
	tokens, err := parseCustomFormat(format)
	if err != nil {
		return Time{}, err
	}
	return parseCustomTokens(tokens, value, cult, defaultLoc)
}

func localeForTag(tag string) *Locale {
	if tag == "" {
		return DefaultLocale()
	}
	loc, err := LookupLocale(tag)
	if err != nil {
		return DefaultLocale()
	}
	return loc
}

func parseCustomFormat(s string) ([]ctoken, error) {
	if s == "" {
		return nil, ErrBadFormat
	}
	var tokens []ctoken
	i := 0
	for i < len(s) {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) {
				return nil, ErrBadFormat
			}
			tokens = append(tokens, ctoken{spec: 'L', lit: s[i+1 : i+2]})
			i += 2
		case '\'':
			lit, ni, err := readQuotedLiteral(s, i+1, '\'')
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, ctoken{spec: 'L', lit: lit})
			i = ni
		case '"':
			lit, ni, err := readQuotedLiteral(s, i+1, '"')
			if err != nil {
				return nil, ErrBadFormat
			}
			tokens = append(tokens, ctoken{spec: 'L', lit: lit})
			i = ni
		case '%':
			if i+1 >= len(s) {
				return nil, ErrBadFormat
			}
			c := s[i+1]
			if !isCustomSpec(c) {
				return nil, ErrBadFormat
			}
			tokens = append(tokens, ctoken{spec: c, width: 1})
			i += 2
		case ':':
			tokens = append(tokens, ctoken{spec: ':'})
			i++
		case '/':
			tokens = append(tokens, ctoken{spec: '/'})
			i++
		default:
			c := s[i]
			if isCustomSpec(c) {
				w := 1
				for i+w < len(s) && s[i+w] == c {
					w++
				}
				w = capCustomWidth(c, w)
				tokens = append(tokens, ctoken{spec: c, width: w})
				i += w
			} else {
				j := i + 1
				for j < len(s) && !isCustomFormatStart(s, j) {
					j++
				}
				tokens = append(tokens, ctoken{spec: 'L', lit: s[i:j]})
				i = j
			}
		}
	}
	return tokens, nil
}

func readQuotedLiteral(s string, i int, quote byte) (lit string, next int, err error) {
	var b []byte
	for i < len(s) {
		if s[i] == quote {
			if i+1 < len(s) && s[i+1] == quote {
				b = append(b, quote)
				i += 2
				continue
			}
			return string(b), i + 1, nil
		}
		b = append(b, s[i])
		i++
	}
	return "", 0, ErrBadFormat
}

func isCustomSpec(c byte) bool {
	switch c {
	case 'd', 'f', 'F', 'g', 'h', 'H', 'K', 'm', 'M', 's', 't', 'y', 'z':
		return true
	}
	return false
}

func capCustomWidth(c byte, w int) int {
	switch c {
	case 'd', 'M':
		if w > 4 {
			return 4
		}
	case 'f', 'F':
		if w > 7 {
			return 7
		}
	case 'g':
		if w > 2 {
			return 2
		}
	case 'h', 'H', 'm', 's', 't':
		if w > 2 {
			return 2
		}
	case 'y':
		if w > 5 {
			return 5
		}
	case 'z':
		if w > 3 {
			return 3
		}
	}
	return w
}

func isCustomFormatStart(s string, i int) bool {
	switch s[i] {
	case '\\', '\'', '"', '%', ':', '/':
		return true
	}
	return isCustomSpec(s[i])
}

func formatCustomTokens(t Time, tokens []ctoken, loc *Locale) string {
	t = t.In(t.Location())
	var b []byte
	for i := 0; i < len(tokens); {
		if run, ni := slashDateRun(tokens, i); run != nil && loc.DateOrder == dateOrderDMY {
			b = appendSlashDateDMY(b, t, run, loc)
			i = ni
			continue
		}
		b = appendCustomToken(b, t, tokens[i], loc)
		i++
	}
	return string(b)
}

// slashDateRun detects MM/dd/yyyy-style runs (d/M/y with '/' separators).
func slashDateRun(tokens []ctoken, i int) (run []ctoken, next int) {
	if i >= len(tokens) || tokens[i].spec != 'M' && tokens[i].spec != 'd' && tokens[i].spec != 'y' {
		return nil, i
	}
	j := i
	for j < len(tokens) {
		switch tokens[j].spec {
		case 'M', 'd', 'y':
			j++
		case '/':
			if j+1 >= len(tokens) {
				return nil, i
			}
			next := tokens[j+1].spec
			if next != 'M' && next != 'd' && next != 'y' {
				return nil, i
			}
			j++
		default:
			goto done
		}
	}
done:
	if j <= i+1 {
		return nil, i
	}
	return tokens[i:j], j
}

func appendSlashDateDMY(b []byte, t Time, run []ctoken, loc *Locale) []byte {
	var day, month, year []ctoken
	for _, tok := range run {
		switch tok.spec {
		case 'd':
			day = append(day, tok)
		case 'M':
			month = append(month, tok)
		case 'y':
			year = append(year, tok)
		}
	}
	emit := func(parts ...[]ctoken) []byte {
		first := true
		for _, group := range parts {
			for _, tok := range group {
				if !first {
					b = append(b, loc.DateSeparator...)
				}
				first = false
				b = appendCustomToken(b, t, tok, loc)
			}
		}
		return b
	}
	return emit(day, month, year)
}

func appendCustomToken(b []byte, t Time, tok ctoken, loc *Locale) []byte {
	if tok.spec == 'L' {
		return append(b, tok.lit...)
	}
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	nsec := t.Nanosecond()
	wd := t.Weekday()

	switch tok.spec {
	case 'd':
		switch tok.width {
		case 1:
			b = appendInt(b, day, 0)
		case 2:
			b = appendInt(b, day, 2)
		case 3:
			b = append(b, loc.WeekdayAbbr[wd]...)
		case 4:
			b = append(b, loc.WeekdayNames[wd]...)
		}
	case 'f':
		b = append(b, formatFracDigits(nsec, tok.width, false)...)
	case 'F':
		b = append(b, formatFracDigits(nsec, tok.width, true)...)
	case 'g':
		b = append(b, loc.EraDesignator...)
	case 'h':
		hr, _ := hour12(hour)
		if tok.width == 1 {
			b = appendInt(b, hr, 0)
		} else {
			b = appendInt(b, hr, 2)
		}
	case 'H':
		if tok.width == 1 {
			b = appendInt(b, hour, 0)
		} else {
			b = appendInt(b, hour, 2)
		}
	case 'K':
		b = append(b, formatZoneK(t)...)
	case 'm':
		if tok.width == 1 {
			b = appendInt(b, min, 0)
		} else {
			b = appendInt(b, min, 2)
		}
	case 'M':
		switch tok.width {
		case 1:
			b = appendInt(b, int(month), 0)
		case 2:
			b = appendInt(b, int(month), 2)
		case 3:
			b = append(b, loc.MonthAbbr[month-1]...)
		case 4:
			b = append(b, loc.MonthNames[month-1]...)
		}
	case 's':
		if tok.width == 1 {
			b = appendInt(b, sec, 0)
		} else {
			b = appendInt(b, sec, 2)
		}
	case 't':
		designator := ampmDesignator(hour, loc)
		if tok.width == 1 {
			if len(designator) > 0 {
				_, size := utf8DecodeRuneInString(designator)
				b = append(b, designator[:size]...)
			}
		} else {
			b = append(b, designator...)
		}
	case 'y':
		b = append(b, formatCustomYear(year, tok.width)...)
	case 'z':
		b = append(b, formatZoneOffset(t, tok.width)...)
	case ':':
		b = append(b, loc.TimeSeparator...)
	case '/':
		b = append(b, loc.DateSeparator...)
	}
	return b
}

func hour12(hour int) (int, bool) {
	pm := hour >= 12
	hr := hour % 12
	if hr == 0 {
		hr = 12
	}
	return hr, pm
}

func ampmDesignator(hour int, loc *Locale) string {
	if hour < 12 {
		return loc.AMDesignator
	}
	return loc.PMDesignator
}

func formatCustomYear(year int, width int) []byte {
	switch width {
	case 1:
		v := year % 100
		if v == 0 && year >= 100 {
			return appendInt(nil, 0, 0)
		}
		return appendInt(nil, v, 0)
	case 2:
		return appendInt(nil, year%100, 2)
	case 3:
		if year < 1000 {
			return appendInt(nil, year, 3)
		}
		return appendInt(nil, year, 0)
	case 4:
		return appendInt(nil, year, 4)
	default:
		return appendInt(nil, year, 5)
	}
}

func formatFracDigits(nsec int, width int, trimTrailing bool) []byte {
	if width <= 0 || width > 7 {
		return nil
	}
	div := 1
	for i := 0; i < 9-width; i++ {
		div *= 10
	}
	v := nsec / div
	b := appendInt(nil, v, width)
	if trimTrailing {
		for len(b) > 0 && b[len(b)-1] == '0' {
			b = b[:len(b)-1]
		}
	}
	return b
}

func formatZoneK(t Time) []byte {
	if t.Location() == UTC {
		return []byte{'Z'}
	}
	return formatZoneOffset(t, 3)
}

func formatZoneOffset(t Time, width int) []byte {
	_, off := t.Zone()
	sign := byte('+')
	if off < 0 {
		sign = '-'
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	switch width {
	case 1:
		return appendInt(append([]byte{sign}, '0'), h, 0)
	case 2:
		return appendInt(append([]byte{sign}, '0'), h, 2)
	default:
		b := appendInt(append([]byte{sign}, '0'), h, 2)
		b = append(b, ':')
		return appendInt(b, m, 2)
	}
}

func utf8DecodeRuneInString(s string) (r rune, size int) {
	if len(s) == 0 {
		return 0, 0
	}
	c := s[0]
	if c < 0x80 {
		return rune(c), 1
	}
	if c < 0xE0 {
		if len(s) < 2 {
			return rune(c), 1
		}
		return rune(c&0x1F)<<6 | rune(s[1]&0x3F), 2
	}
	if c < 0xF0 {
		if len(s) < 3 {
			return rune(c), 1
		}
		return rune(c&0x0F)<<12 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	if len(s) < 4 {
		return rune(c), 1
	}
	return rune(c&0x07)<<18 | rune(s[1]&0x3F)<<12 | rune(s[2]&0x3F)<<6 | rune(s[3]&0x3F), 4
}

type parseFields struct {
	year   int
	month  int
	day    int
	hour   int
	min    int
	sec    int
	nsec   int
	zone   *Location
	hasYMD bool
	hasHMS bool
	hasZone bool
	pmSet  bool
	isPM   bool
}

func parseCustomTokens(tokens []ctoken, value string, loc *Locale, defaultLoc *Location) (Time, error) {
	var f parseFields
	f.month = 1
	f.day = 1
	if defaultLoc != nil {
		f.zone = defaultLoc
	} else {
		f.zone = UTC
	}
	s := value
	for _, tok := range tokens {
		if len(s) == 0 && tok.spec != 'L' {
			return Time{}, ErrParseCustom
		}
		var err error
		s, err = parseCustomToken(s, tok, loc, &f)
		if err != nil {
			return Time{}, ErrParseCustom
		}
	}
	if s != "" {
		return Time{}, ErrParseCustom
	}
	if !f.hasYMD && !f.hasHMS {
		return Time{}, ErrParseCustom
	}
	if !f.hasYMD {
		now := Now()
		y, m, d := now.Date()
		f.year, f.month, f.day = y, int(m), d
	}
	if f.pmSet {
		if f.hour < 12 && f.isPM {
			f.hour += 12
		} else if f.hour == 12 && !f.isPM {
			f.hour = 0
		}
	}
	return Date(f.year, Month(f.month), f.day, f.hour, f.min, f.sec, f.nsec, f.zone), nil
}

func parseCustomToken(s string, tok ctoken, loc *Locale, f *parseFields) (rest string, err error) {
	if tok.spec == 'L' {
		if len(s) < len(tok.lit) || s[:len(tok.lit)] != tok.lit {
			return s, errBad
		}
		return s[len(tok.lit):], nil
	}

	switch tok.spec {
	case 'd':
		switch tok.width {
		case 1, 2:
			w := tok.width
			if w == 1 {
				w = 0
			}
			n, rem, e := parseDigits(s, 1, 2)
			if e != nil {
				return s, e
			}
			f.day = n
			f.hasYMD = true
			return rem, nil
		case 3:
			i, rem, e := lookupLocaleName(loc.weekdayNames(false), s)
			if e != nil {
				return s, e
			}
			_ = i
			return rem, nil
		case 4:
			i, rem, e := lookupLocaleName(loc.weekdayNames(true), s)
			if e != nil {
				return s, e
			}
			_ = i
			return rem, nil
		}
	case 'f', 'F':
		n, rem, e := parseDigits(s, tok.width, tok.width)
		if e != nil {
			return s, e
		}
		scale := 1
		for i := 0; i < 9-tok.width; i++ {
			scale *= 10
		}
		f.nsec += n * scale
		f.hasHMS = true
		return rem, nil
	case 'g':
		if len(s) < len(loc.EraDesignator) || !match(s[:len(loc.EraDesignator)], loc.EraDesignator) {
			return s, errBad
		}
		return s[len(loc.EraDesignator):], nil
	case 'h':
		n, rem, e := parseDigits(s, 1, 2)
		if e != nil || n < 1 || n > 12 {
			return s, errBad
		}
		f.hour = n
		f.hasHMS = true
		return rem, nil
	case 'H':
		n, rem, e := parseDigits(s, 1, 2)
		if e != nil || n > 23 {
			return s, errBad
		}
		f.hour = n
		f.hasHMS = true
		return rem, nil
	case 'K':
		if len(s) > 0 && s[0] == 'Z' {
			f.zone = UTC
			f.hasZone = true
			return s[1:], nil
		}
		zone, rem, e := parseNumericZone(s)
		if e != nil {
			return s, e
		}
		f.zone = zone
		f.hasZone = true
		return rem, nil
	case 'm':
		n, rem, e := parseDigits(s, 1, 2)
		if e != nil || n > 59 {
			return s, errBad
		}
		f.min = n
		f.hasHMS = true
		return rem, nil
	case 'M':
		switch tok.width {
		case 1, 2:
			maxW := 2
			if tok.width == 1 {
				maxW = 2
			}
			n, rem, e := parseDigits(s, 1, maxW)
			if e != nil || n < 1 || n > 12 {
				return s, errBad
			}
			f.month = n
			f.hasYMD = true
			return rem, nil
		case 3:
			i, rem, e := lookupLocaleName(loc.monthNames(false), s)
			if e != nil {
				return s, e
			}
			f.month = i + 1
			f.hasYMD = true
			return rem, nil
		case 4:
			i, rem, e := lookupLocaleName(loc.monthNames(true), s)
			if e != nil {
				return s, e
			}
			f.month = i + 1
			f.hasYMD = true
			return rem, nil
		}
	case 's':
		n, rem, e := parseDigits(s, 1, 2)
		if e != nil || n > 59 {
			return s, errBad
		}
		f.sec = n
		f.hasHMS = true
		return rem, nil
	case 't':
		designator := loc.PMDesignator
		if len(designator) == 0 {
			designator = loc.AMDesignator
		}
		if tok.width == 2 {
			if len(s) >= len(loc.AMDesignator) && match(s[:len(loc.AMDesignator)], loc.AMDesignator) {
				f.pmSet, f.isPM = true, false
				return s[len(loc.AMDesignator):], nil
			}
			if len(s) >= len(loc.PMDesignator) && match(s[:len(loc.PMDesignator)], loc.PMDesignator) {
				f.pmSet, f.isPM = true, true
				return s[len(loc.PMDesignator):], nil
			}
		} else {
			if len(loc.AMDesignator) > 0 && len(s) >= 1 && match(s[:1], loc.AMDesignator[:1]) {
				f.pmSet, f.isPM = true, false
				return s[1:], nil
			}
			if len(loc.PMDesignator) > 0 && len(s) >= 1 && match(s[:1], loc.PMDesignator[:1]) {
				f.pmSet, f.isPM = true, true
				return s[1:], nil
			}
		}
		return s, errBad
	case 'y':
		n, rem, e := parseYear(s, tok.width)
		if e != nil {
			return s, e
		}
		f.year = n
		f.hasYMD = true
		return rem, nil
	case 'z':
		zone, rem, e := parseZoneWidth(s, tok.width)
		if e != nil {
			return s, e
		}
		f.zone = zone
		f.hasZone = true
		return rem, nil
	case ':':
		if len(s) < len(loc.TimeSeparator) || s[:len(loc.TimeSeparator)] != loc.TimeSeparator {
			return s, errBad
		}
		return s[len(loc.TimeSeparator):], nil
	case '/':
		if len(s) < len(loc.DateSeparator) || s[:len(loc.DateSeparator)] != loc.DateSeparator {
			return s, errBad
		}
		return s[len(loc.DateSeparator):], nil
	}
	return s, errBad
}

func parseDigits(s string, min, max int) (n int, rest string, err error) {
	if len(s) == 0 {
		return 0, s, errBad
	}
	q, rem, err := leadingInt(s)
	if err != nil {
		return 0, s, err
	}
	digits := len(s) - len(rem)
	if digits < min || digits > max {
		return 0, s, errBad
	}
	return int(q), rem, nil
}

func parseYear(s string, width int) (year int, rest string, err error) {
	switch width {
	case 1, 2:
		n, rem, e := parseDigits(s, width, width)
		if e != nil {
			return 0, s, e
		}
		if width == 2 {
			if n >= 0 && n <= 69 {
				n += 2000
			} else if n >= 70 && n <= 99 {
				n += 1900
			}
		} else {
			now := Now().Year()
			century := (now / 100) * 100
			n = century + n
			if n > now+50 {
				n -= 100
			}
		}
		return n, rem, nil
	case 3:
		n, rem, e := parseDigits(s, 3, 4)
		if e != nil {
			return 0, s, e
		}
		return n, rem, nil
	default:
		n, rem, e := parseDigits(s, width, width)
		if e != nil {
			return 0, s, e
		}
		return n, rem, nil
	}
}

func parseNumericZone(s string) (*Location, string, error) {
	if len(s) == 0 {
		return nil, s, errBad
	}
	sign := 1
	if s[0] == '-' {
		sign = -1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}
	h, rem, err := parseDigits(s, 1, 2)
	if err != nil {
		return nil, s, err
	}
	off := h * 3600 * sign
	if len(rem) > 0 && rem[0] == ':' {
		m, rem2, err := parseDigits(rem[1:], 2, 2)
		if err != nil {
			return nil, s, err
		}
		off += m * 60 * sign
		rem = rem2
	}
	name := zoneNameForOffset(off)
	return FixedZone(name, off), rem, nil
}

func parseZoneWidth(s string, width int) (*Location, string, error) {
	if len(s) == 0 {
		return nil, s, errBad
	}
	sign := 1
	if s[0] == '-' {
		sign = -1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	} else {
		return nil, s, errBad
	}
	switch width {
	case 1, 2:
		maxW := width
		if width == 1 {
			maxW = 2
		}
		h, rem, err := parseDigits(s, 1, maxW)
		if err != nil {
			return nil, s, err
		}
		off := h * 3600 * sign
		return FixedZone(zoneNameForOffset(off), off), rem, nil
	default:
		h, rem, err := parseDigits(s, 1, 2)
		if err != nil {
			return nil, s, err
		}
		if len(rem) == 0 || rem[0] != ':' {
			return nil, s, errBad
		}
		m, rem2, err := parseDigits(rem[1:], 2, 2)
		if err != nil {
			return nil, s, err
		}
		off := h*3600*sign + m*60*sign
		return FixedZone(zoneNameForOffset(off), off), rem2, nil
	}
}

func zoneNameForOffset(off int) string {
	sign := byte('+')
	if off < 0 {
		sign = '-'
		off = -off
	}
	h := off / 3600
	m := (off % 3600) / 60
	b := appendInt(append([]byte{'U', 'T', 'C', sign}, '0'), h, 2)
	if m > 0 {
		b = append(b, ':')
		b = appendInt(b, m, 2)
	}
	return string(b)
}
