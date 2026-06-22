# Structured errors (`errors.Base`)

Go’s standard [`errors`](https://pkg.go.dev/errors) package gains a public **`Base`** struct that carries a message, an optional **stack trace**, and an optional **inner error** for wrapping. The type implements the standard `error` interface and participates in Go 1.13+ error wrapping (`Unwrap`, `errors.Is`, `errors.As`, `fmt.Errorf` with `%w`).

`type Error = Base` is a type alias so APIs such as **`New`** can return `*Error` without clashing with the **`Error()`** method name on `*Base`.

## Overview

Today, `errors.New("…")` and `errors.New("…", args…)` return a `*Error` (same as `*Base`) with message and stack trace (assignable to `error` as before). When arguments are provided, the format string is interpreted like `fmt.Sprintf`. Diagnosing failures in production often required ad hoc logging or third-party packages to attach stack traces and error chains.

The extended **`errors.Base`** struct makes that structure first-class:

| Field | Type | Role |
| ----- | ---- | ---- |
| `Message` | `string` | Human-readable description for this layer |
| `StackTrace` | `StackTrace` | Call stack captured when this layer was created (innermost frame first) |
| `InnerError` | `*Base` | Previous error in the chain (`nil` for a root error) |

`InnerError` is typed as `*Base` for the common case in this package. Wrapping still interoperates with arbitrary `error` values via `Unwrap` and `%w` (see [Interoperability](#interoperability)).

## Types

```go
package errors

// Base is the embeddable structured-error payload (message, stack trace, inner link).
// The name Base avoids a field/method name clash with Error().
type Base struct {
	Message    string
	StackTrace StackTrace
	InnerError *Base
}

// Error is an alias for Base. APIs such as New and Wrap return *Error.
type Error = Base

// StackTrace is a captured call stack (innermost frame first).
type StackTrace []StackFrame

// StackFrame identifies one stack frame.
type StackFrame struct {
	Function string // qualified name, e.g. "main.readFile"
	File     string
	Line     int
}
```

## API

### `Error()` — `error` interface

```go
func (e *Base) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}
```

`Error()` returns **`Message` only**. It does not walk the chain or include stack traces. Use **`String()`** (or `fmt` formatting below) for full serialization.

### `String()` — `fmt.Stringer`

`*Base` implements **`fmt.Stringer`**. A single `String()` call serializes the **full error chain and stack traces** for logging, persistence, or ad hoc conversion:

```go
func (e *Base) String() string {
	if e == nil {
		return "<nil>"
	}
	var b strings.Builder
	if e.InnerError != nil {
		b.WriteString(e.InnerError.String())
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
	}
	b.WriteString(e.Message)
	if len(e.StackTrace) > 0 {
		b.WriteByte('\n')
		b.WriteString(e.StackTrace.String())
	}
	return b.String()
}
```

Order for each layer (inner layers first via recursion):

1. **`InnerError.String()`** — if `InnerError != nil`, recurse into the inner chain.
2. **`Message`** — this layer’s message.
3. **`StackTrace.String()`** — this layer’s stack frames as text.

Example output for a wrapped error:

```text
permission denied
open file: /etc/app.conf
main.loadConfig (/app/config.go:42)
main.main (/app/main.go:10)
main.readFile (/app/io.go:18)
main.main (/app/main.go:8)
```

`Error()` remains the short form for the `error` interface; **`String()`** is the canonical full dump. Explicit conversion:

```go
log.Println(err.Error()) // message only
log.Println(err.String()) // full chain + traces (when static type is *Base)
s := (*Base)(nil).String() // "<nil>"
```

### `StackTrace.String()`

`StackTrace` implements **`fmt.Stringer`**. It formats **every frame** in the slice, one per line:

```go
func (st StackTrace) String() string {
	if len(st) == 0 {
		return ""
	}
	var b strings.Builder
	for i, f := range st {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s (%s:%d)", f.Function, f.File, f.Line)
	}
	return b.String()
}
```

`StackTrace` is a named slice type so `String()` applies to the field `e.StackTrace` without extra wrapping.

### `New` — root error with stack trace

```go
func New(message string) *Error {
	return newError(message)
}
```

Each call to `New` captures a **fresh stack trace** at the call site. The return type is **`*Error`** (alias for `*Base`, not the `error` interface). Because `*Base` implements `error`, existing APIs and call sites that use `error` remain valid:

```go
var err error = errors.New("permission denied") // OK — *Error / *Base is an error
return errors.New("permission denied")          // OK — return error from func
```

Example:

```go
return errors.New("permission denied")
```

### Choosing how to create errors

| Situation | API | Returns |
| --------- | --- | ------- |
| Generic structured error | **`errors.New(format, args…)`** | `*Error` (`*Base`) |
| Named type, embeds `errors.Base` only (no extra fields) | `errors.NewCustom[T]` `(format, args…)` | `*T` |
| Named type with extra fields (e.g. `Code int`) | **`InitCustom(&err.Base, format, args…)`** or **`NewMyError(...)`** | `*MyError` |

#### `errors.New` — generic structured error

Use for root errors and sentinels when **`errors.Base`** is enough. The first argument is a format string; optional arguments are formatted like `fmt.Sprintf`:

```go
var ErrNotFound = errors.New("not found")

func openFile(path string) error {
	return errors.New("open file: %s", path)
}
```

With no extra arguments, the format string is used as-is (so literals containing `%` are unchanged).

#### `errors.NewCustom` — custom errors that embed `errors.Base`

```go
func NewCustom[T ~struct{ Base }](format string, args ...any) *T
func InitCustom(e *Base, format string, args ...any)
```

`NewCustom` returns a new value of type `*T`. `InitCustom` fills an embedded `Base` field in place (for types with extra domain fields).

**When to use which**

| API | Use when | Example |
| --- | -------- | ------- |
| `NewCustom[T]` `(format, args…)` | Named type embeds **only** `errors.Base` (no extra fields) | `return errors.NewCustom[AppError]("invalid id: %s", id)` |
| **`InitCustom(&err.Base, format, args…)`** | Named type has **extra domain fields**; initialize the embedded layer in place | `InitCustom(&myErr.Base, "code %d", code); myErr.Code = code` |

**`NewCustom[T]` `(format, args…)`** — embed only, no extra fields

```go
type AppError struct {
	errors.Base
}

func validate(id string) *AppError {
	return errors.NewCustom[AppError]("invalid id: %s", id)
}
```

**`InitCustom(&err.Base, format, args…)` — extra fields on the same struct**

```go
type MyError struct {
	errors.Base
	Code int
}

func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.InitCustom(&err.Base, "%s", message)
	err.Code = code
	return &err
}
```

Pass **`&err.Base`** (address of the embedded field), not `&err`.

#### `NewMyError` — alternative constructor with extra fields

Equivalent to `InitCustom(&err.Base, format, args…)` plus assigning domain fields; you can also copy from **`errors.New`**:

```go
func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.InitCustom(&err.Base, "%s", message)
	err.Code = code
	return &err
}
```

Or copy from `*errors.New(format, args…)` in a composite literal:

```go
func NewMyErrorAlt(code int, message string) *MyError {
	return &MyError{
		Base: *errors.New("%s", message),
		Code: code,
	}
}
```

Both set **`Message`** and **`StackTrace`** on the embedded struct. Prefer **`InitCustom(&err.Base, format, args…)`** when building the value field-by-field.

### `Wrap` — add context and a new stack frame layer

```go
func (e *Base) Wrap(message string) *Base {
	if e == nil {
		return newError(message)
	}
	return &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: e,
	}
}
```

`Wrap` prepends a new layer: the outer `Message` describes this hop; `InnerError` points at the previous `*Base`. The outer layer’s `StackTrace` is captured at the **wrap call site**, not at the original failure site (each layer keeps its own trace).

Example:

```go
if err := readConfig(path); err != nil {
	return err.Wrap("load config")
}
```

Package-level convenience (same behavior, accepts any `error`):

```go
func Wrap(err error, message string) *Base {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Base); ok {
		return e.Wrap(message)
	}
	return &Base{
		Message:    message,
		StackTrace: CaptureStackTrace(),
		InnerError: bridgeError(err), // package-internal; preserves chain for Is/As
	}
}
```

When `err` is already `*Base`, `Wrap` delegates to the method receiver. Otherwise the package converts `err` internally (callers never write `bridgeError` themselves).

### `Unwrap` — error-chain inspection

```go
func (e *Base) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.InnerError != nil {
		return e.InnerError
	}
	// ... bridged non-*Base operand when InnerError is nil
}
```

`Unwrap` returns the next error in the chain, so `errors.Is`, `errors.As`, and `%w` chains work unchanged.

### Stack trace capture

```go
// CaptureStackTrace captures the current goroutine stack, skipping
// internal errors, fmt, and log frames. Used by New, Wrap, and single-%w fmt.Errorf.
// The slice is ordered innermost caller first.
func CaptureStackTrace() StackTrace {
	// Walk runtime.Callers / runtime.CallersFrames and populate the slice.
	...
}
```

Stack traces are captured when a `*Base` is created through **`New`**, **`Wrap`**, **`NewCustom`**, **`InitCustom`**, or **`fmt.Errorf` with zero or one `%w` verb** (see [Interoperability](#fmterrorf-and-w)). Other `error` implementations returned without conversion have no trace unless they embed or wrap `*Base`.

**No stack trace at the call site** for:

| Call | Why |
| ---- | --- |
| `fmt.Errorf("… %w … %w", e1, e2)` (two or more `%w`) | Returns upstream `fmt.wrapErrors`, not `*errors.Base` — formatted message and `Unwrap() []error` only |
| `errors.Join(errs…)` | Aggregates existing errors; does not add a new layer |
| Plain custom `error` types | Unless they embed `errors.Base` or are wrapped by `errors.Wrap` / single-`%w` `fmt.Errorf` |

For multi-`%w` `fmt.Errorf`, stack traces on **`e1`** and **`e2`** themselves are unchanged; only the outer `Errorf` hop does not capture a new trace.

## Usage patterns

### Root error

```go
func openFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("open file: %s", path)
	}
	defer f.Close()
	// ...
}
```

### Wrap on return path

```go
func loadUser(id string) (*User, error) {
	u, err := fetchUser(id)
	if err != nil {
		return nil, err.Wrap("fetch user " + id)
	}
	return u, nil
}
```

### `errors.Is` / `errors.As`

```go
var ErrNotFound = errors.New("not found")

if errors.Is(err, ErrNotFound) {
	// works through Wrap / InnerError chain
}

var pathErr *Base
if errors.As(err, &pathErr) {
	fmt.Println(pathErr.Message)
	fmt.Println(pathErr.StackTrace)
}
```

## Interoperability

### `fmt.Errorf` and `%w`

**Today (upstream Go):** `fmt.Errorf` does **not** capture a stack trace. It returns either `errors.New(msg)` (no `%w`) or an opaque `fmt.wrapError` / `fmt.wrapErrors` (with `%w`).

**In this fork:** `fmt.Errorf` is updated to produce **`*errors.Base` when possible**, with a stack trace at the **`Errorf` call site**:

| Form | Result |
| ---- | ------ |
| `fmt.Errorf("msg")` | `*errors.Base` — same as `errors.New("msg")` (already delegated to `errors.New` today) |
| `fmt.Errorf("… %w", err)` (one `%w`) | `*errors.Base` — replaces `fmt.wrapError`; `Message` is the **full formatted string**; `InnerError` links the operand for `Is` / `As` |
| `fmt.Errorf("… %w … %w", e1, e2)` (multiple `%w`) | **`fmt.wrapErrors` unchanged** — multiple operands do not map to a single `InnerError`; **no stack trace** at the `Errorf` call site; no `*errors.Base` outer layer until a structured multi-cause type exists |

Example (single `%w`):

```go
if err := readFile(path); err != nil {
	return fmt.Errorf("readFile error: %w", err)
}
```

The returned value is `*errors.Base` with a fresh `StackTrace` at this line. `errors.Is` / `errors.As` traverse `InnerError` as with `Wrap`. The outer `Error()` string is still the full formatted message (including the wrapped error’s text from `%w`), matching today’s `fmt.wrapError` behavior.

Example (multiple `%w` — **no stack trace**):

```go
err := fmt.Errorf("failed: %w and %w", openErr, readErr)
// err is fmt.wrapErrors, not *errors.Base
// err.Error() → "failed: <openErr> and <readErr>"
// Unwrap() []error → [openErr, readErr]
// No StackTrace captured at this Errorf call site.
```

Multiple `%w` operands still produce a formatted message and `Unwrap() []error` for `errors.Is` / `errors.As`, but the implementation stays the upstream `fmt.wrapErrors` type. Stack traces from earlier `errors.New` / `Wrap` / single-`%w` layers on `openErr` and `readErr` are preserved on those inner errors; only the outer `Errorf` hop does not add a new trace. Prefer **`errors.Join`** when a newline-separated message is enough and you do not need a custom format string.

Implementation sketch (`fmt` package):

```go
case 1:
	wrapped := a[p.wrappedErrs[0]].(error)
	err = errors.NewWrapped(s, wrapped) // returns *Base as error
```

(`NewWrapped` is the internal helper; callers use `fmt.Errorf` with `%w`.)

When you do **not** need `fmt` verbs in the message, **`errors.Wrap`** remains clearer:

```go
return errors.Wrap(err, "readFile error")
```

Both single-`%w` `fmt.Errorf` and `errors.Wrap` capture a stack trace at the call site. `Wrap` sets `Message` to the context string only; `fmt.Errorf` sets `Message` to the entire formatted output.

### Plain `error` values

Functions may still return `error`. Callers that receive a non-`*Base` value use `errors.Is` / `errors.As` as today. Converting at the boundary:

```go
if err != nil {
	return errors.New(err.Error()) // loses typed inner link unless err is already *Base
}
```

Prefer propagating `*Base` (or `*Error`) from APIs that own error construction, and use `%w` or explicit `InnerError` assignment when bridging packages.

### Relation to `T!` result types

[`T!`](result_types.md) lowers to `(T, error)`. An `int!` or function returning `error` may hold an `*errors.Base` like any other `error` implementation:

```go
func readCount() int! {
	n, err := parse()
	if err != nil {
		return err.Wrap("parse count") // *Base assignable to error slot
	}
	return n
}
```

## Formatting and display

`*Base` also implements **`fmt.Formatter`** for fine-grained control; default verbs delegate to `Error()` or `String()` as appropriate:

```go
func (e *Base) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.String()) // %+v → full String() serialization
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error()) // %v, %s → Message only
	default:
		fmt.Fprintf(s, "%%!%c(*errors.Base)", verb)
	}
}
```

| Verb | Output |
| ---- | ------ |
| `%s`, `%v` | `Message` only (same as `Error()`) |
| `%+v` | Full `String()` — inner chain, messages, stack traces |
| `.String()` | Full serialization (same as `%+v`) |

Example `%+v` / `String()` output:

```text
permission denied
open file: /etc/app.conf
main.readFile (/app/io.go:18)
main.main (/app/main.go:8)
load config
main.loadConfig (/app/config.go:42)
main.main (/app/main.go:10)
```

(Innermost layer first in the chain; each layer’s stack follows its message.)

## Migration and upgrade

This change is **opt-in at the call site**. The `error` interface, function signatures, `errors.Is`, and `errors.As` are unchanged — existing code compiles without edits. Upgrade paths below add stack traces and structured wrapping where you want them.

### `errors.New` — root errors

Prefer **`errors.New`** for new root errors and sentinels. Every call captures a stack trace at the creation site:

```go
var ErrNotFound = errors.New("not found")

func validate(id string) error {
	if id == "" {
		return errors.New("empty id")
	}
	return nil
}

func openFile(path string) error {
	return errors.New("open file: %s", path)
}
```

Existing `errors.New("…")` call sites continue to work. Use format arguments instead of string concatenation for dynamic messages.

### `fmt.Errorf` — stack traces without rewrites

Existing **`fmt.Errorf`** call sites also gain stack traces where the implementation can return `*errors.Base`:

- `fmt.Errorf("…")` — via `errors.New`, same as above.
- `fmt.Errorf("… %w", err)` — one `%w` operand returns `*errors.Base` at the `Errorf` site (replaces `fmt.wrapError`).
- `fmt.Errorf("… %w … %w", e1, e2)` — **unchanged** `fmt.wrapErrors`; **no stack trace** at the `Errorf` site (see [Interoperability](#fmterrorf-and-w)).

No migration required for typical wrap-and-return code. Use **`errors.Wrap`** when you want a short context label without `fmt` verbs; use **`fmt.Errorf`** when the message needs formatting (`%s`, `%d`, etc.).

```go
// Both capture a stack trace at this hop (single inner error):
return fmt.Errorf("load config %s: %w", path, err)
return errors.Wrap(err, "load config "+path)
```

### Prefer `errors.Wrap` for simple context

When the message is a fixed prefix plus a wrapped error, **`errors.Wrap`** is still recommended for clarity (no format string). **`fmt.Errorf` with `%w`** is equivalent for stack traces when you need formatting.

### Embed `errors.Base` in custom error types

Custom error structs should **embed `errors.Base`** so stack traces and wrapping come from the standard type while you add domain fields:

```go
// Before
type MyError struct {
	Msg  string
	Code int
}

func (e *MyError) Error() string { return e.Msg }

// After — recommended
type MyError struct {
	errors.Base
	Code int
}

func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.InitCustom(&err.Base, "%s", message)
	err.Code = code
	return &err
}
```

`*MyError` implements `error` via the promoted `Error()` method.

See [Choosing how to create errors](#choosing-how-to-create-errors) for `errors.New`, `NewCustom`, `InitCustom`, and `NewMyError`.

Override `Error()` or `Unwrap()` on `*MyError` only when the default embedded behavior is not enough; otherwise rely on promotion.

For wrapping a custom error on the return path:

```go
if err != nil {
	return errors.Wrap(err, "operation failed")
}
```

If `err` is `*MyError`, `Wrap` adds an outer `*errors.Base` layer whose `InnerError` links back into the chain.

### Optional: `go fix`

Mechanical rewrites (e.g. `fmt.Errorf("… %w", err)` → `errors.Wrap(err, "…")`) may be provided as **`go fix`** analyzers in a later release. Manual migration following the patterns above is always sufficient.

## Standard library integration

Most **public custom error types** in the standard library now **embed `errors.Base`** and capture a **stack trace at construction** via package-local constructors (`NewPathError`, `newSyntaxError`, `errors.InitCustom`, and similar). This applies to errors returned from normal API use — not only to direct `errors.New` / `fmt.Errorf` call sites in application code.

### What was migrated

Custom error structs were updated package-by-package. Each type keeps its existing **`Error()`**, **`Unwrap()`**, **`Is()`**, and **`As()`** behavior; only construction sites were routed through helpers that call **`errors.InitCustom`** (or **`errors.New`** for sentinels that already used it).

| Area | Packages / types (representative) |
| ---- | --------------------------------- |
| Core I/O | `io/fs.PathError`; `os.SyscallError`, `LinkError`; `internal/poll.DeadlineExceededError` |
| Networking | `net.OpError`, `ParseError`, `AddrError`, `DNSError`; `net/netip` parse errors |
| Parsing & encoding | `strconv.NumError`; `encoding/json` (v1, v2, `jsontext`); `encoding/xml`, `encoding/csv`; `encoding/asn1` (`StructuralError`, `SyntaxError`, `invalidUnmarshalError`); `encoding/gob` (`gobError`) |
| HTTP | `net/http` request/response errors; `net/http/internal/http2` frame and stream errors |
| Time & templates | `time.ParseError`, `time.LoadLocationError`; `text/template` execution errors |
| Process & reflection | `os/exec` errors; `reflect` / `internal/reflectlite` value errors |
| Crypto & debug | `crypto/x509` verification errors; `crypto/tls` alert and handshake errors; `debug/elf`, `debug/macho`, `debug/plan9obj`, `debug/gosym` (`DecodingError`, `UnknownLineError`) |
| Archive | `archive/tar` (`headerError`) |
| Syscall (public) | `syscall.DLLError`; `syscall/js` errors |

Sentinels and helpers that already used **`errors.New`** or single-`%w` **`fmt.Errorf`** (for example many `io.EOF`-style values and `internal/oserror` sentinels) already carried stack traces from those APIs and were left as-is unless a named struct wrapper was added.

### What was not migrated (by design)

| Category | Reason |
| -------- | ------ |
| **`runtime/`** | Cannot import `errors`; panics already include runtime stacks |
| **`vendor/`**, most **`internal/*`** | Not public API; low value vs churn (exceptions: `internal/poll`, `internal/reflectlite`) |
| **`cmd/`** | Tooling, not library surface |
| Package-init **sentinels** (`io.EOF`, etc.) | Stack captured at `init`, not at the failing call site |
| **Multi-`%w` `fmt.Errorf`** | Still returns `fmt.wrapErrors` — no outer stack (see [Interoperability](#fmterrorf-and-w)) |
| **Named scalar errors** (`net/url.EscapeError`, `CorruptInputError`, …) | Type shape unchanged; stacks add little for int/string error types |
| **Deprecated / unused types** | e.g. `compress/flate.ReadError` / `WriteError` |
| **`reflect` panics** | Runtime panic stack is usually sufficient |

### Performance

Stack capture runs **only when an error value is constructed**. Success paths are unchanged. Packages that can return errors frequently on failure but not on success (JSON decode, TLS handshake, ASN.1 parse, tar header validation, etc.) pay the capture cost **once per returned error**, which is acceptable for diagnostic value.

### Inspecting traces from stdlib errors

Use **`errors.As`** to reach the embedded layer, then read **`StackTrace`** or format with **`%+v`** / **`String()`**:

```go
var pe *fs.PathError
if errors.As(err, &pe) {
	log.Printf("%+v", &pe.Base) // message + stack at PathError construction site
}
```

For types that embed `errors.Base` by value, take the address of the embedded field (as in the example) or use **`fmt` `%+v`** on a pointer to the outer type when it implements **`fmt.Formatter`** through promotion.

## Design notes

- **One stack trace per layer** — wrapping adds a new `[]StackFrame` at the wrap site; inner errors retain their original traces.
- **Multi-`%w` `fmt.Errorf` is unchanged** — `fmt.Errorf("… %w … %w", e1, e2)` does not capture a stack trace at that call site and does not return `*errors.Base`; use single-`%w` chaining, `errors.Wrap`, or `errors.Join` when you need structured traces instead.
- **`Message` is the layer label** — use short, stable text; put dynamic detail in `Message` or in wrapped inner messages consistently.
- **Nil-safe** — `(*Base)(nil).Error()` and `(*Base)(nil).String()` return `"<nil>"`; `Wrap` on `nil` behaves like `New`.
- **Not valid in upstream Go** — this extension applies to this fork’s standard library.

## Quick reference

| API | Purpose |
| --- | ------- |
| `errors.New(format, args…)` | Root `*Error` / `*Base` with stack trace |
| `errors.NewCustom[T]` `(format, args…)` | Root custom `*T` embedding `errors.Base` |
| `errors.InitCustom(&e, format, args…)` | Fill embedded `*errors.Base` in place (extra domain fields) |
| `NewMyError(...)` | Custom `*MyError` with extra fields |
| `err.Wrap(msg)` | New outer layer; `InnerError = err` |
| `err.Error()` | Returns `Message` only |
| `err.String()` | Full serialization: inner chain + `Message` + `StackTrace` |
| `err.StackTrace.String()` | All frames at this layer, one per line |
| `err.Unwrap()` | Returns `InnerError` or bridged `error` |
| `err.StackTrace` | `StackTrace` (`[]StackFrame`) at this layer |
| `fmt.Errorf("…")` | `*errors.Base` with stack trace (via `errors.New`) |
| `fmt.Errorf("… %w", err)` | `*errors.Base` with stack trace (one `%w`) |
| `fmt.Errorf("… %w … %w", e1, e2)` | `fmt.wrapErrors` — **no stack trace** at `Errorf` site |
| `errors.Join(errs…)` | Combined error — **no stack trace** (aggregates operands) |
