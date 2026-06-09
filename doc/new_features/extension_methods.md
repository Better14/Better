# Extension Methods

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

Method-local type parameters appear after the name, as for [generic methods](generic_methods.md):

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
- **LINQ:** implemented as extensions on `[]T` in `import "linq"` and receiver methods on `linq.Lazy[T]`. See [Built-in LINQ](linq.md).

### Limitations

- **Ambiguity:** two imported extension packages define the same method on the same receiver type → error.
- **Instance methods win** over extensions.
- **Import required:** the calling file must import each extension package it relies on (normal import, not dot import).
- **Not in upstream Go.**
