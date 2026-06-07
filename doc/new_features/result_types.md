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
	return a
}
```

A `T!` function may return a single value `v`; the compiler treats it as `return v, nil`. You may still write `return v, nil` or `return zero, err` explicitly.

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

`T!` is also a **value type** (lowered to a struct with `value` and `err` fields), in addition to the function-result shorthand.

```go
var a int! = 0                            // value=0, err=nil
var b int! = errors.New("Some error")
if a.err == nil { doSomething(a.value) } else { handleError(a.err) }
val, err := a                              // destructure into value and error
```

Inside a function with result type `T!`, use `!.value` / `!.field` for propagation:

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
- Use `expr!.value` or `expr!.field` only in contexts where early-returning an error is valid for the enclosing function's signature (typically a `T!` result function).
- You can still do `if err != nil { panic(err) }` or `log.Fatal` as today.
