# New `!` and `?` Types

This document describes result types (`T!`), nullable types (`T?`), and postfix `!.value` / `!.field` access for error propagation.

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

## Nullable Types (`T?`)

`T?` on a value type `T` means **either a `T` or `nil`** — an optional value with no error channel. This is separate from `T!`, which means **value or `error`**.


| Syntax | Meaning                                    |
| ------ | ------------------------------------------ |
| `int?` | `int` or `nil` (nullable)                  |
| `int!` | `int` or `error` (result / `(int, error)`) |


Nullable types are useful for primitives and structs that cannot otherwise hold `nil` in Go. Reference types (`*T`, `map`, `slice`, `chan`, `func`, `interface`) are already “nullable” via `nil`; `T?` is most important for `int`, `bool`, `float64`, struct types, etc.

### Declaration and assignment

```go
var a int? = 5
var b int? = nil

var c int?       // nil by default
c = 10
c = nil
```

### Functions returning nullable types

A function may return `T?` and supply either a value or `nil`:

```go
func myFunc(num int) int? {
	if num < 5 {
		return num
	}
	return nil
}
```

Returning a plain `T` where `T?` is expected wraps the value (same idea as assigning `5` to `int?`).

### Checking for a value

```go
v := myFunc(3)
if v == nil {
	// no value
} else {
	use(v) // v is int in this branch
}
```

Conceptually, `int?` is a optional value (value + “has value” flag). The compiler may lower it to a struct or pointer; the source-level model is **value or `nil`**, not value + `error`.

### Null-conditional operator (`?.`)

The `?.` operator (null-conditional / “Elvis” access) short-circuits when the left-hand value is `nil`. No panic is raised; the rest of that access or assignment chain is skipped.

**Chained reads** — result type is nullable; any `nil` in the chain yields `nil` for the whole expression:

```go
var a string? = someObject?.someProp?.someProp2
```

If `someObject` is `nil`, the chain stops and `a` is `nil` (without evaluating `.someProp` or `.someProp2`). If `someObject` is set but `someProp` is `nil`, `a` is `nil` likewise.

**Assignments** — if any `?.` step on the left-hand side is `nil`, the assignment does not run:

```go
someObject?.someProp = "value"
```

If `someObject` is `nil`, there is no write to `someProp` and no error. Execution continues with the next statement (only that assignment is skipped).

Equivalent read logic:

```go
var a string?
if someObject == nil {
	a = nil
} else if someObject.someProp == nil {
	a = nil
} else {
	a = someObject.someProp.someProp2
}
```

Equivalent assignment guard:

```go
if someObject != nil {
	someObject.someProp = "value"
}
```

`?.` applies to fields, methods, and indexers where supported (`obj?.method()`, `arr?.[i]`). It is for **nullable / nil** values only, not for `T!` error results (use `!.value` / `!.field` there).

Do not confuse:


| Form          | Role                                               |
| ------------- | -------------------------------------------------- |
| `int?`        | Type: `int` or `nil`                               |
| `expr?.field` | Operator: access `field` only if `expr` is non-nil |


### Null-coalescing operator (`??`)

`??` picks the left-hand value when it is non-`nil`; otherwise it evaluates and uses the right-hand side. The right-hand side is evaluated **only** when the left is `nil` (short-circuit).

```go
var n int = count ?? 0                    // count is int?; n is 0 when count is nil
var s string = name ?? "anonymous"        // string? ?? string → string
var label string = obj?.title ?? "untitled"
```

With `?.` and `??` together:

```go
var a string? = someObject?.someProp
var b string = a ?? "default"             // b is string, never nil here
var c string = someObject?.someProp ?? "default"
```

**Chaining** — `??` is left-associative:

```go
var x int = first ?? second ?? 0
// same as (first ?? second) ?? 0
```

If `first` is non-`nil`, neither `second` nor `0` is evaluated.

**Types** — the result type is the non-nullable `T` when the right operand is `T` and the left is `T?`. Both operands may be nullable if the fallback is also optional:

```go
var m int? = a ?? b   // a, b are int?; m is nil only if both are nil
```

Equivalent:

```go
var n int
if count == nil {
	n = 0
} else {
	n = count // unwrap int? to int
}
```

`??` applies to nullable types and other nil-able values (`*T`, maps, slices, pointers). It does **not** apply to `T!` error results; handle errors with `if err != nil`, `!.value`, or explicit checks.


| Form               | Role                                                  |
| ------------------ | ----------------------------------------------------- |
| `int?`             | Type: `int` or `nil`                                  |
| `expr?.field`      | Null-conditional access                               |
| `expr ?? fallback` | Null-coalescing: `expr` if non-`nil`, else `fallback` |


### Notes

- Do not confuse `int?` (nullable type) with `int!` (result type). They use different suffixes on purpose.
- Do not confuse `?.` (null-conditional) or `??` (null-coalescing) with `!.` (error propagation on `T!`).
- `T?` does not support `!.value` error propagation; that syntax applies only to `T!` / `(T, error)`.
- Nullable defaults in function parameters (e.g. `x int? = nil`) follow the same compile-time constant rules as other default arguments when/if defaults are added for nullable parameters.

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

### IDE support (gopls)

When you type a call such as `myFunc(`, signature help lists **every** overload signature. Diagnostics for invalid calls include the available overload signatures in the error message.

### Errors

The type checker reports:

- **Redeclared** — two overloads with the same parameter types (`redeclared function f` / `redeclared method m`).
- **No matching overload** — no overload accepts the argument types at a call site.
- **Ambiguous overloaded call** — more than one overload fits equally well (for example, two overloads taking different named types with the same underlying type and an untyped constant argument).

### Notes

- Overloading applies to both package-level functions and methods.
- Overload sets must be unambiguous for all valid calls.
- The blank identifier `_` is not overloadable; duplicate `func _()` declarations are still invalid.

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

## Default Function Arguments

Go supports default parameter values, following these rules: only trailing parameters may have defaults, and once one parameter has a default, every parameter to its right must also have a default.

### Declaration

```go
func myFunc(a, b, c int = 5, d int = 7) {
	// a, b, c are int; c defaults to 5; d defaults to 7
}
```

