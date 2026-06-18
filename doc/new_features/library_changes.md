# Library changes

Standard library improvements in this fork that are **not** language syntax — new APIs and behavior in existing packages.

Other extensions: [structured errors](errors.md), [LINQ](linq.md), [data structures](data_structures.md).

## `time`: .NET custom date/time format strings (`FormatCustom`)

### The problem

Go formats times with **reference-time layouts** (`"2006-01-02"`, `"Mon Jan 2 15:04:05 MST 2006"`). Developers from **.NET** expect [custom date and time format strings](https://learn.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings): a pattern built from **tokens** (`MM`, `dd`, `yyyy`, `g`, …) combined with literals:

```csharp
dateTime.ToString("MM/dd/yyyy g");   // e.g. 06/15/2009 A.D.
dateTime.ToString("MMMM dd, yyyy");
dateTime.ToString("HH:mm:ss");
```

There is no stdlib equivalent using .NET tokens.

### The fix

**`FormatCustom`** and **`ParseCustom`** implement [.NET custom date and time format strings](https://learn.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings) exactly. The format argument is always a **custom format string**: one or more **format specifiers** (single letters or **groups** of the same letter) plus literal text, separators, and escapes.

This is **not** Go’s reference-time layout (`2006`, `01`, `02`).  
This is **not** the separate [.NET standard format](https://learn.microsoft.com/en-us/dotnet/standard/base-types/standard-date-and-time-format-strings) one-character aliases (`"G"`, `"O"`, …) — use explicit custom patterns instead (e.g. `"yyyy-MM-ddTHH:mm:ss"` for sortable, `RFC3339` via Go’s **`Format`**, or a documented round-trip pattern).

### How a format string works

A pattern is scanned left to right. Each position matches the **longest** specifier from the table below, or a literal / escape:

```text
"MM/dd/yyyy g"
 │  │  │    │ └─ era token → "A.D." (en-US)
 │  │  │    └── literal space
 │  │  └─────── 4-digit year → "2009"
 │  └────────── date separator → "/" (locale)
 └───────────── 2-digit month → "06"
```

Result (en-US): **`06/15/2009 A.D.`**

Other examples:

```go
t := time.Date(2009, 6, 15, 13, 45, 30, 617000000, time.UTC)

t.FormatCustom("d")              // "15" — day token alone
t.FormatCustom("MM/dd/yyyy g")   // "06/15/2009 A.D."
t.FormatCustom("MMMM dd, yyyy")  // "June 15, 2009"
t.FormatCustom("HH:mm:ss.fff")   // "13:45:30.617"
t.FormatCustom("ddd, dd MMM yyyy HH':'mm:ss 'GMT'") // "Mon, 15 Jun 2009 13:45:30 GMT"
```

**Single custom specifier:** In .NET, a lone `"d"` can be ambiguous with the standard `"d"` alias. This fork’s **`FormatCustom`** always treats the string as a **custom** pattern; `"d"` is the **day** token. To match .NET’s `%` escape for “force custom”, **`"%d"`** is also accepted and equivalent to **`"d"`**.

### API

```go
// FormatCustom formats t with a .NET custom format string.
// Locale-sensitive tokens (MMMM, dddd, tt, /, :, …) use the default locale
// (LC_TIME / LANG, fallback "en-US").
func (t Time) FormatCustom(format string) string

// FormatCustomLocale is FormatCustom with an explicit BCP 47 tag.
func (t Time) FormatCustomLocale(format string, tag string) string

// ParseCustom parses value using the same custom format string.
// loc is the location when the pattern has no zone offset (like ParseInLocation).
func ParseCustom(format string, value string, loc *Location) (Time, error)

// ParseCustomLocale parses with an explicit locale tag.
func ParseCustomLocale(format string, value string, tag string, loc *Location) (Time, error)
```

Empty `format` is invalid. Parse/format mismatch returns an error (C# **`FormatException`** equivalent).

### Custom format specifiers

Specifiers are **case-sensitive**. Repeated letters form a group (`d` ≠ `dd` ≠ `ddd`). Implementation follows the [Microsoft specifier table](https://learn.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings).

Examples below use **`2009-06-15T13:45:30.6175425`** (Monday, UTC) unless noted. Default culture is **en-US**; locale-sensitive rows note other cultures where relevant.

| Spec | Description | Example output |
| ---- | ----------- | -------------- |
| `d` | Day of month, 1–31, no leading zero | `2009-06-01` → `1`; `2009-06-15` → `15` |
| `dd` | Day of month, 01–31 | `2009-06-01` → `01`; `2009-06-15` → `15` |
| `ddd` | Abbreviated weekday name (locale) | en-US → `Mon`; ru-RU → `Пн`; fr-FR → `lun.` |
| `dddd` | Full weekday name (locale) | en-US → `Monday`; ru-RU → `понедельник`; fr-FR → `lundi` |
| `f` | Tenths of a second | `.6170000` → `6`; `.05` → `0` |
| `ff` | Hundredths of a second | `.6170000` → `61`; `.0050000` → `00` |
| `fff` | Milliseconds | `.617` → `617`; `.0005` → `000` |
| `ffff` | Ten thousandths of a second | `.6175000` → `6175`; `.0000500` → `0000` |
| `fffff` | Hundred thousandths of a second | `.6175400` → `61754`; `.000005` → `00000` |
| `ffffff` | Millionths of a second | `.6175420` → `617542`; `.0000005` → `000000` |
| `fffffff` | Ten millionths of a second | `.6175425` → `6175425`; `.0001150` → `0001150` |
| `F` | Tenths of a second; omitted if zero | `.6170000` → `6`; `.0500000` → *(empty)* |
| `FF` | Hundredths; omitted if zero | `.6170000` → `61`; `.0050000` → *(empty)* |
| `FFF` | Milliseconds; omitted if zero | `.6170000` → `617`; `.0005000` → *(empty)* |
| `FFFF` | Ten thousandths; omitted if zero | `.5275000` → `5275`; `.0000500` → *(empty)* |
| `FFFFF` | Hundred thousandths; omitted if zero | `.6175400` → `61754`; `.0000050` → *(empty)* |
| `FFFFFF` | Millionths; omitted if zero | `.6175420` → `617542`; `.0000005` → *(empty)* |
| `FFFFFFF` | Ten millionths; omitted if zero | `.6175425` → `6175425`; `.0001150` → `000115` |
| `g`, `gg` | Period or era (locale) | `A.D.` |
| `h` | Hour, 12-hour clock, 1–12 | `01:45:30` → `1`; `13:45:30` → `1` |
| `hh` | Hour, 12-hour clock, 01–12 | `01:45:30` → `01`; `13:45:30` → `01` |
| `H` | Hour, 24-hour clock, 0–23 | `01:45:30` → `1`; `13:45:30` → `13` |
| `HH` | Hour, 24-hour clock, 00–23 | `01:45:30` → `01`; `13:45:30` → `13` |
| `K` | Time zone | UTC `Kind=UTC` → `Z`; offset `-07:00` → `-07:00` |
| `m` | Minute, 0–59 | `01:09:30` → `9`; `13:29:30` → `29` |
| `mm` | Minute, 00–59 | `01:09:30` → `09`; `13:45:30` → `45` |
| `M` | Month, 1–12 | `6` |
| `MM` | Month, 01–12 | `06` |
| `MMM` | Abbreviated month name (locale) | en-US → `Jun`; fr-FR → `juin` |
| `MMMM` | Full month name (locale) | en-US → `June`; da-DK → `juni` |
| `s` | Second, 0–59 | `13:45:09` → `9` |
| `ss` | Second, 00–59 | `13:45:09` → `09` |
| `t` | First character of AM/PM (locale) | en-US → `P`; ja-JP → `午` |
| `tt` | AM/PM designator (locale) | en-US → `PM`; ja-JP → `午後` |
| `y` | Year, 0–99 | `0001` → `1`; `2009` → `9`; `2019` → `19` |
| `yy` | Year, 00–99 | `0001` → `01`; `1900` → `00`; `2019` → `19` |
| `yyy` | Year, minimum 3 digits | `0001` → `001`; `0900` → `900`; `2009` → `2009` |
| `yyyy` | Year, four digits | `0001` → `0001`; `2009` → `2009` |
| `yyyyy` | Year, five digits | `0001` → `00001`; `2009` → `02009` |
| `z` | UTC offset hours, no leading zero | `-07:00` → `-7` |
| `zz` | UTC offset hours, leading zero | `-07:00` → `-07` |
| `zzz` | UTC offset hours and minutes | `-07:00` → `-07:00` |
| `:` | Time separator (locale) | en-US → `:`; it-IT → `.` |
| `/` | Date separator (locale) | en-US → `/`; de-DE → `.`; ar-DZ → `-` |
| `'…'`, `"…"` | Literal string | `"arr:" h:m t` → `arr: 1:45 P` |
| `%` | Next character is a custom specifier | `%h` → `1` (same as `h`) |
| `\` | Escape next character | `h \h` → `1 h` |
| other | Copied literally | `arr hh:mm t` → `arr 01:45 P` |

Composite example (tokens combined in one pattern):

```go
t := time.Date(2009, 6, 15, 13, 45, 30, 0, time.UTC)
t.FormatCustom("yyyy-MM-dd hh:mm:ss tt") // "2009-06-15 01:45:30 PM"
```

Locale affects **`ddd`**, **`dddd`**, **`MMM`**, **`MMMM`**, **`tt`**, **`g`**, **`/`**, **`:`**, and related tokens — not fixed numeric patterns like **`yyyy-MM-dd`**.

### Locale (`FormatCustomLocale`)

**`FormatCustomLocale`** only changes output when the pattern contains **locale-sensitive tokens**. It does not change the pattern itself.

```go
t.FormatCustom("yyyy-MM-dd")                      // "2009-06-15" — same in any locale
t.FormatCustom("MM/dd/yyyy g")                    // "06/15/2009 A.D." (en-US)
t.FormatCustomLocale("MM/dd/yyyy g", "de-DE")     // "15.06.2009 n. Chr." (era/month names/separators per culture)
t.FormatCustomLocale("dddd, MMMM dd, yyyy", "fr-FR") // "lundi, juin 15, 2009"
```

```go
type Locale struct { /* DateTimeFormatInfo-like */ }
func DefaultLocale() *Locale
func LookupLocale(tag string) (*Locale, error)
```

### Parse

**`ParseCustom`** mirrors C# **`DateTime.ParseExact`**: the input must match the pattern exactly (including literals, separators, and spacing).

```go
time.ParseCustom("MM/dd/yyyy", "06/15/2009", time.UTC)
time.ParseCustom("MM/dd/yyyy g", "06/15/2009 A.D.", time.UTC)
time.ParseCustom("HH:mm:ss", "13:45:30", time.UTC)
time.ParseCustomLocale("dd.MM.yyyy", "15.06.2009", "de-DE", time.UTC)
```

### Relation to Go `Format` / `Parse`

| API | Format string | Tokens |
| --- | ------------- | ------ |
| `t.Format("2006-01-02")` | Go reference time | `2006`, `01`, `02` |
| `t.Format(time.RFC3339)` | Go constant | Go layout |
| `t.FormatCustom("yyyy-MM-dd")` | .NET custom | `yyyy`, `MM`, `dd` |
| `t.FormatCustom("MM/dd/yyyy g")` | .NET custom composite | month/day/year + era |

Keep **`Format`** / **`Parse`** for existing Go code. Use **`FormatCustom`** / **`ParseCustom`** for .NET-style patterns.

### Errors

```go
var ErrBadFormat = errors.New("time: bad custom format string")
var ErrParseCustom = errors.New("time: FormatCustom parse failed")
```

### Reference

Full rules (character literals, `%` single-specifier rules, parser notes):  
https://learn.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings

## Quick reference

| Task | API |
| ---- | --- |
| Composite date + era | `t.FormatCustom("MM/dd/yyyy g")` |
| ISO-style date | `t.FormatCustom("yyyy-MM-dd")` |
| Day only | `t.FormatCustom("d")` |
| Localized long date | `t.FormatCustomLocale("dddd, MMMM dd, yyyy", tag)` |
| Parse exact pattern | `time.ParseCustom(format, value, loc)` |

## Structured errors in the standard library

Most public custom error types in the stdlib now embed **`errors.Error`** and capture stack traces at construction. See [Structured errors — Standard library integration](errors.md#standard-library-integration) for migrated packages, intentional exclusions, and how to inspect traces from `PathError`, `json.SyntaxError`, `tls.RecordHeaderError`, and similar types.

## Future library changes

Additional stdlib improvements will be listed here as they are specified or implemented.
