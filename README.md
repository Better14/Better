# BetterGo — a better version of Go

BetterGo is a fork of the [Go programming language](https://go.dev/) with language and standard-library extensions aimed at clearer, more expressive code. Full design notes live in [`doc/new_features/`](doc/new_features/new_features.md).

![Gopher image](https://golang.org/doc/gopher/fiveyears.jpg)
*Gopher image by [Renee French][rf], licensed under [Creative Commons 4.0 Attribution license][cc4-by].*

Unless otherwise noted, the Go source files are distributed under the BSD-style license found in the LICENSE file.

---

## Why improve Go?

[Go is small, not simple](https://medium.com/@the_atomic_architect/go-isnt-simple-it-s-just-missing-features-and-that-s-costing-teams-millions-8d84cd9cc7a7). Error handling in Go is very verbose and makes code hard to read. Go is also missing tons of modern features that make code simpler — and in many cases easier to read, in our opinion.

BetterGo keeps Go’s strengths (fast compile times, goroutines, a strong stdlib, explicit style) while adding the pieces teams often reach for third-party libraries or code generators to get.

---

## What have you improved?

### Type system

- **[Result types (`T!`)](doc/new_features/result_types.md)** — `(T, error)` shorthand and `!.value` / `!.field` error propagation
- **[Structured errors](doc/new_features/errors.md)** — stack traces on `errors.New`, `fmt.Errorf`, and most stdlib error types
- **[Nullable types (`T?`)](doc/new_features/nullable_types.md)** — optional values, `?.`, and `??`
- **[Nullable pointer types (`*T` / `*T?`)](doc/new_features/nullable_pointer_types.md)** — compile-time null safety for pointers (proposed)
- **[Enums](doc/new_features/enums.md)** — algebraic enums, variants, and exhaustive switching
- **[Operator overloading](doc/new_features/operator_overloading.md)** — user-defined operators via operator methods

### Functions and syntax

- **[Function and method overloading](doc/new_features/overloading.md)**
- **[Default function arguments](doc/new_features/default_arguments.md)**
- **[If and switch expressions](doc/new_features/expressions.md)**
- **[Lambda syntax (`=>`)](doc/new_features/lambda_syntax.md)**
- **[Struct and interface shorthand](doc/new_features/syntax.md)** — `struct T { … }`, `interface I { … }`

### Generics and methods

- **[Generic methods](doc/new_features/generic_methods.md)** (Go 1.27+)
- **[Extension methods](doc/new_features/extension_methods.md)** — methods on foreign and predeclared types

### Standard library

- **[Built-in LINQ](doc/new_features/linq.md)** — C#-style `Where`, `Select`, `OrderBy`, and more on slices and `iter.Seq[T]`
- **[Data structures](doc/new_features/data_structures.md)** — `list`, `set`, `queue`, `stack`, heaps, trees
- **[Library changes](doc/new_features/library_changes.md)** — e.g. .NET-style `time.FormatCustom` / `ParseCustom`

### Tooling and performance

- **[gopls (IDE support)](doc/new_features/gopls.md)** — language-server support for fork features
- **[Compiler performance](doc/new_features/compiler_performance.md)** — indexes to keep type-checking fast with new features

Syntax examples for each area follow below. See the [quick reference](doc/new_features/new_features.md#quick-reference) for a one-page cheat sheet.

---

## Table of contents

- [Why improve Go?](#why-improve-go)
- [What have you improved?](#what-have-you-improved)

### Install and build

- [Download and install](#download-and-install)
- [Build from source (bootstrap)](#build-from-source-bootstrap)

### Language features

- [Result types (`T!`) and the `!` operator](#result-types-t-and-the--operator) — [doc/new_features/result_types.md](doc/new_features/result_types.md)
- [Structured errors and stack traces](#structured-errors-and-stack-traces) — [doc/new_features/errors.md](doc/new_features/errors.md)
- [Nullable types (`T?`)](#nullable-types-t) — [doc/new_features/nullable_types.md](doc/new_features/nullable_types.md)
- [Enums](#enums) — [doc/new_features/enums.md](doc/new_features/enums.md)
- [LINQ, extension methods, and generic methods](#linq-extension-methods-and-generic-methods) — [doc/new_features/linq.md](doc/new_features/linq.md), [extension_methods.md](doc/new_features/extension_methods.md), [generic_methods.md](doc/new_features/generic_methods.md)
- [Function and method overloading](#function-and-method-overloading) — [doc/new_features/overloading.md](doc/new_features/overloading.md)
- [Operator overloading](#operator-overloading) — [doc/new_features/operator_overloading.md](doc/new_features/operator_overloading.md)
- [More language features](#more-language-features)

### Standard library and tooling

- [Data structures](doc/new_features/data_structures.md) — `list`, `set`, `queue`, `stack`, heaps, trees
- [Library changes](doc/new_features/library_changes.md) — e.g. .NET-style `time.FormatCustom`
- [Nullable pointer types](doc/new_features/nullable_pointer_types.md) — proposed `*T` / `*T?` null safety
- [Struct and interface shorthand](doc/new_features/syntax.md) — `struct T { … }`, `interface I { … }`
- [Compiler performance](doc/new_features/compiler_performance.md)
- [gopls (IDE support)](doc/new_features/gopls.md)

See also the [quick reference table](doc/new_features/new_features.md#quick-reference) in the feature index.

---

## Download and install

### Step 1: Install upstream Go (bootstrap)

Building BetterGo requires a working **upstream Go toolchain** (Go **1.24.6 or later**). Download an official binary release from Google:

**https://go.dev/dl/**

Follow the install guide for your platform: **https://go.dev/doc/install**

Verify the bootstrap toolchain:

```bash
go version
```

On Windows (PowerShell):

```powershell
go version
```

### Step 2: Build BetterGo

Clone or copy this repository, then follow the [bootstrap build instructions](#build-from-source-bootstrap) below. After the build, add BetterGo’s `bin` directory to your `PATH` and set `GOROOT` to the fork root.

---

## Build from source (bootstrap)

You need a bootstrap Go tree (`GOROOT_BOOTSTRAP`) that is **not** BetterGo — typically the upstream install from [go.dev/dl](https://go.dev/dl/). The build scripts compile BetterGo using that bootstrap compiler, then reinstall the toolchain into this tree.

Set `GOROOT_BOOTSTRAP` explicitly if the bootstrap Go is not on your `PATH` or not in the default search locations.

| Platform | Build command | Notes |
| -------- | ------------- | ----- |
| **Linux** | `./make.bash` | Run from `$GOROOT/src`. Use `./all.bash` to build and run tests. |
| **macOS** | `./make.bash` | Same as Linux. Do not use `make.bash` on Windows. |
| **Windows** | `make.bat` | Run from `%GOROOT%\src` in **cmd** or PowerShell. Use `all.bat` to build and test. |

Replace `/path/to/fork/go` with the absolute path to this repository’s `go` directory.

### Linux and macOS

```bash
# Point bootstrap at upstream Go (adjust path if needed)
export GOROOT_BOOTSTRAP=$(go env GOROOT)

# Optional: enable generic methods experiment if your bootstrap requires it
export GOEXPERIMENT=genericmethods

cd /path/to/fork/go/src
./make.bash          # build toolchain only
# ./all.bash         # build + run tests (long)
```

After a successful build:

```bash
export GOROOT=/path/to/fork/go
export PATH=$GOROOT/bin:$PATH
go version
```

Add the `export` lines to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make the fork permanent.

### Windows

Install upstream Go from [go.dev/dl](https://go.dev/dl/) first. Then:

```powershell
# Bootstrap: upstream Go (adjust if go.exe is elsewhere)
$env:GOROOT_BOOTSTRAP = (go env GOROOT)

# Optional
$env:GOEXPERIMENT = "genericmethods"

cd C:\path\to\fork\go\src
.\make.bat           # build toolchain only
# .\all.bat          # build + run tests (long)
```

After a successful build:

```powershell
$env:GOROOT = "C:\path\to\fork\go"
$env:PATH = "$env:GOROOT\bin;$env:PATH"
go version
```

Add `GOROOT` and update `PATH` in System Environment Variables if you want the fork available in every session.

### Bootstrap details

- **`GOROOT_BOOTSTRAP`** must contain `bin/go` (or `bin\go.exe` on Windows) from upstream Go ≥ 1.24.6.
- If unset, the scripts search common locations (`$HOME/go1.24.6`, `$HOME/sdk/go1.24.6`, etc.) and any other `go` on `PATH` whose `GOROOT` is not BetterGo.
- **`make.bash` / `make.bat`** only build the toolchain. **`all.bash` / `all.bat`** also run the full test suite.
- For IDE support, build [gopls](doc/new_features/gopls.md) from the `go_tools` repository against BetterGo’s `GOROOT`.

---

## Result types (`T!`) and the `!` operator

`(T, error)` can be written as **`T!`**. The postfix **`!`** operator unwraps the success value or early-returns the error from the current function.

```go
func readCount() int! {
	n := parse()!           // unwrap (T, error) or early-return err
	return n
}

func readName() string! {
	return fetch()!.name    // propagate error, then access .name
}

func setup(db *sql.DB) int! {
	db.Exec("CREATE TABLE t (id INT)")!  // check err, discard result
	return 0
}

func doThing() int! {
	err := possibleError()
	err!                    // early-return if err != nil
	return run()
}
```

Function signatures use the shorthand:

```go
func myFunc() int! {      // same as func myFunc() (int, error)
	a := myFunc2()!
	return a
}
```

Details: [doc/new_features/result_types.md](doc/new_features/result_types.md)

---

## Structured errors and stack traces

The `errors` package provides a structured **`errors.Error`** type with message, stack trace, and inner error chain. **`errors.New`** and **`fmt.Errorf`** (with zero or one `%w`) capture a stack trace at the call site. Most **built-in custom error types** in the standard library now embed `errors.Error` and record a trace when constructed.

```go
import "errors"
import "fmt"

// Root error with stack trace
err := errors.New("permission denied")

// Formatted root error
err = errors.New("open file: %s", path)

// Wrap adds a new layer with its own stack frame
if err != nil {
	return err.Wrap("load config")
}

// fmt.Errorf with one %w also returns *errors.Error with a trace
if err != nil {
	return fmt.Errorf("readFile %s: %w", path, err)
}

// Full chain + stacks for logging
log.Printf("%+v", err)   // or err.String() when type is *errors.Error
```

Inspect traces from stdlib errors with `errors.As`:

```go
var pe *fs.PathError
if errors.As(err, &pe) {
	log.Printf("%+v", &pe.Error)
}
```

Details: [doc/new_features/errors.md](doc/new_features/errors.md)

---

## Nullable types (`T?`)

**`T?`** means an optional value — `T` or `nil` — without an error channel. Use **`?.`** for null-conditional access and **`??`** for null-coalescing defaults.

```go
var a int? = 5
var b int? = nil

func lookup(id int) int? {
	if id < 0 {
		return nil
	}
	return id
}

// Null-conditional and coalescing
label := obj?.title ?? "untitled"
n := count ?? 0

// Unwrap after nil check
if v := lookup(3); v != nil {
	use(v)              // v is int in this branch
}
```

Do not confuse **`int?`** (nullable) with **`int!`** (result / error). Details: [doc/new_features/nullable_types.md](doc/new_features/nullable_types.md)

---

## Enums

Rust-style **algebraic enums** with unit, tuple, and struct variants. Switches can be exhaustive; missing variants without a `default` case is a compile error.

```go
enum Message {
	Quit
	Write { text string, bytes int }
	ChangeColor { r, g, b uint8 }
}

m := Write{ text: "hi", bytes: 5 }

desc := switch m {
case Quit:
	"quit"
case Write { text }:
	text
case ChangeColor { r, g, b }:
	fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

enum Option[T] {
	None
	Some(T)
}
```

Details: [doc/new_features/enums.md](doc/new_features/enums.md)

---

## LINQ, extension methods, and generic methods

### LINQ

Import **`linq`** for C#-style query methods on slices and `iter.Seq[T]`. Chains are lazy until a terminal operator runs.

```go
import "linq"

nums := []int{1, 2, 3, 4, 5, 6}

firstEvenDouble := nums.Where(n => n%2 == 0).Select(n => n * 2).First()
topThree := nums.Where(n => n > 2).OrderByDescending(n => n).Take(3).ToList()
sumOfSquares := nums.Select(n => n * n).Sum()
```

Details: [doc/new_features/linq.md](doc/new_features/linq.md)

### Extension methods

Attach methods to predeclared types, slices, generics, or types defined in other packages — using ordinary `func` + receiver syntax.

```go
func (i int) Square() int { return i * i }

func (p person.Person) Hello() string {
	return "Hi, " + p.Name
}

a := person.Person{Name: "Ada"}
a.Hello()   // method syntax, not personext.Hello(a)
```

Details: [doc/new_features/extension_methods.md](doc/new_features/extension_methods.md)

### Generic methods (Go 1.27+)

Methods on generic types and methods with their own type parameters:

```go
type List[E any] []E

func (l List[E]) Select[F any](f func(E) F) List[F] {
	r := make(List[F], len(l))
	for i, x := range l {
		r[i] = f(x)
	}
	return r
}

type Pair[A, B any] struct{ a A; b B }

func (p Pair[A, B]) Swap() Pair[B, A] {
	return Pair[B, A]{a: p.b, b: p.a}
}
```

Details: [doc/new_features/generic_methods.md](doc/new_features/generic_methods.md)

---

## Function and method overloading

Multiple functions or methods may share a name when their parameter lists differ by type and/or arity.

```go
func connect(host string) { /* … */ }
func connect(host string, port int) { /* … */ }
func connect(host string, port int, timeout time.Duration) { /* … */ }

connect("example.com")                    // first overload
connect("example.com", 8080)              // second
connect("example.com", 8080, time.Second) // third
```

Default arguments work on trailing parameters:

```go
func log(msg string, level int = 1) {}

log("started")       // level defaults to 1
log("warn", 2)
```

Details: [doc/new_features/overloading.md](doc/new_features/overloading.md), [default_arguments.md](doc/new_features/default_arguments.md)

---

## Operator overloading

User-defined types implement operators via **operator methods** — package-level functions named after the operator token.

```go
func +(l, r *Matrix) *Matrix { /* … */ }
func -(a *Vector) *Vector     { /* unary negation */ }
func ==(l, r *Matrix) bool
func !=(l, r *Matrix) bool    // comparison pairs required

func [](m *Matrix, i, j int) float64
func []=(m *Matrix, i, j int, v float64)

m := a + b
if m == other { /* … */ }
v := m[0, 1]
m[0, 1] = 3.14
```

Compound assignment (`+=`, `*=`, …) is lowered to `a = a + b` when the binary operator is defined. Details: [doc/new_features/operator_overloading.md](doc/new_features/operator_overloading.md)

---

## More language features

### If and switch expressions

```go
a := if 5 < 6 { 1 } else { 2 }

label := switch x {
case 1:
	"one"
case 2:
	"two"
default:
	"other"
}
```

Details: [doc/new_features/expressions.md](doc/new_features/expressions.md)

### Lambda syntax (`=>`)

Single-expression functions with inferred parameter types:

```go
sum := (a, b, c) => a + b + c
evens := nums.Where(n => n%2 == 0)
```

Details: [doc/new_features/lambda_syntax.md](doc/new_features/lambda_syntax.md)

### Data structures

First-class containers with literal syntax where noted:

```go
nums := list.Of(1, 2, 3)
nums.Append(4)

tags := {}string{"go", "linq", "go"}   // set literal; duplicates dropped

q := queue.Of("a", "b")
q.Enqueue("c")
```

Details: [doc/new_features/data_structures.md](doc/new_features/data_structures.md)

### Struct and interface shorthand

```go
struct Person {
	Name string
	Age  int
}

interface Reader {
	Read(p []byte) (n int, err error)
}
```

Details: [doc/new_features/syntax.md](doc/new_features/syntax.md)

### Library changes

.NET-style custom date/time format strings:

```go
t.FormatCustom("MM/dd/yyyy g")   // e.g. "06/15/2009 A.D."
t.FormatCustom("MMMM dd, yyyy")  // e.g. "June 15, 2009"
parsed, _ := time.ParseCustom("yyyy-MM-dd", "2009-06-15", time.UTC)
```

Details: [doc/new_features/library_changes.md](doc/new_features/library_changes.md)

### Nullable pointer types (proposed)

Optional compile-time null safety for pointers via `go.mod`:

```go
nullable_pointers enable   // *T vs *T?
```

Details: [doc/new_features/nullable_pointer_types.md](doc/new_features/nullable_pointer_types.md)

---

## Contributing

Go is the work of thousands of contributors. We appreciate your help!

To contribute upstream, read https://go.dev/doc/contribute. For BetterGo, see [`doc/new_features/`](doc/new_features/new_features.md) for the feature index and design docs.

[rf]: https://reneefrench.blogspot.com/
[cc4-by]: https://creativecommons.org/licenses/by/4.0/
