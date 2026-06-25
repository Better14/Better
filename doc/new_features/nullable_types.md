# Nullable Types (`T?`)

`T?` on a value type `T` means **either a `T` or `nil`** — an optional value with no error channel. This is separate from [`T!`](result_types.md), which means **value or `error`**.


| Syntax | Meaning                                    |
| ------ | ------------------------------------------ |
| `int?` | `int` or `nil` (nullable)                  |
| `int!` | `int` or `error` (result / `(int, error)`) |


Nullable types are useful for primitives and structs that cannot otherwise hold `nil` in Go. Reference types (`*T`, `map`, `slice`, `chan`, `func`, `interface`) are already “nullable” via `nil`; `T?` is most important for `int`, `bool`, `float64`, struct types, etc.

## Suffix binding with composite types

`?` is a **postfix suffix on the entire type expression** to its left (the same parsing rule as `!`). It does not bind only to the innermost type name.

| Written | Meaning |
| ------- | ------- |
| `int?` | `int` or `nil` |
| `[]int?` | `([]int)?` — the **slice itself** may be `nil` (nullable slice) |
| `[](int?)` | `[]` of `int?` — each element is `int` or `nil` |
| `*int?` | `(*int)?` — nullable pointer to `int` (not the same as `*int`, which is already nil-able) |

So `[]int?` means **slice or null** (the whole slice is optional), not “slice of int-or-null elements”. For per-element optionals, use parentheses:

```go
var counts []int?      // counts is nil, or a []int
var items [](int?)     // slice where each element is int?
var name string? = row?.title
```

Slices, maps, channels, pointers, and interfaces are already nil-able in ordinary Go. `[]T?` is mainly useful when you need to distinguish “missing collection” (`nil` slice) from “empty collection” (`[]T{}`) at the type level. `[](T?)` is for when **each element** may be absent.

See [Suffix binding with composite types](result_types.md#suffix-binding-with-composite-types) for the same rules with `!` (`[]BookRow!` → `([]BookRow, error)`).

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

Conceptually, `int?` is an optional value (value + “has value” flag). The source-level model is **value or `nil`**, not value + `error`.

### Compiler lowering

At compile time, `T?` is lowered to a struct with the same layout as `Option[T]`:

```go
type Option[T any] struct {
	hasValue bool
	value  T
}
```

For example, `int?` becomes:

```go
struct {
	hasValue bool
	value  int
}
```

| Source              | Lowered meaning                                      |
| ------------------- | ---------------------------------------------------- |
| `nil`               | `{hasValue: false, value: zero}`                     |
| `5`                 | `{hasValue: true, value: 5}`                       |
| `v == nil`          | `!v.hasValue`                                        |
| `v != nil`          | `v.hasValue` (then `v` narrows to `T` in type check) |
| `int(v)`            | read `v.value`; panic if `!v.hasValue`               |
| `v ?? fallback`     | if `!v.hasValue` then `fallback` else `v.value`      |
| `*T` assignable to `T?` | `nil` pointer → nil optional; non-nil → `{true, *p}` |

Field names `hasValue` and `value` are fixed; the compiler generates anonymous structs with this shape (there is no user-visible `Option` type name in object code unless you define one yourself).

See [Compiler lowering](result_types.md#compiler-lowering) for the analogous `T!` → `Result[T]` representation.

### Assignability: `T?` is not `T`

`T?` and `T` are distinct types. A nullable value cannot be passed or assigned where a plain `T` is required without an explicit unwrap.

```go
func myPrint(a int) {
	fmt.Println(a)
}

func main() {
	var a int?
	myPrint(a) // compile error: int? is not assignable to int
}
```

To call `myPrint`, unwrap `a` in one of these ways:

**Nil check** — only call when a value is present:

```go
func main() {
	var a int?
	if a != nil {
		myPrint(a) // a is int in this branch
	}
}
```

**Force cast** — unwrap with `T(expr)`; panics if `expr` is `nil`:

```go
func main() {
	var a int?
	myPrint(int(a)) // panics if a == nil
}
```

**Null coalescing** — supply a default when `a` is `nil`:

```go
func main() {
	var a int?
	myPrint(a ?? 0) // int? ?? int → int
}
```

The same rules apply to assignment, return values, and other contexts that expect `T` rather than `T?`.

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

`??` applies to nullable types and other nil-able values (`*T`, maps, slices, pointers). For [`T!`](result_types.md) result values, `??` uses the success value when `err == nil` and the fallback when `err != nil`; see [Assignability: `T!` is not `T`](result_types.md#assignability-t-is-not-t).


| Form               | Role                                                  |
| ------------------ | ----------------------------------------------------- |
| `int?`             | Type: `int` or `nil`                                  |
| `expr?.field`      | Null-conditional access                               |
| `expr ?? fallback` | Null-coalescing: `expr` if non-`nil`, else `fallback` |


### Notes

- Do not confuse `int?` (nullable type) with `int!` (result type). They use different suffixes on purpose.
- Do not confuse `?.` (null-conditional) or `??` (null-coalescing) with `!.` (error propagation on `T!`).
- `T?` does not support `!.value` error propagation; that syntax applies only to `T!` / `(T, error)`.
- Nullable defaults in function parameters (e.g. `x int? = nil`) follow the same compile-time constant rules as other [default arguments](default_arguments.md) when/if defaults are added for nullable parameters.
- For pointer nullability (`*T` vs `*T?`), see [Nullable pointer types](nullable_pointer_types.md) (proposed; separate from value-type `T?`).
