
package time

import (
	"errors"
	"internal/stringslite"
	"syscall"
)

// Locale holds culture-specific date and time formatting data (DateTimeFormatInfo-like).
type Locale struct {
	Tag           string
	MonthNames    [12]string
	MonthAbbr     [12]string
	WeekdayNames  [7]string
	WeekdayAbbr   [7]string
	AMDesignator  string
	PMDesignator  string
	DateSeparator string
	TimeSeparator string
	EraDesignator string
	DateOrder     dateOrder // component order for slash-separated short dates
}

// dateOrder controls locale-specific reordering of MM/dd/yyyy-style patterns.
type dateOrder byte

const (
	dateOrderMDY dateOrder = iota // month/day/year (default)
	dateOrderDMY                  // day/month/year (e.g. de-DE)
)

var (
	errUnknownLocale = errors.New("time: unknown locale")
	builtinLocales   map[string]*Locale
)

func init() {
	builtinLocales = map[string]*Locale{
		"en-US": {
			Tag: "en-US",
			MonthNames: [12]string{
				"January", "February", "March", "April", "May", "June",
				"July", "August", "September", "October", "November", "December",
			},
			MonthAbbr: [12]string{
				"Jan", "Feb", "Mar", "Apr", "May", "Jun",
				"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
			},
			WeekdayNames: [7]string{
				"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
			},
			WeekdayAbbr: [7]string{
				"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat",
			},
			AMDesignator:  "AM",
			PMDesignator:  "PM",
			DateSeparator: "/",
			TimeSeparator: ":",
			EraDesignator: "A.D.",
		},
		"de-DE": {
			Tag: "de-DE",
			MonthNames: [12]string{
				"Januar", "Februar", "März", "April", "Mai", "Juni",
				"Juli", "August", "September", "Oktober", "November", "Dezember",
			},
			MonthAbbr: [12]string{
				"Jan", "Feb", "Mär", "Apr", "Mai", "Jun",
				"Jul", "Aug", "Sep", "Okt", "Nov", "Dez",
			},
			WeekdayNames: [7]string{
				"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag",
			},
			WeekdayAbbr: [7]string{
				"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa",
			},
			AMDesignator:  "",
			PMDesignator:  "",
			DateSeparator: ".",
			TimeSeparator: ":",
			EraDesignator: "n. Chr.",
			DateOrder:     dateOrderDMY,
		},
		"fr-FR": {
			Tag: "fr-FR",
			MonthNames: [12]string{
				"janvier", "février", "mars", "avril", "mai", "juin",
				"juillet", "août", "septembre", "octobre", "novembre", "décembre",
			},
			MonthAbbr: [12]string{
				"janv.", "févr.", "mars", "avr.", "mai", "juin",
				"juil.", "août", "sept.", "oct.", "nov.", "déc.",
			},
			WeekdayNames: [7]string{
				"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi",
			},
			WeekdayAbbr: [7]string{
				"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam.",
			},
			AMDesignator:  "",
			PMDesignator:  "",
			DateSeparator: "/",
			TimeSeparator: ":",
			EraDesignator: "ap. J.-C.",
		},
		"ru-RU": {
			Tag: "ru-RU",
			MonthNames: [12]string{
				"января", "февраля", "марта", "апреля", "мая", "июня",
				"июля", "августа", "сентября", "октября", "ноября", "декабря",
			},
			MonthAbbr: [12]string{
				"янв.", "фев.", "мар.", "апр.", "мая", "июн.",
				"июл.", "авг.", "сен.", "окт.", "ноя.", "дек.",
			},
			WeekdayNames: [7]string{
				"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота",
			},
			WeekdayAbbr: [7]string{
				"вс", "пн", "вт", "ср", "чт", "пт", "сб",
			},
			AMDesignator:  "",
			PMDesignator:  "",
			DateSeparator: ".",
			TimeSeparator: ":",
			EraDesignator: "н.э.",
		},
		"it-IT": {
			Tag: "it-IT",
			MonthNames: [12]string{
				"gennaio", "febbraio", "marzo", "aprile", "maggio", "giugno",
				"luglio", "agosto", "settembre", "ottobre", "novembre", "dicembre",
			},
			MonthAbbr: [12]string{
				"gen", "feb", "mar", "apr", "mag", "giu",
				"lug", "ago", "set", "ott", "nov", "dic",
			},
			WeekdayNames: [7]string{
				"domenica", "lunedì", "martedì", "mercoledì", "giovedì", "venerdì", "sabato",
			},
			WeekdayAbbr: [7]string{
				"dom", "lun", "mar", "mer", "gio", "ven", "sab",
			},
			AMDesignator:  "",
			PMDesignator:  "",
			DateSeparator: "/",
			TimeSeparator: ".",
			EraDesignator: "d.C.",
		},
		"ja-JP": {
			Tag: "ja-JP",
			MonthNames: [12]string{
				"1月", "2月", "3月", "4月", "5月", "6月",
				"7月", "8月", "9月", "10月", "11月", "12月",
			},
			MonthAbbr: [12]string{
				"1", "2", "3", "4", "5", "6",
				"7", "8", "9", "10", "11", "12",
			},
			WeekdayNames: [7]string{
				"日曜日", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日",
			},
			WeekdayAbbr: [7]string{
				"日", "月", "火", "水", "木", "金", "土",
			},
			AMDesignator:  "午前",
			PMDesignator:  "午後",
			DateSeparator: "/",
			TimeSeparator: ":",
			EraDesignator: "西暦",
		},
	}
}

