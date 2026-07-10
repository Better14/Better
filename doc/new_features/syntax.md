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

## Feedback

This feature is implemented in the compiler, type checker, and gopls. If you have thoughts on readability, tooling impact, or migration from the `type` form, please share feedback.
