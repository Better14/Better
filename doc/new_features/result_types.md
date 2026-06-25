# Result Types (`T!`)

## Overview

Go now supports a result shorthand:

- `(T, error)` can be written as `T!`
- `expr!` unwraps the success value and early-returns the error if `err != nil`
- `expr!.field` does the same, then accesses a field on the success value
- `err!` on a plain `error` early-returns if `err != nil` (no value to unwrap)

The goal is to reduce boilerplate for error propagation while preserving the same runtime behavior as explicit `if err != nil { ... }` checks.

## Function Result Type Shorthand

You can declare a function returning `(int, error)` as `int!`.

Before:

```go
func myFunc() (int, error) {
	a, err := myFunc2()
	if err != nil {
		return 0, err
	}
	return a, nil
}
```

After:

```go
func myFunc() int! {
	a := myFunc2()!
	return a
}
```

A `T!` function may return a single value `v`; the compiler treats it as `return v, nil`. You may still write `return v, nil` or `return zero, err` explicitly.

`int!` is semantically equivalent to `(int, error)`.

## Suffix binding with composite types

`!` is a **postfix suffix on the entire type expression** to its left (the same parsing rule as `?`). It does not bind only to the innermost type name.

| Written | Meaning |
| ------- | ------- |
| `BookRow!` | `(BookRow, error)` |
| `[]BookRow!` | `([]BookRow, error)` — slice result **or** one error for the whole operation |
| `[](BookRow!)` | `[]` of `BookRow!` — each element is a `BookRow!` value type (value + `err` fields) |
| `*BookRow!` | `(*BookRow, error)` |

So `[]BookRow!` is **not** “a slice where each element is `BookRow` or `error`”. It is the usual Go pattern: return a `[]BookRow` on success, or return a single `error` for the call.

Use **parentheses** when you want `!` to bind to an inner type before an outer constructor (`[]`, `*`, etc.) applies:

```go
func QueryBooks(...) []BookRow! {
	// ([]BookRow, error): one slice or one error
	return out
}

var perRow [](BookRow!) // [] of BookRow! value type — uncommon; not a function result shorthand
```