Mixed required and optional parameters (optional parameters are always on the right):

```go
func connect(host string, port int = 443, timeout time.Duration = 30*time.Second) {
	// ...
}

func log(msg string, level int = 1) {}
```

Invalid (a required parameter may not follow an optional one):

```go
// func bad(a int = 1, b int) {}  // compile error
// func bad(a int, b int = 2, c int) {}  // compile error: c has no default but follows b
```

### Call sites

Arguments are filled from left to right. Omitted trailing arguments use their defaults:

```go
myFunc(1, 2)           // a=1, b=2, c=5, d=7
myFunc(1, 2, 3)        // a=1, b=2, c=3, d=7
myFunc(1, 2, 3, 4)     // a=1, b=2, c=3, d=4
myFunc(1, 2, 0, 9)     // a=1, b=2, c=0, d=9

connect("example.com")                    // port 443, timeout 30s
connect("example.com", 8080)              // timeout 30s
connect("example.com", 8080, time.Second) // all explicit
```

You may not skip a non-trailing argument while passing a later one (no “hole” syntax like `f(1, , 3)`).

### Methods and overloads

Default arguments apply to methods and work with function overloading: each overload has its own default list; overload resolution uses the argument count and types actually passed at the call site.

```go
func (s *Server) Start(addr string, port int = 80) {}

s.Start("localhost")     // port 80
s.Start("localhost", 443)
```

### Default values (compile-time only)

Default values must be known at compile time. Parameters may use literals and **constant expressions** built from them (not arbitrary runtime code).

Implementation is staged in two layers; both are evaluated at compile time and inlined at call sites when arguments are omitted.

#### Tier 1 — literals and named constants

- Untyped and typed literals: `42`, `3.14`, `"ok"`, `true`, `false`
- `nil` where valid for the parameter type (pointer, map, slice, chan, func, interface)
- Identifiers naming **constants** in scope: package `const`, file `const`, or imported const

```go
const (
	DefaultPort   = 443
	DefaultLevel  = 1
	DefaultWindow = 30 * time.Second
)

func connect(host string, port int = DefaultPort, timeout time.Duration = DefaultWindow) {}
func log(msg string, level int = DefaultLevel) {}
```

#### Tier 2 — constant expressions (Go `const` rules)

The default expression may be any expression that is legal in a Go `const` declaration with the parameter’s type. The compiler uses the same constant evaluation as for `const` (including typed constants and conversions).

Allowed examples:

```go
func f(n int = 1 << 20) {}
func g(d time.Duration = 30 * time.Second) {}
func h(s string = "go" + "lang") {}
func k(addr string = string(DefaultIP)) {} // conversion of const

type Port int
func listen(p Port = Port(8080)) {} // typed constant + conversion
```

### Valid examples

#### Simple primitives and constants

```go
func Log(message string, level int = 1, verbose bool = false) {
	// ...
}

const DefaultPort = 8080

func Connect(host string, port int = DefaultPort) {
	// ...
}
```

#### Enum and nil defaults

