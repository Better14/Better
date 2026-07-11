
package time_test

import (
	"testing"
	"time"
)

var refTime = time.Date(2009, 6, 15, 13, 45, 30, 617000000, time.UTC)

func TestFormatCustomBasic(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"d", "15"},
		{"dd", "15"},
		{"MM/dd/yyyy g", "06/15/2009 A.D."},
		{"MMMM dd, yyyy", "June 15, 2009"},
		{"HH:mm:ss.fff", "13:45:30.617"},
		{"yyyy-MM-dd hh:mm:ss tt", "2009-06-15 01:45:30 PM"},
		{"ddd, dd MMM yyyy HH':'mm:ss 'GMT'", "Mon, 15 Jun 2009 13:45:30 GMT"},
		{"yyyy-MM-dd", "2009-06-15"},
		{"%d", "15"},
		{"h", "1"},
		{"hh", "01"},
		{"H", "13"},
		{"HH", "13"},
		{"t", "P"},
		{"tt", "PM"},
		{"yyyy", "2009"},
		{"yy", "09"},
		{"K", "Z"},
	}
	for _, tt := range tests {
		got := refTime.FormatCustom(tt.format)
		if got != tt.want {
			t.Errorf("FormatCustom(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestFormatCustomLocale(t *testing.T) {
	tests := []struct {
		format string
		tag    string
		want   string
	}{
		{"MM/dd/yyyy g", "de-DE", "15.06.2009 n. Chr."},
		{"dddd, MMMM dd, yyyy", "fr-FR", "lundi, juin 15, 2009"},
		{"yyyy-MM-dd", "de-DE", "2009-06-15"},
		{":", "it-IT", "."},
	}
	for _, tt := range tests {
		got := refTime.FormatCustomLocale(tt.format, tt.tag)
		if got != tt.want {
			t.Errorf("FormatCustomLocale(%q, %q) = %q, want %q", tt.format, tt.tag, got, tt.want)
		}
	}
}

func TestParseCustom(t *testing.T) {
	tests := []struct {
		format string
		value  string
		want   time.Time
	}{
		{"MM/dd/yyyy", "06/15/2009", time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"MM/dd/yyyy g", "06/15/2009 A.D.", time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC)},
		{"HH:mm:ss", "13:45:30", time.Time{}}, // time-only; date from Now()
		{"yyyy-MM-dd hh:mm:ss tt", "2009-06-15 01:45:30 PM", time.Date(2009, 6, 15, 13, 45, 30, 0, time.UTC)},
	}
	for _, tt := range tests {
		got, err := time.ParseCustom(tt.format, tt.value, time.UTC)
		if err != nil {
			t.Errorf("ParseCustom(%q, %q): %v", tt.format, tt.value, err)
			continue
		}
		if tt.format == "HH:mm:ss" {
			if got.Hour() != 13 || got.Minute() != 45 || got.Second() != 30 {
				t.Errorf("ParseCustom(%q, %q) = %v, want 13:45:30", tt.format, tt.value, got)
			}
			continue
		}
		if !got.Equal(tt.want) {
			t.Errorf("ParseCustom(%q, %q) = %v, want %v", tt.format, tt.value, got, tt.want)
		}
	}
}

func TestParseCustomLocale(t *testing.T) {
	got, err := time.ParseCustomLocale("dd.MM.yyyy", "15.06.2009", "de-DE", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFormatCustomErrors(t *testing.T) {
	if refTime.FormatCustom("") != "" {
		t.Error("expected empty result for empty format")
	}
	_, err := time.ParseCustom("", "x", time.UTC)
	if err != time.ErrBadFormat {
		t.Errorf("ParseCustom empty format: got %v, want ErrBadFormat", err)
	}
	_, err = time.ParseCustom("MM/dd/yyyy", "bad", time.UTC)
	if err != time.ErrParseCustom {
		t.Errorf("ParseCustom bad input: got %v, want ErrParseCustom", err)
	}
}

func TestLookupLocale(t *testing.T) {
	loc, err := time.LookupLocale("en-US")
	if err != nil || loc.Tag != "en-US" {
		t.Fatalf("LookupLocale(en-US): %v, %v", loc, err)
	}
	_, err = time.LookupLocale("xx-YY")
	if err == nil {
		t.Error("expected error for unknown locale")
	}
}

func TestFormatCustomRoundTrip(t *testing.T) {
	format := "yyyy-MM-dd HH:mm:ss"
	tm := time.Date(2009, 6, 15, 13, 45, 30, 0, time.UTC)
	s := tm.FormatCustom(format)
	got, err := time.ParseCustom(format, s, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(tm) {
		t.Errorf("round trip: got %v, want %v", got, tm)
	}
}
