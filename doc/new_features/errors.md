# Structured errors (`errors.Error`)

Go’s standard [`errors`](https://pkg.go.dev/errors) package gains a public **`Error`** type that carries a message, an optional **stack trace**, and an optional **inner error** for wrapping. The type implements the standard `error` interface and participates in Go 1.13+ error wrapping (`Unwrap`, `errors.Is`, `errors.As`, `fmt.Errorf` with `%w`).

## Overview

Today, `errors.New("…")` returns a `*Error` with message and stack trace (assignable to `error` as before). Diagnosing failures in production often required ad hoc logging or third-party packages to attach stack traces and error chains.

The extended **`errors.Error`** struct makes that structure first-class:

| Field | Type | Role |
| ----- | ---- | ---- |
| `Message` | `string` | Human-readable description for this layer |
| `StackTrace` | `StackTrace` | Call stack captured when this layer was created (innermost frame first) |
| `InnerError` | `*Error` | Previous error in the chain (`nil` for a root error) |

`InnerError` is typed as `*Error` for the common case in this package. Wrapping still interoperates with arbitrary `error` values via `Unwrap` and `%w` (see [Interoperability](#interoperability)).

## Types

```go
package errors

// Error is a structured error with message, stack trace, and optional inner error.
type Error struct {
	Message    string
	StackTrace StackTrace
	InnerError *Error
}

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
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}
```

`Error()` returns **`Message` only**. It does not walk the chain or include stack traces. Use **`String()`** (or `fmt` formatting below) for full serialization.

### `String()` — `fmt.Stringer`

`*Error` implements **`fmt.Stringer`**. A single `String()` call serializes the **full error chain and stack traces** for logging, persistence, or ad hoc conversion:

```go
func (e *Error) String() string {
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
log.Println(err.String()) // full chain + traces (when static type is *Error)
s := (*Error)(nil).String() // "<nil>"
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
	return &Error{
		Message:    message,
		StackTrace: createStackTrace(),
		InnerError: nil,
	}
}
```

Each call to `New` captures a **fresh stack trace** at the call site. The return type is **`*Error`** (not the `error` interface). Because `*Error` implements `error`, existing APIs and call sites that use `error` remain valid:

```go
var err error = errors.New("permission denied") // OK — *Error is an error
return errors.New("permission denied")          // OK — return error from func
```

Example:

```go
return errors.New("permission denied")
```

### Choosing how to create errors

| Situation | API | Returns |
| --------- | --- | ------- |
| Generic structured error | **`errors.New(msg)`** | `*Error` |
| Named type, embeds `errors.Error` only (no extra fields) | **`errors.NewCustom[T](msg)`** | `*T` |
| Named type with extra fields (e.g. `Code int`) | **`NewCustom(&err.Error, msg)`** or **`NewMyError(...)`** | `*MyError` |

#### `errors.New` — generic structured error

Use for root errors and sentinels when **`errors.Error`** is enough:

```go
var ErrNotFound = errors.New("not found")

func openFile(path string) error {
	return errors.New("open file: " + path)
}
```

#### `errors.NewCustom` — custom errors that embed `errors.Error`

Two overloads share the same name; the compiler picks by argument types ([overloading](overloading.md)):

```go
func NewCustom[T ~struct{ Error }](message string) *T
func NewCustom(e *Error, message string)
```

Both set **`Message`**, **`StackTrace`**, and **`InnerError`** (`nil`) on the embedded layer. Use **`errors.Wrap`** to add outer layers.

**When to use which**

| Overload | Use when | Example |
| -------- | -------- | ------- |
| **`NewCustom[T](msg)`** | Named type embeds **only** `errors.Error` (no extra fields) | `return errors.NewCustom[AppError](msg)` |
| **`NewCustom(&err.Error, msg)`** | Named type has **extra domain fields**; initialize the embedded layer in place | `errors.NewCustom(&myErr.Error, msg); myErr.Code = 404` |

**`NewCustom[T](msg)` — embed only, no extra fields**

```go
type AppError struct {
	errors.Error
}

func validate(id string) *AppError {
	return errors.NewCustom[AppError]("invalid id: " + id)
}
```

**`NewCustom(&err.Error, msg)` — extra fields on the same struct**

```go
type MyError struct {
	errors.Error
	Code int
}

func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.NewCustom(&err.Error, message)
	err.Code = code
	return &err
}
```

Pass **`&err.Error`** (address of the embedded field), not `&err`.

#### `NewMyError` — alternative constructor with extra fields

Equivalent to `NewCustom(&err.Error, msg)` plus assigning domain fields; you can also copy from **`errors.New`**:

```go
func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.NewCustom(&err.Error, message)
	err.Code = code
	return &err
}
```

Or copy from `*errors.New(message)` in a composite literal:

```go
func NewMyErrorAlt(code int, message string) *MyError {
	return &MyError{
		Error: *errors.New(message),
		Code:  code,
	}
}
```

Both set **`Message`** and **`StackTrace`** on the embedded struct. Prefer **`NewCustom(&err.Error, msg)`** when building the value field-by-field.

### `Wrap` — add context and a new stack frame layer

```go
func (e *Error) Wrap(message string) *Error {
	if e == nil {
		return New(message)
	}
	return &Error{
		Message:    message,
		StackTrace: createStackTrace(),
		InnerError: e,
	}
}
```

`Wrap` prepends a new layer: the outer `Message` describes this hop; `InnerError` points at the previous `*Error`. The outer layer’s `StackTrace` is captured at the **wrap call site**, not at the original failure site (each layer keeps its own trace).

Example:

```go
if err := readConfig(path); err != nil {
	return err.Wrap("load config")
}
```

Package-level convenience (same behavior, accepts any `error`):

```go
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e.Wrap(message)
	}
	return &Error{
		Message:    message,
		StackTrace: createStackTrace(),
		InnerError: fromError(err), // package-internal; preserves chain for Is/As
	}
}
```

When `err` is already `*Error`, `Wrap` delegates to the method receiver. Otherwise the package converts `err` internally (callers never write `asError` themselves).

### `Unwrap` — error-chain inspection

```go
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.InnerError
}
```

`Unwrap` returns `InnerError` as `error`, so `errors.Is`, `errors.As`, and `%w` chains work unchanged.

### Stack trace capture

```go
// createStackTrace captures the current goroutine stack, skipping
// internal errors-package frames. Used by New, Wrap, and fmt.Errorf.
// The slice is ordered innermost caller first.
func createStackTrace() []StackFrame {
	// Walk runtime.Callers / runtime.CallersFrames and populate the slice.
	...
}
```

Stack traces are captured when an `*Error` is created through **`New`**, **`Wrap`**, or **`fmt.Errorf`** (see below). Other `error` implementations returned without conversion have no trace unless they embed or wrap `*Error`.

## Usage patterns

### Root error

```go
func openFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("open file: " + path)
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

var pathErr *Error
if errors.As(err, &pathErr) {
	fmt.Println(pathErr.Message)
	fmt.Println(pathErr.StackTrace)
}
```

## Interoperability

### `fmt.Errorf` and `%w`

**Today (upstream Go):** `fmt.Errorf` does **not** capture a stack trace. It returns either `errors.New(msg)` (no `%w`) or an opaque `fmt.wrapError` / `fmt.wrapErrors` (with `%w`).

**In this fork:** `fmt.Errorf` is updated to produce **`*errors.Error` when possible**, with a stack trace at the **`Errorf` call site**:

| Form | Result |
| ---- | ------ |
| `fmt.Errorf("msg")` | `*errors.Error` — same as `errors.New("msg")` (already delegated to `errors.New` today) |
| `fmt.Errorf("… %w", err)` (one `%w`) | `*errors.Error` — replaces `fmt.wrapError`; `Message` is the **full formatted string**; `InnerError` links the operand for `Is` / `As` |
| `fmt.Errorf("… %w … %w", e1, e2)` (multiple `%w`) | **`fmt.wrapErrors` unchanged** — multiple operands do not map to a single `InnerError`; no `*errors.Error` outer layer until a structured multi-cause type exists |

Example (single `%w`):

```go
if err := readFile(path); err != nil {
	return fmt.Errorf("readFile error: %w", err)
}
```

The returned value is `*errors.Error` with a fresh `StackTrace` at this line. `errors.Is` / `errors.As` traverse `InnerError` as with `Wrap`. The outer `Error()` string is still the full formatted message (including the wrapped error’s text from `%w`), matching today’s `fmt.wrapError` behavior.

Implementation sketch (`fmt` package):

```go
case 1:
	wrapped := a[p.wrappedErrs[0]].(error)
	err = &errors.Error{
		Message:    s,
		StackTrace: errors.CaptureStackTrace(), // or shared createStackTrace helper
		InnerError: errors.Link(wrapped),
	}
```

(`CaptureStackTrace` / `Link` names are illustrative; exact exports are an implementation detail.)

When you do **not** need `fmt` verbs in the message, **`errors.Wrap`** remains clearer:

```go
return errors.Wrap(err, "readFile error")
```

Both single-`%w` `fmt.Errorf` and `errors.Wrap` capture a stack trace at the call site. `Wrap` sets `Message` to the context string only; `fmt.Errorf` sets `Message` to the entire formatted output.

### Plain `error` values

Functions may still return `error`. Callers that receive a non-`*Error` value use `errors.Is` / `errors.As` as today. Converting at the boundary:

```go
if err != nil {
	return errors.New(err.Error()) // loses typed inner link unless err is already *Error
}
```

Prefer propagating `*Error` from APIs that own error construction, and use `%w` or explicit `InnerError` assignment when bridging packages.

### Relation to `T!` result types

[`T!`](result_types.md) lowers to `(T, error)`. An `int!` or function returning `error` may hold an `*errors.Error` like any other `error` implementation:

```go
func readCount() int! {
	n, err := parse()
	if err != nil {
		return err.Wrap("parse count") // *Error assignable to error slot
	}
	return n
}
```

## Formatting and display

`*Error` also implements **`fmt.Formatter`** for fine-grained control; default verbs delegate to `Error()` or `String()` as appropriate:

```go
func (e *Error) Format(s fmt.State, verb rune) {
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
		fmt.Fprintf(s, "%%!%c(*errors.Error)", verb)
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

Prefer **`errors.New`** for new root errors and sentinels. Every call captures a stack trace at the creation site; no API signature changes are required:

```go
var ErrNotFound = errors.New("not found")

func validate(id string) error {
	if id == "" {
		return errors.New("empty id")
	}
	return nil
}
```

Existing `errors.New("…")` call sites automatically gain stack traces once this stdlib update is in place.

### `fmt.Errorf` — stack traces without rewrites

Existing **`fmt.Errorf`** call sites also gain stack traces where the implementation can return `*errors.Error`:

- `fmt.Errorf("…")` — via `errors.New`, same as above.
- `fmt.Errorf("… %w", err)` — one `%w` operand returns `*errors.Error` at the `Errorf` site (replaces `fmt.wrapError`).

No migration required for typical wrap-and-return code. Use **`errors.Wrap`** when you want a short context label without `fmt` verbs; use **`fmt.Errorf`** when the message needs formatting (`%s`, `%d`, etc.).

```go
// Both capture a stack trace at this hop (single inner error):
return fmt.Errorf("load config %s: %w", path, err)
return errors.Wrap(err, "load config "+path)
```

### Prefer `errors.Wrap` for simple context

When the message is a fixed prefix plus a wrapped error, **`errors.Wrap`** is still recommended for clarity (no format string). **`fmt.Errorf` with `%w`** is equivalent for stack traces when you need formatting.

### Embed `errors.Error` in custom error types

Custom error structs should **embed `errors.Error`** so stack traces and wrapping come from the standard type while you add domain fields:

```go
// Before
type MyError struct {
	Msg  string
	Code int
}

func (e *MyError) Error() string { return e.Msg }

// After — recommended
type MyError struct {
	errors.Error
	Code int
}

func NewMyError(code int, message string) *MyError {
	var err MyError
	errors.NewCustom(&err.Error, message)
	err.Code = code
	return &err
}
```

`*MyError` implements `error` via the promoted `Error()` method.

See [Choosing how to create errors](#choosing-how-to-create-errors) for `errors.New`, `NewCustom`, and `NewMyError`.

Override `Error()` or `Unwrap()` on `*MyError` only when the default embedded behavior is not enough; otherwise rely on promotion.

For wrapping a custom error on the return path:

```go
if err != nil {
	return errors.Wrap(err, "operation failed")
}
```

If `err` is `*MyError`, `Wrap` adds an outer `*errors.Error` layer whose `InnerError` links back into the chain.

### Optional: `go fix`

Mechanical rewrites (e.g. `fmt.Errorf("… %w", err)` → `errors.Wrap(err, "…")`) may be provided as **`go fix`** analyzers in a later release. Manual migration following the patterns above is always sufficient.

## Naming: `Error`, not `error`

The structured type is named **`Error`** (capital E). It **cannot** be named **`error`**, because `error` is already a **predeclared interface** in Go’s universe block:

```go
type error interface {
	Error() string
}
```

Every package uses that identifier today — function results (`func f() error`), parameters, variables, type assertions, and `errors.As` targets. If the standard library (or the language) introduced a concrete **`error` struct** in place of the interface, the name would no longer denote an interface type. Existing code would fail to compile or change meaning in subtle ways:

```go
var err error = errors.New("fail")   // today: interface holding *errors.Error (*Error implements error)
func work() error { ... }            // today: any error implementation
errors.As(err, &target)              // today: inspect dynamic type via interface
```

Replacing the predeclared **`error` interface** with a struct (or redefining `error` as a single concrete type) is a **breaking language change**, not a drop-in stdlib swap. The ecosystem would need an **automated migrator** (e.g. `go fix`, a dedicated codemod, or IDE-driven rewrites) to retarget APIs, type annotations, and assertions across modules.

This proposal therefore keeps the interface as **`error`** and adds a separate concrete type **`errors.Error`**. Callers continue to use `error` in signatures; they opt into `*errors.Error` when they need `StackTrace`, `InnerError`, or `Wrap`.

## Design notes

- **One stack trace per layer** — wrapping adds a new `[]StackFrame` at the wrap site; inner errors retain their original traces.
- **`Message` is the layer label** — use short, stable text; put dynamic detail in `Message` or in wrapped inner messages consistently.
- **Nil-safe** — `(*Error)(nil).Error()` and `(*Error)(nil).String()` return `"<nil>"`; `Wrap` on `nil` behaves like `New`.
- **Not valid in upstream Go** — this extension applies to this fork’s standard library.

## Quick reference

| API | Purpose |
| --- | ------- |
| `errors.New(msg)` | Root `*Error` with stack trace |
| `errors.NewCustom[T](msg)` | Root custom `*T` embedding `errors.Error` |
| `errors.NewCustom(&e, msg)` | Fill embedded `*errors.Error` in place (extra domain fields) |
| `NewMyError(...)` | Custom `*MyError` with extra fields |
| `err.Wrap(msg)` | New outer layer; `InnerError = err` |
| `err.Error()` | Returns `Message` only |
| `err.String()` | Full serialization: inner chain + `Message` + `StackTrace` |
| `err.StackTrace.String()` | All frames at this layer, one per line |
| `err.Unwrap()` | Returns `InnerError` |
| `err.StackTrace` | `StackTrace` (`[]StackFrame`) at this layer |
| `fmt.Errorf("…")` | `*errors.Error` with stack trace (via `errors.New`) |
| `fmt.Errorf("… %w", err)` | `*errors.Error` with stack trace (one `%w`); multi-`%w` unchanged |
