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

## If Expressions

Go supports `if` as an expression that evaluates to a value.

```go
a := if 5 < 6 { 1 } else { 2 }
```

Both branches must be expressions with compatible types. The result type is the common type of the branch expressions.

## Switch Expressions

Go supports `switch` as an expression that evaluates to a value.

```go
a := switch x {
case 1:
	"one"
case 2:
	"two"
default:
	"other"
}
```

Each case arm must be an expression (or a single expression after `:`). All arms must have compatible types. The result type is the common type of the case expressions.

## Lambda Syntax (`=>`)

For single-expression functions, Go supports arrow lambda syntax. Parameter types are inferred from context.

Before:

```go
func(a, b, c int) int { return a + b + c }
```

After:

```go
(a, b, c) => a + b + c
```

If the function body requires more than one expression or any statement, use the standard `func` syntax:

```go
func(a, b, c int) {
	sum := a + b + c
	return sum
}
```

### Notes

- `=>` lambdas are limited to a single expression body.
- Parameter types are inferred when the lambda appears in a typed context (e.g. assignment, argument, return).
- Multi-statement or multi-expression bodies must use `func(...) { ... }`.

## Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (e.g. deferred iteration until materialization)
- Minimal allocations; iterators and pipelines should avoid unnecessary intermediate slices

Step-by-step example:

```go
nums := []int{1, 2, 3, 4, 5}
evens := nums.Where(n => n%2 == 0)
doubled := evens.Select(n => n * 2)
first := doubled.First()
```

### Chained one-liners

Pipelines compose left-to-right; each stage is lazy until a terminal operator (`First`, `ToList`, `Sum`, etc.) runs.

```go
nums := []int{1, 2, 3, 4, 5, 6, 7, 8}

// filter → map → first
firstEvenDouble := nums.Where(n => n%2 == 0).Select(n => n * 2).First()

// filter → order → take
topThree := nums.Where(n => n > 2).OrderByDescending(n => n).Take(3).ToList()

// map → aggregate
sumOfSquares := nums.Select(n => n * n).Sum()

// skip → take → map
page := nums.Skip(10).Take(20).Select(n => fmt.Sprintf("%d", n)).ToList()

// any / all over a chain
hasLargeEven := nums.Where(n => n%2 == 0).Any(n => n > 100)
allPositive := nums.Select(n => n - 1).All(n => n >= 0)

// strings: filter → project → join
names := []string{"alice", "", "bob", "carol"}
line := names.Where(s => len(s) > 0).Select(s => strings.ToUpper(s)).Aggregate((a, b) => a + ", " + b)

// grouping (lazy until enumerated)
byMod := nums.GroupBy(n => n % 3).Select(g => (g.Key, g.Count())).ToList()

// distinct after transform
unique := nums.Select(n => n / 2).Distinct().OrderBy(n => n).ToList()

// first match or default
found := users.Where(u => u.Active).Select(u => u.Email).FirstOrDefault()
```

Predicate and projection arguments are typically single-expression lambdas using `=>`; parameter types are inferred from the LINQ method signature.

LINQ extensions are provided as methods on supported sequence types (slices, arrays, and other iterable types as defined by the standard library).

---

## Optional Features

The following features are planned or under consideration; they are not required for the core language extensions above.

### Default Function Arguments

Functions may declare default values for trailing parameters:

```go
func example(a int, b int = 3) {
	// ...
}
```

Call sites may omit arguments that have defaults:

```go
example(1)      // b is 3
example(1, 10)  // b is 10
```

### Operator Overloading

Types may define operators via special method syntax:

```go
func (m *Matrix) +{ /* m + other */ }
func (m *Matrix) +={ /* m += other */ }
func (m *Matrix) *{ /* m * other */ }
func (m *Matrix) []{ /* indexer: m[key] */ }
```

Supported operators and their method forms include (non-exhaustive):

- Arithmetic: `+`, `-`, `*`, `/`, `%`, and compound forms (`+=`, `-=`, etc.)
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=` (where defined)
- Indexing: `[]` for get/set via indexer methods

Overload resolution selects the receiver type’s operator method when the corresponding built-in operator is used with that type.