This matches pointer binding: `*T!` means `(*T, error)`, not `*(T!)`. See [Suffix binding with composite types](nullable_types.md#suffix-binding-with-composite-types) for the same rules with `?`.

## Postfix `!` and `!.field`

### `(T, error)` and `T!`

Use `!` or `!.field` when `expr` has type `(T, error)` or `T!`.

Behavior for `expr!`:

1. Evaluate `expr`
2. If `err != nil`, return early from the current function with:
  - zero value of the function's value result
  - the error
3. Otherwise, use the unwrapped `T` (or discard it when `expr!` is used as a statement)

Behavior for `expr!.field` is the same early-return on error, then access `.field` on the success value.

When you only need error propagation and do not use the success value, `expr!` may appear as a **statement** on its own. The unwrapped `T` is discarded, the same way a multi-value call like `conn.Exec(...)` may appear as a statement in standard Go:

```go
func setup(db *sql.DB) int! {
	conn.Exec("CREATE DATABASE mydb")!  // ok: check err, discard sql.Result
	return 0
}
```

You do **not** need `_ =` for this pattern. `_ = expr!` is still valid when you want to be explicit.

```go
func readName() string! {
	return fetch()!.name   // propagate error, else return User.name
}

func readUser() User! {
	u := fetch()!        // unwrap type, or early return on error
	return u
}

func someFunc() int! {
	return other()!     // redundant case. return value or error
}
```

### Plain `error`

When you already have an `error` value (not a `(T, error)` pair), use `err!` as a statement to early-return on failure.

```go
err := possibleError()  // error or nil
err!
```

Behavior for `err!`:

1. Evaluate `err`
2. If `err != nil`, return early from the current function with:
  - zero value of the function's value result
  - `err`
3. Otherwise, continue

Before:

```go
func doThing() int! {
	err := possibleError()
	if err != nil {
		return 0, err
	}
	return run()
}
```

After:

```go
func doThing() int! {
	err := possibleError()
	err!
	return run()
}
```

`err!` is only valid in a function that can return an error (typically a `T!` result). It does not produce a value; it is a control-flow statement like `return`.

## Interaction with `defer`

`expr!`, `expr!.field`, and `err!` lower to an ordinary `if err != nil { return ... }`. They are not a special exit path. **Any `defer` already registered in the function runs (in LIFO order) before an `!`-triggered early return completes**, the same as for an explicit `return` or `if err != nil { return zero, err }`.

```go
func example() int! {
	defer cleanup()       // runs even when other()! fails
	x := other()!
	return x
}
```

This is equivalent to:

```go
func example() (int, error) {
	defer cleanup()
	t0, t1 := other()
	if t1 != nil {
		return 0, t1   // defer runs here
	}
	return t0, nil
}
```

**A `defer` only runs if execution reaches its statement before an `!` fires.** If the failing `!` appears earlier in the function, later `defer`s are never registered:

```go
func saveFile(path string) error {
	f := os.Create(path)! // error here returns before defer is registered
	defer f.Close()
	_, err := f.WriteString("ok")
	return err
}
```

On the success path, `defer f.Close()` runs normally — including when a later statement returns an error through a plain `return err`.

## Conceptual Representation

A `T!` can be thought of as a pair:

```go
struct TValueOrErr {
	value T
	err   error
}
```

For example, `int!` corresponds conceptually to:

```go
struct intOrErr {
	value int
	err   error
}
```

This is a conceptual model for documentation; syntax-level behavior is defined by compiler lowering and type checking.

## Usage Patterns

`T!` is also a **value type** (lowered to a struct with `value` and `err` fields), in addition to the function-result shorthand.

```go
var a int! = 0                            // value=0, err=nil
var b int! = errors.New("Some error")
if a.err == nil { doSomething(a.value) } else { handleError(a.err) }
val, err := a                              // destructure into value and error
```

Inside a function with result type `T!`, use `!` / `!.field` for propagation:

```go
func example() int! {
	y := someFunc()!.someProperty
	return y
}
```

Assigning plain `T` to `T!` sets `value` and leaves `err` as `nil`. Assigning an `error` to `T!` sets `err` and leaves `value` at the zero value of `T`. You can also set `a.err` directly on an existing `T!` value.

### Assignability: `T!` is not `T`

`T!` and `T` are distinct types. A result value cannot be passed or assigned where a plain `T` is required without an explicit unwrap.

```go
func myPrint(a int) {
	fmt.Println(a)
}

func main() {
	var a int!
	myPrint(a) // compile error: int! is not assignable to int
}
```

To call `myPrint`, unwrap `a` in one of these ways:

**Error check** — only use the value when there is no error:

```go
func main() {
	var a int!
	if a.err == nil {
		myPrint(a.value)
	}
}
```

**Null coalescing** — supply a default when `err != nil`:

```go
func main() {
	var a int!
	myPrint(a ?? 0) // int! ?? int → int
}
```

When the left operand is `T!`, `??` uses the `.value` field if `err == nil`; otherwise it evaluates and uses the right-hand side (short-circuit). The result type is `T` when the right operand is `T`.

**Direct `.value` access** — reading `.value` without checking panics if `err != nil`:

```go
func main() {
	var a int!
	myPrint(a.value) // panics if a.err != nil
}
```

The same rules apply to assignment, return values, and other contexts that expect `T` rather than `T!`.

See [Null-coalescing operator (`??`)](nullable_types.md#null-coalescing-operator-) in nullable types for general `??` syntax; for `T!`, the left-hand side is treated as failed when `err != nil` (not when the value is `nil`).

## Notes

- `T!` is the canonical shorthand for `(T, error)` in function signatures and a value type elsewhere.
- Use `expr!`, `expr!.field`, or `err!` only in contexts where early-returning an error is valid for the enclosing function's signature (typically a `T!` result function).
- `expr!` and `err!` may be used as standalone statements when only error propagation is needed; the success value of `expr!` is discarded without requiring `_ =`.
- You can still do `if err != nil { panic(err) }` or `log.Fatal` as today.

