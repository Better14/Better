# New `?` Result Syntax

This document describes the new result-type and postfix-unwrap syntax based on `?`.

## Overview

Go now supports a result shorthand:

- `(T, error)` can be written as `T?`
- `expr?` unwraps a `(T, error)` expression

The goal is to reduce boilerplate for error propagation while preserving the same runtime behavior as explicit `if err != nil { ... }` checks.

## Function Result Type Shorthand

You can declare a function returning `(int, error)` as `int?`.

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
func myFunc() int? {
	a := myFunc2()?
	return a, nil
}
```

`int?` is semantically equivalent to `(int, error)`.

## Postfix `?` Operator

`expr?` can be used when `expr` has type `(T, error)` (or `T?`).

Behavior:

1. Evaluate `expr`
2. If `err != nil`, return early from the current function with:
   - zero value of the function's value result
   - the error
3. Otherwise, yield the unwrapped `T` value

This is similar to Rust's `?` operator in spirit.

## Conceptual Representation

A `T?` can be thought of as a pair:

```go
struct TValueOrErr {
	value T
	err   error
}
```

For example, `int?` corresponds conceptually to:

```go
struct intOrErr {
	value int
	err   error
}
```

This is a conceptual model for documentation; syntax-level behavior is defined by compiler lowering and type checking.

## Additional Intended Usage Patterns

The requested usage patterns are:

```go
var a := int?
if a.err != nil { a.value }
if a.err == nil { a.err } else { a.value }
var a, err := int?   // destructure into value and error
a?.someProperty      // early-return if error, then access property
```

These examples describe the intended ergonomics around result-style values and chained unwrapping.

## Notes

- `T?` should be treated as the canonical shorthand for `(T, error)`.
- `expr?` should only be used in contexts where early-returning an error is valid for the enclosing function's signature.

## Postfix `!` Force-Unwrap Operator

Go now also supports postfix `!` as a force-unwrap operator.

`expr!` evaluates a result-like expression and:

1. Panics with the error if the error is non-nil
2. Otherwise returns the underlying value

This is useful when failure is unexpected and should be fatal, similar to a forced unwrap.

### `(T, error)` Example

```go
func someFunc() (int, error) {
	// ...
}

var a = someFunc()! // if err != nil: panic(err); otherwise a is int
```

Equivalent behavior:

```go
tmp, err := someFunc()
if err != nil {
	panic(err)
}
var a = tmp
```

### `T?` Example

```go
func someFunc2() int? {
	// ...
}

var b = someFunc2()! // if result err != nil: panic(err); otherwise b is int
```

Equivalent conceptual behavior:

```go
tmp, err := someFunc2()
if err != nil {
	panic(err)
}
var b = tmp
```

### Notes

- `expr!` accepts values of type `(T, error)` and `T?`.
- `expr!` does not early-return; it panics on error.
- `expr!` should be used sparingly and only when panic-on-error is desired.

## Function and Method Overloading

Go now supports function and method overloading.

Multiple declarations with the same name are valid as long as their parameter lists differ by type and/or arity.

For example:

```go
func myFunc(a int) {
}

func myFunc(a int64) {
}

func myFunc(a int, b int) {
}
```

All of the above declarations are valid.

### Overload Resolution

At a call site, the compiler resolves the overload by matching argument count and argument types.

- `myFunc(10)` resolves to `func myFunc(a int)`
- `myFunc(int64Value)` resolves to `func myFunc(a int64)`
- `myFunc(1, 2)` resolves to `func myFunc(a int, b int)`

### Notes

- Overloading applies to both package-level functions and methods.
- Overload sets must be unambiguous for all valid calls.