Use an `enum` type for mode defaults (see [Enums](#enums)). Pointer parameters use `nil` where a nullable default is intended.

```go
enum Mode {
	Read
	Write
	Both
}

func Open(path string, mode Mode = Mode.Read) {
	// ...
}

func Save(path *string = nil) {
	// nil means “not provided”
}
```

#### Compile-time expression results (tier 2)

```go
const Base = 2

func Multiply(x int, factor int = Base*3) {
	// allowed: constant expression
}
```

#### Optional parameters after required ones

```go
func Send(to, message string, urgent bool = false) {
	// required to, message; optional urgent on the right
}
```

### Invalid examples

#### Non-constant defaults (compile-time error)

```go
now := time.Now()

// func Schedule(t time.Time = now) {} // ERROR: 'now' is not a compile-time constant

// func Schedule(t time.Time = time.Now()) {} // ERROR: call not constant
```

#### Instance members as defaults (compile-time error)

```go
type C struct {
	x int
}

// func (c *C) M(a int = c.x) {} // ERROR: instance field cannot be used in default
```

#### Method calls or new objects as defaults (compile-time error)

```go
// func F(s string = GetDefault()) {}              // ERROR: function call not allowed
// func G(l []int = make([]int, 0)) {}             // ERROR: make not allowed
// func H(m map[string]int = map[string]int{}) {}  // ERROR: composite literal allocation not constant
```

Runtime defaults (`make`, `new`, non-const calls, package `var`s) are **not** supported in v1. See tier 1 and tier 2 above.


| Tier | Allowed in defaults                                             |
| ---- | --------------------------------------------------------------- |
| 1    | Literals, `nil`, named `const`                                  |
| 2    | Any Go constant expression (same rules as `const` declarations) |
| —    | Function calls, `make`, `new`, mutable `var`s                   |


### Other notes

- Default expressions are type-checked against the parameter type; untyped constants follow the same conversion rules as in `const` declarations.
- A default may not refer to other parameters of the same function.
- Default arguments are not supported on `=>` lambdas; use a named `func` or a wrapper.
- Not valid in upstream Go.

## Generic Methods (Go 1.27)

Go 1.27 extends generic types with **methods whose receivers are generic**. This is the mechanism behind C#-style “attach query methods to any enumerable” — but Go expresses it through **named generic types** and (for built-in slices) **compiler desugaring**, not C# `static` extension methods.

### Methods on generic types

If the receiver base type is generic, the receiver specification must declare matching type parameters. Those parameters are in scope for the method body (like type parameters on a generic struct):

```go
type Pair[A, B any] struct {
	a A
	b B
}

func (p Pair[A, B]) Swap() Pair[B, A] {
	return Pair[B, A]{a: p.b, b: p.a}
}
```

A method on `Lazy[T]` uses the same pattern:

```go
type Lazy[T any] struct { /* iterator state */ }

func (l Lazy[T]) Where(pred func(T) bool) Lazy[T] {
	// filter l using pred; return a new lazy pipeline
}
```

Here `T` comes from the receiver `Lazy[T]` — you do **not** write a separate `[T any]` on the method when `T` is already declared by the receiver.

Invalid (receiver must be a defined type, not a bare type parameter):

```go
func (o *T) Where(pred func(T) bool) Lazy[T]  // invalid
```

### Generic methods (extra type parameters on the method)

A method may declare **additional** type parameters after the method name, like a generic function bound to a receiver. This mirrors C# `Select` projecting to a new element type:

```go
type List[E any] []E

func (l List[E]) Select[F any](f func(E) F) List[F] {
	r := make(List[F], len(l))
	for i, x := range l {
		r[i] = f(x)
	}
	return r
}
```

Such a declaration is a **generic method**. It must be **instantiated** (explicitly or by inference) before it can be called, the same as generic functions.

### C# extension methods vs Go

In C#, `Where`, `Select`, etc. are **extension methods** on `IEnumerable<T>` — any type implementing that interface picks them up.

Go does not have `IEnumerable<T>`, but this fork adds **[Extension Methods](#extension-methods)** — the general mechanism C# uses, adapted for Go generics and `iter.Seq[T]`. Until extensions are implemented, enumerables also use the legacy paths below:


| Enumerable shape | How methods attach | Example |
| ---------------- | ------------------ | ------- |
| Slice / array `[]T` | **Extension methods** (planned) or legacy compiler **desugar** to `linq` when `import "linq"` | `a.Where(s => s == "a")` |
| `iter.Seq[T]` | **Extension methods** on `iter.Seq[T]` | `seq.Where(pred)` |
| Named generic sequence type | Real **receiver methods** on the type | `func (l Lazy[T]) Where(…)`, `func (l list[T]) Where(…)` |
| Other iterables | Extension methods after adaptation, or convert then chain | `set.Values().Where(…)` |


Go still does **not** allow methods on `[]T` itself (slice types are not defined types). To add methods directly in library code, use a defined generic type such as `type List[T any] []T` or the std `list[T]`, `Lazy[T]`, etc.

### Limitations (Go 1.27 and this fork)

- **No methods on slice types.** Use a type alias/definition (`type List[E any] []E`), a wrapper (`Lazy[T]`), or slice LINQ desugaring (below).
- **Receiver base type** must be a defined type in the same package; it cannot be a pointer or interface type, and generic aliases have restrictions (see the language spec).
- **Generic methods with method-local type parameters** (e.g. `Select[F any]` on `Lazy[T]`) are part of the Go 1.27 language, but the compiler in this fork may **ICE** when exporting some generic methods from generic types (`internal compiler error` in `noder/writer.go`). Until that is fixed, operations that need an extra type parameter (`Select`, `OrderBy` with key type `K`, `GroupBy` with key type `K`) may remain **package functions** (`linq.Select`, `linq.LazySelectBy`, …) even when simpler methods (`Where`, `Take`, `Skip`) work as receivers.
- **Instantiation:** generic methods must be instantiated; type inference at the call site applies when the compiler can infer method type arguments from arguments (same rules as generic functions).
- **Not in upstream Go** before 1.27.

### Relation to LINQ

Built-in LINQ syntax (`nums.Where(…).Select(…).ToList()`) is intended to be implemented with **[Extension Methods](#extension-methods)** on `iter.Seq[T]` (with automatic `slices.Values` for `[]T`). Until that lands, the compiler also uses:

1. **Legacy desugaring** for slices/arrays (and continued chains on `linq.Lazy[T]`) via a fixed `linqMethods` table.
2. **Receiver methods** on `linq.Lazy[T]` and container types where supported.
3. **Package functions** in `import "linq"` as the lowering target.

See [Built-in LINQ](#built-in-linq) below for usage examples.

## Extension Methods

Extension methods use **ordinary `func` syntax with a receiver**. There is no `extension` keyword. The compiler classifies the declaration from the **receiver type** and resolves **method call syntax** at use sites.

Requires Go 1.27+ (generic receivers, generic methods). Not valid in upstream Go.

### Goals

**Any type — with or without generic arguments:**

```go
func (i int) Square() int { return i * i }

func (s mypkg.String) Length() int { return len(s) } // mypkg.String defined elsewhere
```

**Foreign struct — extension in another package:**

```go
// person/person.go
package person

type Person struct { Name string }

// personext/hello.go
package personext

import "person"

func (p person.Person) Hello() string {
    return "Hi, my name is " + p.Name
}

// app/app.go
package app

import (
    "person"
    "personext"
)

func myFunc() {
    a := person.Person{}
    a.Hello() // method syntax — not personext.Hello(a)
}
```

**Generics / iter:**

```go
// query/where.go
func (seq iter.Seq[T]) Where[T any](pred func(T) bool) iter.Seq[T] { ... }

// app.go
import "query"

a := []string{"a", "b", "c"}
b := a.Where(s => s == "a")
```

### How a declaration becomes an extension

A `func` with a receiver is an **extension** when it cannot be a normal Go method on that receiver, **or** the receiver’s named base type is defined in **another** package.

| Situation | Example | Kind |
| --------- | ------- | ---- |
| **Predeclared** type (`int`, `string`, `bool`, …) | `func (i int) Square() int` | **Extension** |
| **Composite** with type params | `func (s []T) Where(...)` | **Extension** |
| **Named type in another package** | `func (p person.Person) Hello() string` | **Extension** |
| **Named generic type in another package** | `func (seq iter.Seq[T]) Where(...)` | **Extension** |
| **Named type in this package** | `func (p Person) Greet() string` in `package person` | **Ordinary method** |

Extensions work **with or without** type arguments on the receiver: `Person`, `*Person`, `[]T`, `iter.Seq[T]`, `int`, `mypkg.String`, etc.

There is no opt-in marker.

### Syntax (same as `func` + receiver)

Extensions use the normal method declaration form:

```go
MethodDecl = "func" Receiver MethodName [ TypeParameters ] Signature [ FunctionBody ] .
Receiver   = "(" identifier ReceiverType ")" .
```

#### Type parameters on the receiver

The receiver type may use type parameters that are **declared by the receiver**, matching Go 1.27’s rules for methods on generic types—extended so the base type may live in another package:

```go
func (seq iter.Seq[T]) Where(pred func(T) bool) iter.Seq[T]
```

Here `T` is declared by `(seq iter.Seq[T])` and is in scope for the whole declaration (signature + body). The type parameter corresponds to `V` in `iter`’s definition `type Seq[V any] func(...)`.

#### Type parameters on the method name

Method-local type parameters appear after the name, as for [generic methods](#generic-methods-go-127):

```go
func (seq iter.Seq[T]) Select[U any](fn func(T) U) iter.Seq[U] { ... }
```

`T` from the receiver; `U` introduced on `Select`.

#### Restating receiver type parameters (`Where[T any]`)

When a method uses the receiver’s type parameter `T`, the method name must **explicitly** restate `T` in its type parameter list. The first type parameter must be that same `T` (with a constraint such as `any`); additional parameters follow for new names.

```go
func (seq iter.Seq[T]) Where[T any](pred func(T) bool) iter.Seq[T]
func (seq iter.Seq[T]) Select[T, U any](fn func(T) U) iter.Seq[U]
```

Rules:

- The first method type parameter must be the receiver’s `T` (same name); it binds to the receiver type parameter, not a new declaration.
- Further method type parameters introduce new names (`Select[T, U any]`, `OrderBy[T, K cmp.Ordered]`, …).
- Omitting `[T …]` when the signature uses `T` is an error on **extension** receivers such as `[]T`.
- On ordinary generic receivers (`Lazy[T]`, `iter.Seq[T]`), restating `T` is optional but supported the same way.
- Parameter names in function types (`pred func(a T) bool`) are optional; `a` is not related to outer variables.

Invalid:

```go
func (s []T) Where(pred func(T) bool) any              // must be Where[T any](…)
func (seq iter.Seq[T]) Where[U any](pred func(T) bool) iter.Seq[T]  // first name must be T
func (o *T) Where[T any](pred func(T) bool) any       // *T not a defined base type
```

#### Direct extension on slices

```go
func (s []T) Where[T any](pred func(T) bool) iter.Seq[T] {
    return Where(slices.Values(s), pred)
}
```

`T` is declared by `[]T` in the receiver and must be restated on the method name when used in the signature. No `import` of a wrapper type required at call sites.

### Visibility and imports

At the call site you always use **method syntax**: `a.Hello()`, not `personext.Hello(a)`.

| Import | Required in calling file? |
| ------ | ------------------------- |
| Type package (`import "person"`) | **Yes** — to name `person.Person` |
| Extension package (`import "personext"`) | **Yes** — so the compiler can resolve extensions on `person.Person` |
| Dot import (`import . "personext"`) | **No** — never required |

A normal import of the extension package is enough. You do **not** qualify calls as `personext.Hello(a)` and you do **not** need a dot import.

```go
import (
    "person"
    "personext"
)

a := person.Person{}
a.Hello()
```

The extension package must appear in **that file’s** import block (transitive imports are not enough: importing only `person` does not bring in `personext`). Export data from `personext` lists which receiver types and methods it extends.

#### Resolution for `x.M(args)`

1. **Instance method** on `x`’s type — always wins.
2. **Extension methods** named `M` from packages **imported in the current file** whose receiver type matches `x`.
3. If multiple extensions match → **ambiguity error**.

Extensions are resolved at **compile time**, not via reflection.

### Receiver matching and adaptation

Extension `func (seq iter.Seq[T]) Where(...)` matches:

| Value type `x` | Behavior |
| -------------- | -------- |
| `iter.Seq[T]` | Direct match |
| `[]T`, `[N]T` | **Adapt:** pass `slices.Values(x)` as the synthetic first argument |
| Other | No match unless assignable to receiver after inference |

```go
a.Where(s => s == "a")
// → query.Where(slices.Values(a), s => s == "a")
```

Optional slice extension `(s []T) Where` avoids adaptation at the cost of a second declaration.

### Lowering (implementation model)

An extension is compiled as a generic function in its package; the receiver becomes the first parameter:

```go
// source (extension syntax)
func (seq iter.Seq[T]) Where[T any](pred func(T) bool) iter.Seq[T] { ... }

// compiled shape (conceptual)
func Where[T any](seq iter.Seq[T], pred func(T) bool) iter.Seq[T] { ... }
```

Call site:

```go
x.Where(args)  →  query.Where(/* adapted */ x, args...)
```

The type checker rewrites the AST to a package-level call.

### Cross-package `Person` example

```go
// person/person.go
package person

type Person struct { Name string }
```

```go
// personext/hello.go
package personext

import "person"

func (p person.Person) Hello() string {
    return "Hi, my name is " + p.Name
}
```

```go
// app/app.go
package app

import (
    "person"
    "personext"
)

func Run() {
    a := person.Person{Name: "Ada"}
    a.Hello()
}
```

### `int` / `String` extensions

```go
// builtinext/int.go
package builtinext

func (i int) Square() int { return i * i }
```

```go
// app.go
import "builtinext"

func example() {
    x := 5
    y := x.Square() // 25 → builtinext.Square(x)
}
```

### LINQ / `iter.Seq` example

```go
// query/where.go
package query

import "iter"

func (seq iter.Seq[T]) Where[T any](pred func(T) bool) iter.Seq[T] { ... }
```

```go
// main.go
package main

import "query"

func main() {
    a := []string{"a", "b", "c"}
    b := a.Where(s => s == "a")
}
```

### Parser and type checker changes

1. **Classifier** — predeclared, composite, or foreign receiver → `Extension` object.
2. **Export data** — receiver type identity, method name, type params, defining package.
3. **Selector** — instance method first, then extensions from **imported** extension packages; desugar to `pkg.M(x, args…)`.
4. **Adaptation** — `[]T` / `[N]T` → `slices.Values` when extension is on `iter.Seq[T]`.

### Relation to generic methods and LINQ

- **Ordinary method:** `func (p Person) Greet()` in `package person`.
- **Extension:** `func (p person.Person) Hello()` in `package personext`.
- **LINQ:** implement as extensions on `iter.Seq[T]` / `[]T`; remove legacy `linqMethods` desugar.

### Limitations

- **Ambiguity:** two imported extension packages define the same method on the same receiver type → error.
- **Instance methods win** over extensions.
- **Import required:** the calling file must import each extension package it relies on (normal import, not dot import).
- **Not in upstream Go.**

## Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics. Prefer **[Extension Methods](#extension-methods)** for new code; `import "linq"` currently enables legacy method desugaring on slices and `linq.Lazy[T]`.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (e.g. deferred iteration until materialization)
- Minimal allocations; iterators and pipelines should avoid unnecessary intermediate slices
- Target model: extensions on `iter.Seq[T]` with `[]T` adaptation; see [Extension Methods](#extension-methods)

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

```go
import "linq"

nums := []int{1, 2, 3, 4, 5}
evens := nums.Where(n => n%2 == 0)       // desugared to linq.Where
doubled := evens.Select(n => n * 2)    // linq.LazySelectBy on Lazy[int]
first := doubled.First()                 // terminal: materializes one element
```

LINQ extensions are provided as methods on supported sequence types (slices, arrays, `linq.Lazy[T]`, and other iterable types as defined by the standard library). Under the hood, slice/array calls desugar to `linq` functions; named types such as `Lazy[T]` may define real receiver methods where supported.

## Data Structures

Standard Go does not provide a built-in `set` type, queue/stack abstractions, or a generic binary tree. It does ship lower-level building blocks:


| Need                | Standard library today                                           |
| ------------------- | ---------------------------------------------------------------- |
| Growable sequence   | `[]T` + `append` (must reassign: `s = append(s, x)`)             |
| Set-like membership | `map[T]struct{}` (manual; no literal, no set algebra)            |
| Doubly linked list  | `container/list` (not typed; not a dedicated queue/stack API)    |
| Min-heap            | `container/heap` (you implement `heap.Interface`; min-heap only) |
| Max-heap            | `container/heap` with inverted `Less`                            |
| Binary tree         | Not in the standard library                                      |


This fork adds first-class container types in the standard library (or as built-in generic types) with literal syntax where noted below.

### List

A `list` is a growable, ordered sequence (like a slice) with methods that mutate in place. You no longer need to reassign the result of `append`:

Before (slice):

```go
var nums []int
nums = append(nums, 1)
nums = append(nums, 2, 3)
```

After (`list`):

```go
nums := list.New[int]()
nums.Append(1)
nums.Append(2, 3)   // variadic; same as multiple appends

names := list.Of("a", "b", "c")  // from values
names.Append("d")
```

**Literal syntax** (optional; mirrors slice literals):

```go
a := list[int]{1, 2, 3}
b := list[int]{}           // empty
```

Conversion and indexing:

```go
s := a.ToSlice()           // []int — share or copy per implementation
c := list.FromSlice(s)

x := a.At(0)               // or a[0] if indexer syntax is enabled
a.Set(1, 99)
n := a.Len()
a.Insert(1, 42)            // insert at index
a.RemoveAt(1)
last, ok := a.Pop()        // remove and return last element
a.Clear()
```

`Append` returns nothing (or returns `*list[T]` for chaining, e.g. `nums.Append(1).Append(2)`). Slices remain valid and interoperate via `ToSlice` / `FromSlice`; use `list` when you want method-style growth without `s = append(s, x)`.

LINQ methods (`Where`, `Select`, `OrderBy`, etc.) are defined on `list[T]` the same as on slices.

### LinkedList

`LinkedList[T]` is a doubly linked, ordered sequence. Unlike `list[T]` (slice-backed), inserting or removing in the middle does not shift a backing array; growth does not trigger slice reallocation.

Standard Go’s `container/list` is untyped (`Value any`) and uses external `*list.Element` handles. `LinkedList[T]` is generic and keeps nodes internal unless you opt into cursor APIs.

```go
ll := linkedlist.New[int]()
ll.Append(1)           // tail
ll.Append(2, 3)
ll.Prepend(0)          // head

ll.PushFront(99)       // aliases for head/tail
ll.PushBack(100)
v := ll.PopFront()     // 99
w := ll.PopBack()      // 100

n := ll.Len()
first := ll.Front()    // *int or (int, bool) — peek head value
last := ll.Back()

ll.InsertAt(2, 42)     // by index; O(n) walk; prefer node APIs when hot
ll.RemoveAt(1)
ll.RemoveValue(42)     // first matching element

for _, x := range ll.Iterate() {  // or range ll
	_ = x
}

s := ll.ToSlice()      // snapshot in linked order
other := linkedlist.FromSlice([]int{4, 5, 6})
```

**When to use which**


|                           | `list[T]`                             | `LinkedList[T]`                      |
| ------------------------- | ------------------------------------- | ------------------------------------ |
| Backing                   | Dynamic array (slice)                 | Doubly linked nodes                  |
| Index access `At(i)`      | O(1)                                  | O(n)                                 |
| Append / pop at end       | O(1) amortized                        | O(1)                                 |
| Insert / remove at front  | O(n) shift                            | O(1)                                 |
| Insert / remove in middle | O(n) shift                            | O(1) with node cursor; O(n) by index |
| Memory                    | Contiguous; less overhead per element | Pointer per node; extra allocations  |


Use `LinkedList[T]` for frequent front/middle edits, stable iterators while mutating elsewhere (with cursor API), or algorithms that splice sublists. Use `list[T]` for index-heavy work and cache-friendly sequential access.

LINQ methods apply on `LinkedList[T]` via iteration or after `ToSlice()`, depending on implementation.

### Set

A `set` is an unordered collection of unique elements. Elements must be `comparable`.

**Literal syntax** (mirrors `[]T{...}` but uses `{}` instead of `[]`):

```go
a := {}string{"foo", "bar", "bar", "baz"}
// len(a) == 3; duplicate "bar" is dropped at construction
```

Yes — `{}T{ ... }` is the intended literal form for this extension. It is not valid in upstream Go. An empty set is written `{}string{}`.

Other construction and use:

```go
b := set.Make[int]()           // empty set
b.Add(1)
b.Add(2)
b.Contains(1)                  // true
b.Delete(2)

c := {}int{1, 2, 3}
d := c.Union({}int{3, 4, 5})   // {}int{1, 2, 3, 4, 5}
e := c.Intersect({}int{2, 99}) // {}int{2}

for v := range a.Values() {    // or range a
	_ = v
}
```

Set operations: `Add`, `Delete`, `Contains`, `Len`, `Union`, `Intersect`, `Difference`, `Subset`, `Equal`, and iteration. LINQ-style methods (`Where`, `Select`, etc.) apply when converting via `.Values()` or when defined on the set type.

### Queue (FIFO)

```go
q := queue.New[int]()
q.Enqueue(1)
q.Enqueue(2)
v := q.Dequeue()   // 1
n := q.Len()
front := q.Peek()  // 2; does not remove
```

Use for breadth-first traversal, work queues, and ordered processing. Not a substitute for channels when you need concurrency safety; use channels for goroutine communication.

### Stack (LIFO)

```go
s := stack.New[int]()
s.Push(10)
s.Push(20)
v := s.Pop()       // 20
top := s.Peek()    // 10
```

Use for depth-first traversal, undo buffers, and expression parsing.

### Min-heap and max-heap

Typed heaps replace the boilerplate of implementing `heap.Interface` by hand.

```go
h := minheap.New[int]()   // smallest int at top
h.Push(5)
h.Push(1)
h.Push(3)
x := h.Pop()              // 1

mh := maxheap.New[string]()
mh.Push("a")
mh.Push("z")
mh.Pop()                  // "z"
```

Supports `Push`, `Pop`, `Peek`, `Len`, `Fix` (after changing a stored element), and optional `PushHeap`/`PopHeap` batch patterns. `container/heap` remains available for custom ordering; prefer `minheap` / `maxheap` when the element type is the priority key.

### Binary tree

A generic ordered binary tree (typically BST) for keyed lookup and ordered traversal:

```go
t := tree.New[int, string]()  // key int, value string
t.Insert(2, "two")
t.Insert(1, "one")
v, ok := t.Search(2)          // "two", true
t.Inorder(func(k int, v string) { /* sorted by k */ })
t.Delete(1)
```

Operations: `Insert`, `Search`, `Delete`, `Min`, `Max`, `Inorder`, `Preorder`, `Postorder`, `Len`, `Height`. Balanced variants (e.g. AVL/red-black) may be provided as `tree.NewBalanced` or a separate `balancetree` package depending on implementation.

### Summary


| Type                  | Literal              | Standard Go equivalent                     |
| --------------------- | -------------------- | ------------------------------------------ |
| `list`                | `list[T]{...}`       | `[]T` + `append`                           |
| `LinkedList`          | — (`linkedlist.New`) | `container/list` (`any`, manual `Element`) |
| `set`                 | `{}T{...}`           | `map[T]struct{}`                           |
| `queue`               | — (`queue.New`)      | slice + mutex, or `LinkedList` discipline  |
| `stack`               | — (`stack.New`)      | slice, or `container/list`                 |
| `minheap` / `maxheap` | —                    | `container/heap` + custom `Less`           |
| `tree`                | —                    | third-party or hand-rolled                 |


---

## Enums

Go supports Rust-style **algebraic enums** (tagged unions): each variant is one of several named forms, with optional payloads and optional explicit discriminants.

### Declaration

```go
enum SomeEnum {
	Value1
	Value2(String)
	Value3(int)
	Value4 = 3
}
```

- **Unit variant** — `Value1` carries no data.
- **Tuple variant** — `Value2(String)`, `Value3(int)` attach one or more payload types (tuple variants).
- **Explicit discriminant** — `Value4 = 3` assigns a fixed numeric tag (for C/interop or stable layout); variants without `=` get auto-incremented tags where applicable.

Struct-style variants (named fields) are also supported:

```go
enum Message {
	Quit
	Write { text string, bytes int }
	ChangeColor { r, g, b uint8 }
}
```

### Construction

When the expected type is known, **omit the enum name** and use the variant alone:

```go
var a SomeEnum = Value1
var b SomeEnum = Value2("hello")
c := Value3(42)              // type inferred from context
```

The qualified form is always valid:

```go
a := SomeEnum.Value1
b := SomeEnum.Value2("hello")
c := SomeEnum.Value3(42)
d := SomeEnum.Value4

m := Message.Write{ text: "hi", bytes: 5 }
```

### Switching on enums

Use a `**switch` statement** or `**switch` expression** to branch on the active variant and bind payloads. There is no `match` keyword.

**Exhaustiveness** — if a switch has no `default` case, the compiler requires every variant to be covered; missing any variant is a **compile-time error**. A `default` case is **allowed** and satisfies exhaustiveness (you may switch on a subset of variants and handle the rest in `default`).

Inside a `switch` on an enum, **case labels omit the enum type name** — write `case Value1:` not `case SomeEnum.Value1:`.

Switch statement:

```go
switch v {
case Value1:
	fmt.Println("value1")
case Value2(s):
	fmt.Println(s)
case Value3(n):
	fmt.Println(n)
case Value4:
	fmt.Println("value4")
}
```

Switch expression:

```go
n := switch v {
case Value1:
	0
case Value2(s):
	len(s)
case Value3(n):
	n
case Value4:
	3
}

desc := switch m {
case Quit:
	"quit"
case Write { text }:
	text
case ChangeColor { r, g, b }:
	int(r) + int(g) + int(b)
}
```

Invalid (compile error — missing `Value4` and no `default`):

```go
// switch v {
// case Value1:
// case Value2(s):
// case Value3(n):
// } // ERROR: switch on SomeEnum is not exhaustive
```

Valid ( `default` satisfies exhaustiveness):

```go
switch v {
case Value1:
	fmt.Println("value1")
default:
	fmt.Println("other")
}
```

### Methods and generics

Enums may have methods and type parameters:

```go
enum Option[T] {
	None
	Some(T)
}

func (o Option[int]) IsSome() bool {
	switch o {
	case Some(_):
		return true
	case None:
		return false
	}
}
```

### Notes

- Enums are distinct from `int?` nullable types and from `iota` constant groups.
- Variant names live in the enum’s namespace. Use unqualified names when the type is known (`var a SomeEnum = Value1`, `case Value2(s):`) or the qualified form (`SomeEnum.Value2`) anywhere.
- Memory layout is implementation-defined; explicit discriminants (`Value4 = 3`) document ABI intent.
- `enum` variants may appear in default arguments when the default is a compile-time constant variant (e.g. `mode Mode = Mode.Read`).
- When new variants are added to an enum, every non-`default` `switch` on that type must be updated or the build fails (exhaustiveness checking).

## Operator Overloading

User-defined types may implement built-in operators by declaring **operator methods**: ordinary functions whose names are operator tokens. At a use site, the compiler selects the operator method when at least one operand has a type that defines that operator (same rules as other overload resolution: count and types must match unambiguously).

Overload rules follow the same limits as **C#** where practical: only the operators in the tables below may be defined; **compound assignment** (`+=`, `-=`, `*=`, `&=`, …) is **not** overloadable—the compiler lowers `a += b` to `a = a + b` (and similarly for other compound forms) when the corresponding binary operator is defined. **`++` and `--`** may be overloaded (prefix and postfix use the same operator method; see below). **Comparison operators** must be declared in **pairs** (`==` with `!=`, `<` with `>`, `<=` with `>=`).

### `~` vs `^` (Go vs C#)

**C#** uses `~` for bitwise complement (`~x`). **Go does not use `~` as an expression operator.** In Go, `~` appears only in **type sets** (interface constraints), e.g. `interface { ~int | ~string }`.

For user-defined types, **bitwise complement** is overloaded with **unary `^`**, the same token as binary XOR:


| Language          | Bitwise complement                          | Bitwise XOR                                         |
| ----------------- | ------------------------------------------- | --------------------------------------------------- |
| C#                | `~a` → `operator ~`                         | `a ^ b` → `operator ^`                              |
| Go (this feature) | `^a` → `func ^(a T)` (unary, one parameter) | `a ^ b` → `func ^(l, r T)` (binary, two parameters) |


There is no `func ~(a T)` overload; `~` is not an overloadable operator name.

### Declaration syntax

**Binary** operators take two parameters (conventionally `l`, `r`):

```go
func +(l, r *Matrix) *Matrix { /* ... */ }
func ==(l, r *Matrix) bool   { /* ... */ }
func <<(l *Matrix, r int) *Matrix { /* ... */ }
```

**Unary** operators take one parameter:

```go
func +(a *Vector) *Vector { /* unary + */ }
func -(a *Vector) *Vector { /* unary - */ }
func !(a BitSet) bool      { /* logical not */ }
func ^(a BitSet) BitSet   { /* bitwise complement (C# ~) */ }
func ++(a *Counter) *Counter { /* prefix/postfix ++ */ }
func --(a *Counter) *Counter { /* prefix/postfix -- */ }
```

**Indexing** uses `[]` for read and `[]=` for write when both are needed:

```go
func [](m *Matrix, i, j int) float64       { /* m[i, j] */ }
func []=(m *Matrix, i, j int, v float64)  { /* m[i, j] = v */ }
```

**Compound assignment** (`+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `&^=`, `<<=`, `>>=`) is **not** overloadable (same as C#). The compiler rewrites them when the matching binary operator exists, e.g. `m += other` → `m = m + other` if `func +(l, r *Matrix) *Matrix` is defined and `m` is assignable.

Operator methods are declared at package scope (like ordinary functions), not as methods with a receiver. The “receiver” is always the left operand for binary ops and the sole operand for unary ops.

### Supported operators

The following tables list every overloadable operator. Tokens not listed under [Not overloadable](#not-overloadable) are reserved for future use and remain compile-time errors if used as an operator name.

#### Unary


| Operator     | Method name      | Meaning                                    |
| ------------ | ---------------- | ------------------------------------------ |
| `+a`         | `func +(a T) U`  | Unary plus                                 |
| `-a`         | `func -(a T) U`  | Unary minus / negation                     |
| `!a`         | `func !(a T) U`  | Logical NOT                                |
| `^a`         | `func ^(a T) U`  | Bitwise complement (C# `~`; not `~` in Go) |
| `++a`, `a++` | `func ++(a T) U` | Increment (C# `operator ++`)               |
| `--a`, `a--` | `func --(a T) U` | Decrement (C# `operator --`)               |


Unary `+` and `-` are distinguished from binary `+` and `-` by arity. **Unary `*` (indirection) is not overloadable**; only built-in pointer dereference applies. Binary `*` (multiplication) remains overloadable. `**++` / `--`:** one overload `func ++(a T) U` serves both prefix and postfix; postfix is lowered as `t := a; _ = ++a; t` (and similarly for `--`) unless the implementation documents a different lowering.

#### Binary — arithmetic


| Operator | Method name        | Compound form (not overloadable; synthesized) |
| -------- | ------------------ | --------------------------------------------- |
| `a + b`  | `func +(l, r T) U` | `a += b` → `a = a + b`                        |
| `a - b`  | `func -(l, r T) U` | `a -= b` → `a = a - b`                        |
| `a * b`  | `func *(l, r T) U` | `a *= b` → `a = a * b`                        |
| `a / b`  | `func /(l, r T) U` | `a /= b` → `a = a / b`                        |
| `a % b`  | `func %(l, r T) U` | `a %= b` → `a = a % b`                        |


#### Binary — bitwise and shifts


| Operator | Method name         | Compound form (synthesized) |
| -------- | ------------------- | --------------------------- |
| `a & b`  | `func &(l, r T) U`  | `a &= b` → `a = a & b`      |
| `a | b`  | `func |(l, r T) U`  | `a |= b` → `a = a | b`      |
| `a ^ b`  | `func ^(l, r T) U`  | `a ^= b` → `a = a ^ b`      |
| `a &^ b` | `func &^(l, r T) U` | `a &^= b` → `a = a &^ b`    |
| `a << b` | `func <<(l, r T) U` | `a <<= b` → `a = a << b`    |
| `a >> b` | `func >>(l, r T) U` | `a >>= b` → `a = a >> b`    |


For unary `^` vs binary `^`, the compiler picks the unary or binary overload from arity. Bitwise XOR and bitwise complement share the `^` token in source; the unary form is always `func ^(a T)` with one parameter, the binary form `func ^(l, r T)` with two.

#### Binary — comparison

Comparison operators return a boolean (or a type assignable to the context, e.g. a custom `Bool` type, if `==` is also overloaded for the result).

**Pairing rule (C#):** comparison overloads must be declared in **pairs**. If you define one member of a pair, you must define the other in the same package for the same type; otherwise it is a **compile-time error**.

| Pair | Operators | Method names |
|------|-----------|--------------|
| Equality | `==` and `!=` | `func ==(l, r T) U` and `func !=(l, r T) U` |
| Ordering | `<` and `>` | `func <(l, r T) U` and `func >(l, r T) U` |
| Ordering | `<=` and `>=` | `func <=(l, r T) U` and `func >=(l, r T) U` |

Examples:

- Defining only `func <` without `func >` → error.
- Defining `==` without `!=` → error.
- Defining `<`, `>`, `<=`, and `>=` together is valid (both ordering pairs complete).
- Defining no comparison operators at all is valid (built-in rules or other ops only).

There is no `==` compound assignment. You may omit **all** comparison overloads for a type; partial sets are not allowed.

#### Indexing and slicing


| Form                              | Method name                                                                                                          |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `a[i]` or `a[i, j]` (multi-index) | `func [](a T, indices ...IndexTypes) U`                                                                              |
| `a[i] = v`                        | `func []=(a T, indices ..., v V)`                                                                                    |
| `a[i:j]` / `a[i:j:k]`             | Optional: `func [:](a T, i, j, k ...) U` if slice syntax is overloaded for the type (otherwise use ordinary methods) |


Index expressions use the same arity as the `[]` operator method parameters (receiver plus indices).

### Resolution

- For `l op r`, if `l`’s type defines `func op(l, r …)`, that overload is used; else if `r`’s type defines it, that overload is used; if both define it, the left operand’s package-level operator wins unless one signature is an exact match and the other requires conversion (same ambiguity rules as function overloading).
- For unary `op a`, only `a`’s type is consulted.
- Built-in operators still apply when neither operand defines an overload and the operands are predeclared types (numbers, strings, channels, etc.).
- Operator methods must be exported if the type is used from other packages and the operator should apply there (`func +(l, r *Matrix)` with exported `Matrix`).

### Example

```go
type BitSet uint64

func +(a BitSet) BitSet  { return a }
func ^(a BitSet) BitSet  { return ^uint64(a) }
func &(l, r BitSet) BitSet { return BitSet(uint64(l) & uint64(r)) }
func |(l, r BitSet) BitSet { return BitSet(uint64(l) | uint64(r)) }
func <<(l BitSet, r uint) BitSet { return BitSet(uint64(l) << r) }
func ==(l, r BitSet) bool { return l == r }
func !=(l, r BitSet) bool { return l != r }

type Matrix struct { /* ... */ }

func +(l, r *Matrix) *Matrix { /* ... */ }
func -(l, r *Matrix) *Matrix { /* ... */ }
func *(l, r *Matrix) *Matrix { /* matrix multiply */ }
func -(m *Matrix) *Matrix    { /* unary negation */ }
func [](m *Matrix, i, j int) float64 { /* ... */ }
func []=(m *Matrix, i, j int, v float64) { /* ... */ }

type Counter int

func ++(c *Counter) *Counter { *c = *c + 1; return c }

func main() {
	var a, b BitSet
	_ = ^a          // bitwise complement (not ~)
	_ = a & b
	_ = a << 3

	m := &Matrix{}
	_ = -m
	_ = m[0, 0]
	m += m          // lowered to m = m + m; no func += overload

	var n Counter
	_ = ++&n
}
```

### Not overloadable (C#-aligned)

These cannot be declared as operator methods. Several match C# restrictions; compound assignment is listed here because only the **binary** operator is defined, not `+=` itself.

| Token / form | Reason |
|--------------|--------|
| `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `&^=`, `<<=`, `>>=` | **Not overloadable** (C#); synthesized from `+`, `-`, `*`, … |
| `&&`, `||` | Short-circuit boolean logic (C#) |
| `*` (unary) | Indirection / dereference; built-in pointers only (C# does not allow `operator*` unary either) |
| `~` (expression) | Not an expression operator in Go; used only in type sets; use unary `^` for complement |
| `<-` | Channel send/receive |
| `&` (unary) | Address-of |
| `:=`, `=` | Assignment, not operators |
| `...` | Variadic / slice unpacking |
| `?.`, `??`, `!.` | Nullable / result syntax (see above) |
| `=>` | Lambda syntax |


Built-in `++` and `--` on numeric types remain when no user overload exists.

### C# comparison (quick reference)


| C#                                 | Go (this feature)                                       |
| ---------------------------------- | ------------------------------------------------------- |
| `operator ~` (complement)          | `func ^(a T)` unary                                     |
| `operator ++` / `operator --`      | `func ++(a T)` / `func --(a T)`                         |
| `operator +=` etc.                 | **Not allowed**; use binary `+` and assignment lowering |
| `operator true` / `operator false` | Not supported (no custom boolean conversion operators)  |
| `operator &` (unary)               | Not supported (address-of stays built-in)               |
| `operator*` (unary)                | Not supported; use built-in `*p` on pointers            |
| `==` / `!=`, `<` / `>`, `<=` / `>=` | Must be overloaded in **pairs** (compile error if incomplete) |


### Notes

- Define only the operators that make sense for the type; undefined operators fall back to built-in rules or are compile errors if no built-in applies.
- **Comparison pairs** are enforced at declaration time: e.g. `func <` without `func >` for the same type is rejected.
- Prefer consistent signatures: binary arithmetic usually returns a new value or the same type; `[]=` mutates the container.
- For `a += b`, the compiler requires `a` to be assignable and a matching `func +(…)` (or other binary op); there is no separate `+=` overload to implement.
- `interface{}` / interface types do not get automatic operator lifting; each concrete type supplies its own operators.
- Operator overloading does not change evaluation order except where the language already guarantees it (e.g. `&&` / `||` remain non-overloadable).

