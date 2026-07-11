# Struct and interface declarations

**Implemented.** Community feedback still welcome.

This document describes a shorthand for declaring named struct and interface types. The syntax drops the leading `type` keyword and places the type name between the keyword and the body.

## Struct declarations

Before:

```go
type MyStruct struct {
	Field1 int
	Field2 string
}
```

After:

```go
struct MyStruct {
	Field1 int
	Field2 string
}
```

The body is unchanged: fields, embedded types, and struct tags use the same rules as `type Name struct { … }`.

## Interface declarations

Before:

```go
type MyInterface interface {
	Method() error
}
```

After:

```go
interface MyInterface {
	Method() error
}
```

Embedded interfaces, method sets, and constraint syntax inside the body are unchanged.

## Compatibility

The existing `type` form remains valid and continues to work:

```go
type MyStruct struct { /* … */ }
type MyInterface interface { /* … */ }
```

Both forms declare the same named types and are interchangeable at the language level. New code may use either style; mixed use in one package is allowed.

## gofix

[`go fix`](https://pkg.go.dev/golang.org/x/tools/cmd/fix) includes the `shorthandtypes` modernizer, which rewrites the old `type` form to the new shorthand:

```go
// before gofix
type Person struct {
	Name string
}

type Stringer interface {
	String() string
}
```

```go
// after gofix
struct Person {
	Name string
}

interface Stringer {
	String() string
}
```

Running gofix on a package or file would rewrite only struct and interface type declarations that match the old pattern. Other `type` declarations (aliases, defined non-struct/interface types, type parameters, and so on) would be left unchanged.

The shorthand form is valid at package level and inside function bodies:

```go
func f() {
	struct subRow {
		ID string
	}
	var rows []subRow
}
```

After fixes are applied, `go fix` runs gofmt on each file so multi-line method chains use leading dots on continuation lines (`.Select`, `.Where`, and so on).

Packages under `GOROOT/src` (the Go toolchain and standard library tree) are never rewritten, so `go fix` can be run safely while developing the compiler itself.

## For-in loops

**Implemented.** Community feedback still welcome.

Iterate over slice, map, channel, and string elements with a Python-style `in` clause. The old `range` form remains valid.

### Value iteration

```go
for item in list {
	use(item)
}
```

This is equivalent to `for _, item := range list`. A single name before `in` always binds the **element/value**, not the index.

### Index and value

```go
for i, item in list {
	use(i, item)
}
```

Equivalent to `for i, item := range list`.

### Index-only loops

Index-only iteration uses a blank value with `in`:

```go
for i, _ in list {
	use(i)
}
```

Equivalent to `for i := range list`. There is no `for i in list` form for indices alone.

### Legacy syntax

```go
for _, item := range list { /* … */ }
```

still compiles and behaves identically. New code should prefer `for item in list`.

## gofix

[`go fix`](https://pkg.go.dev/golang.org/x/tools/cmd/fix) includes the `forin` modernizer, which rewrites range loops to for-in syntax:

```go
// before gofix
for _, item := range items {
	_ = item
}
for i, item := range items {
	_ = i
}
for i := range items {
	_ = i
}
```

```go
// after gofix
for item in items {
	_ = item
}
for i, item in items {
	_ = i
}
for i, _ in items {
	_ = i
}
```

Index-only loops (`for i := range items`) are rewritten to `for i, _ in items`. The `modernize` tool applies the same rewrite when `for_in_syntax` is enabled (default).

## Feedback

This feature is implemented in the compiler, type checker, and gopls. If you have thoughts on readability, tooling impact, or migration from the `type` form, please share feedback.

## Array, map, and set literals

**Implemented.** Community feedback still welcome.

Shorthand literals infer element types from their entries, similar to composite literals without an explicit type.

### Array literals

```go
a := ["string", "asdf"]
```

Equivalent to:

```go
a := []string{"string", "asdf"}
```

### Dict literals

```go
m := {"a": "b"}
```

Equivalent to:

```go
m := map[string]string{"a": "b"}
```

### Set literals

```go
s := {"a", "b", "c", "c"}
```

Equivalent to:

```go
s := {}string{"a", "b", "c", "c"}
```

which is equivalent to:

```go
s := set.Of("a", "b", "c", "c")
```

Typed set literals use `{}T{…}`:

```go
t := {}string{"x", "y"}
```

## Spread operator

**Implemented.** Community feedback still welcome.

### Variadic calls

Prefix spread is preferred; suffix spread remains valid.

```go
func myFunc(arg ...int) {}

a := [1, 2, 3]
myFunc(...a) // preferred
myFunc(a...) // still valid
```

### Literals

Spread works in array, map, and set literals:

```go
a := ["apple", "banana"]
b := ["fruit", ...a]

map1 := {"a": "b"}
map2 := {"c": "d", ...map1}

set1 := {"a"}
set2 := {"b", ...set1}
```

## gofix

[`go fix`](https://pkg.go.dev/golang.org/x/tools/cmd/fix) includes modernizers that rewrite the long forms above to shorthand literal and prefix-spread syntax. The `modernize` tool applies the same rewrites when `shorthand_literals` and `spread_call_syntax` are enabled (default).

## Negative slice indices

**Implemented.** Community feedback still welcome.

Slice expressions accept Python-style negative bounds. A negative index counts from the end of the sequence, so `-1` is the last element and `-2` is the second-to-last.

Before:

```go
// list has length 7
list[len(list)-2:]
```

After:

```go
list[-2:] // second-to-last element through the end
```

### Omitted bounds

Standard Go slice syntax allows omitting the low bound (defaults to 0), the high bound (defaults to length), or both:

```go
list[:5]  // first five elements — same as list[0:5]
list[3:]  // from index 3 through the end
list[:]   // full slice (copy of backing array for slices)
```

These forms remain valid and are not rewritten by the modernizer.

With negative indices, omitted low bounds work the same way:

```go
list[:-1]  // all but the last element — same as list[0:-1]
```

Both bounds may be negative:

```go
list[:-2]   // all but the last two elements
list[-3:-1] // third-to-last through second-to-last
```

## gofix

[`go fix`](https://pkg.go.dev/golang.org/x/tools/cmd/fix) includes the `negativeslice` modernizer, which rewrites `len(x)-n` slice bounds to `-n` and `len(x)` high bounds to an omitted end. The `modernize` tool applies the same rewrite when `negative_slice_indices` is enabled (default).
