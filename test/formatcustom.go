// run

package main

import "time"

func main() {
	t := time.Date(2009, 6, 15, 13, 45, 30, 617000000, time.UTC)

	check := func(format, want string) {
		if got := t.FormatCustom(format); got != want {
			panic("FormatCustom(" + format + ") = " + got + ", want " + want)
		}
	}

	check("d", "15")
	check("MM/dd/yyyy g", "06/15/2009 A.D.")
	check("MMMM dd, yyyy", "June 15, 2009")
	check("HH:mm:ss.fff", "13:45:30.617")
	check("yyyy-MM-dd hh:mm:ss tt", "2009-06-15 01:45:30 PM")
	check("ddd, dd MMM yyyy HH':'mm:ss 'GMT'", "Mon, 15 Jun 2009 13:45:30 GMT")

	if got := t.FormatCustomLocale("MM/dd/yyyy g", "de-DE"); got != "15.06.2009 n. Chr." {
		panic("de-DE: " + got)
	}
	if got := t.FormatCustomLocale("dddd, MMMM dd, yyyy", "fr-FR"); got != "lundi, juin 15, 2009" {
		panic("fr-FR: " + got)
	}

	got, err := time.ParseCustom("MM/dd/yyyy", "06/15/2009", time.UTC)
	if err != nil || !got.Equal(time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC)) {
		panic("ParseCustom MM/dd/yyyy")
	}

	got, err = time.ParseCustom("yyyy-MM-dd hh:mm:ss tt", "2009-06-15 01:45:30 PM", time.UTC)
	if err != nil || !got.Equal(time.Date(2009, 6, 15, 13, 45, 30, 0, time.UTC)) {
		panic("ParseCustom yyyy-MM-dd hh:mm:ss tt")
	}

	got, err = time.ParseCustomLocale("dd.MM.yyyy", "15.06.2009", "de-DE", time.UTC)
	if err != nil || !got.Equal(time.Date(2009, 6, 15, 0, 0, 0, 0, time.UTC)) {
		panic("ParseCustomLocale dd.MM.yyyy")
	}
}