// DefaultLocale returns the locale from LC_TIME or LANG, or en-US.
func DefaultLocale() *Locale {
	if tag, ok := syscall.Getenv("LC_TIME"); ok && tag != "" {
		if loc, err := LookupLocale(tag); err == nil {
			return loc
		}
	}
	if tag, ok := syscall.Getenv("LANG"); ok && tag != "" {
		if loc, err := LookupLocale(tag); err == nil {
			return loc
		}
	}
	return builtinLocales["en-US"]
}

// LookupLocale returns a built-in locale for tag (BCP 47, e.g. "en-US", "de-DE").
func LookupLocale(tag string) (*Locale, error) {
	tag = normalizeLocaleTag(tag)
	if loc, ok := builtinLocales[tag]; ok {
		return loc, nil
	}
	// Try language-only fallback (e.g. "de" -> "de-DE").
	if i := stringslite.Index(tag, "-"); i > 0 {
		lang := tag[:i]
		for k, loc := range builtinLocales {
			if stringslite.HasPrefix(k, lang+"-") {
				return loc, nil
			}
		}
	}
	return nil, errUnknownLocale
}

func normalizeLocaleTag(tag string) string {
	tag = trimSpace(tag)
	if i := stringslite.Index(tag, "."); i >= 0 {
		tag = tag[:i]
	}
	var b []byte
	for i := 0; i < len(tag); i++ {
		if tag[i] == '_' {
			b = append(b, '-')
		} else {
			b = append(b, tag[i])
		}
	}
	return string(b)
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\t' || c == '\n' {
			s = s[:len(s)-1]
		} else {
			break
		}
	}
	return s
}

func (loc *Locale) monthNames(full bool) []string {
	if full {
		return loc.MonthNames[:]
	}
	return loc.MonthAbbr[:]
}

func (loc *Locale) weekdayNames(full bool) []string {
	if full {
		return loc.WeekdayNames[:]
	}
	return loc.WeekdayAbbr[:]
}

func lookupLocaleName(names []string, val string) (int, string, error) {
	// Prefer longest match (full names before abbreviations when both appear in one list).
	best := -1
	bestLen := 0
	for i, name := range names {
		if name == "" {
			continue
		}
		if len(val) >= len(name) && match(val[:len(name)], name) && len(name) > bestLen {
			best = i
			bestLen = len(name)
		}
	}
	if best < 0 {
		return -1, val, errBad
	}
	return best, val[bestLen:], nil
}
