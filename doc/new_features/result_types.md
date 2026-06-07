# Result Types (`T!`)

## Overview

Go now supports a result shorthand:

- `(T, error)` can be written as `T!`
- `expr!.value` (or `expr!.field`) propagates an error from `expr` or accesses the success value

The goal is to reduce boilerplate for error propagation while preserving the same runtime behavior as explicit `if err != nil { ... }` checks. There is no postfix `expr!` that panics on error.

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
	a := myFunc2()!.value
	return a, nil
}
```

`int!` is semantically equivalent to `(int, error)`.

## Postfix `!.value` and `!.field`

Use `!.value` or `!.field` when `expr` has type `(T, error)` or `T!`.

Behavior for `expr!.value`:

1. Evaluate `expr`
2. If `err != nil`, return early from the current function with:
  - zero value of the function's value result
  - the error
3. Otherwise, use the `.value` field (the unwrapped `T`)

Behavior for `expr!.someField` is the same early-return on error, then access `someField` on the success value.

```go
func someFunc() int! {
	return other()!.value
}

func readUser() User! {
	u := fetch()!.value
	return u, nil
}

func readUser2() User! {
	u := fetch()!.value
	return u
}

func readName() string! {
	return fetch()!.name   // propagate error, else return User.name
}
```

Standalone `someFunc()!` (panic on error) is **not** supported.

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

```go
var a := int!                              // a has type int! (value + error)
if a.err != nil { _ = a.value }
if a.err == nil { _ = a.value } else { _ = a.err }
var a, err := int!                         // destructure into value and error
x := someFunc()!.value                     // propagate error or read .value
y := someFunc()!.someProperty              // propagate error or access a field
```

## Notes

- `T!` is the canonical shorthand for `(T, error)`.
- Use `expr!.value` or `expr!.field` only in contexts where early-returning an error is valid for the enclosing function's signature (typically a `T!` result function).
- Do not use `expr!` alone; it is not a panic unwrap. Handle unexpected failures with explicit `if err != nil { panic(err) }` or `log.Fatal` as today.
