# Better

Better is a fork of the [Go programming language](https://go.dev/) with language and standard-library extensions aimed at clearer, more expressive code. Full design notes live in `[doc/new_features/](doc/new_features/new_features.md)`.

---

## Why improve Go?

[Go is small, not simple](https://medium.com/@the_atomic_architect/go-isnt-simple-it-s-just-missing-features-and-that-s-costing-teams-millions-8d84cd9cc7a7). Error handling in Go is very verbose and makes code hard to read. Go is also missing tons of modern features that make code simpler — and in many cases easier to read, in our opinion.

Better keeps Go’s strengths (fully compiled, fast compile times, goroutines / no async, simple syntax, a strong stdlib, explicit style) while greatly improving the language.

---

## What have you improved?

### 1. No more `if err != nil { ... }` checks everywhere

### Result types (`T!`) and the `!` operator

`(T, error)` can be represented with `T!`. Early return errors with the `!` operator, like Rust's `?` operator.

```go
func readCount() int! {
	n := parse()!           // unwrap or early-return err
	return n
}

func readName() string! {
	return fetch()!.name    // propagate error, else access .name
}

err := possibleError()
err!                      // early-return if err != nil
```

Full spec: [doc/new_features/result_types.md](doc/new_features/result_types.md)

### 2. Stack traces everywhere

### Structured errors and stack traces

`errors.New`, `fmt.Errorf` (with one `%w`), `log.Fatal`, and most stdlib error types capture stack traces.

```go
err := errors.New("permission denied")
err = errors.New("open file: %s", path)

if err != nil {
	return err.Wrap("load config")
}

return fmt.Errorf("readFile %s: %w", path, err)

log.Printf("%+v", err)   // full chain + stacks
```

Custom error types embed `errors.Base` for message, stack trace, and wrapping:

```go
type AppError struct {
	errors.Base
}

func validate(id string) error {
	return errors.NewCustom[AppError]("invalid id: %s", id)
}

type NotFoundError struct {
	errors.Base
	Code int
}

func newNotFound(code int, msg string) *NotFoundError {
	var err NotFoundError
	errors.InitCustom(&err.Base, "%s", msg)
	err.Code = code
	return &err
}

log.Printf("%+v", newNotFound(404, "user missing"))
```

Full spec: [doc/new_features/errors.md](doc/new_features/errors.md)

### 3. Nullable types (`T?`)

Optional values with `?.` and `??` — distinct from `T!` (result / error).

```go
var count int? = 5

label := obj?.title ?? "untitled"
n := count ?? 0

if v := lookup(3); v != nil {
	use(v)
}
```

Full spec: [doc/new_features/nullable_types.md](doc/new_features/nullable_types.md)

### 4. Enums

Algebraic enums with exhaustive switching. No more flimsy `iota`.

```go
enum Message {
	Quit
	Write { text string, bytes int }
	ChangeColor { r, g, b uint8 }
}

desc := switch m {
case Quit:
	"quit"
case Write { text }:
	text
case ChangeColor { r, g, b }:
	fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
```

Full spec: [doc/new_features/enums.md](doc/new_features/enums.md)

### 5.1 LINQ

C#-style query methods on slices and `iter.Seq[T]`.

```go
import "linq"

nums := []int{1, 2, 3, 4, 5, 6}

firstEvenDouble := nums.Where(n => n % 2 == 0).Select(n => n * 2).First()
topThree := nums.Where(n => n > 2).OrderByDescending(n => n).Take(3).ToList()
sumOfSquares := nums.Select(n => n * n).Sum()
```

Full spec: [doc/new_features/linq.md](doc/new_features/linq.md)

### 5.2 Lambda syntax (`=>`)

Single expression-only functions with inferred parameter types. No multi-line functions.

```go
evens := nums.Where(n => n % 2 == 0)
```

Full spec: [doc/new_features/lambda_syntax.md](doc/new_features/lambda_syntax.md)

### 5.3 Extension methods

Methods on predeclared types, slices, generics, or types from other packages.

```go
func (i int) Square() int { return i * i }

func (p person.Person) Hello() string {
	return "Hi, " + p.Name
}

a := person.Person{Name: "Ada"}
a.Hello()
```

Full spec: [doc/new_features/extension_methods.md](doc/new_features/extension_methods.md)

### 5.4 Generic methods

Methods on generic types and methods with their own type parameters (Enhanced version of Go 1.27+ generic methods).

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

Full spec: [doc/new_features/generic_methods.md](doc/new_features/generic_methods.md)

### 6. Function and method overloading

Same name, different parameter types or arity.

```go
func connect(host string) { /* … */ }
func connect(host string, port int) { /* … */ }

connect("example.com")
connect("example.com", 8080)
```

Full spec: [doc/new_features/overloading.md](doc/new_features/overloading.md)

### 7. Default function arguments

Trailing parameters may have compile-time defaults.

```go
func log(msg string, level int = 1) {}

log("started")   // level defaults to 1
log("warn", 2)
```

Full spec: [doc/new_features/default_arguments.md](doc/new_features/default_arguments.md)

### 8. Operator overloading

User-defined operators via package-level operator methods. Good for custom data types, data science, game programming, etc.

```go
func +(l, r *Matrix) *Matrix { /* … */ }
func ==(l, r *Matrix) bool
func !=(l, r *Matrix) bool

m := a + b
if m == other { /* … */ }
```

Full spec: [doc/new_features/operator_overloading.md](doc/new_features/operator_overloading.md)

### 9. If and switch expressions

`if` and `switch` as expressions that evaluate to a value.

```go
a := if 5 < 6 { 1 } else { 2 } // The ternary operated adapted to Go.

label := switch x {
case 1:
	"one"
case 2:
	"two"
default:
	"other"
}
```

Full spec: [doc/new_features/expressions.md](doc/new_features/expressions.md)

### 10. Data structures

First-class `list`, `set`, `queue`, `stack`, heaps, and trees.

```go
nums := list.Of(1, 2, 3)
nums.Append(4)

tags := {}string{"go", "linq", "go"}   // set literal

q := queue.Of("a", "b")
q.Enqueue("c")
```

Full spec: [doc/new_features/data_structures.md](doc/new_features/data_structures.md)

### 11. Better time formatting

### Library changes

Better time formatting using .NET-style date/time formatting.

```go
t.FormatCustom("MM/dd/yyyy g")    // "06/15/2009 A.D."
t.FormatCustom("MMMM dd, yyyy")   // "June 15, 2009"
parsed, _ := time.ParseCustom("yyyy-MM-dd", "2009-06-15", time.UTC)
```

Full spec: [doc/new_features/library_changes.md](doc/new_features/library_changes.md)

### 12. Struct and interface shorthand (optional, feedback wanted)

Drop the leading `type` keyword for named struct and interface types. Shorter syntax that matches other languages.

```go
struct Person {
	Name string
	Age  int
}

interface Reader {
	Read(p []byte) (n int, err error)
}
```

Full spec: [doc/new_features/syntax.md](doc/new_features/syntax.md)

### 13. Nullable pointer types (proposed)

Compile-time null safety for pointers via `go.mod`.

```go
// go.mod
nullable_pointers enable   // *T vs *T?
```

Full spec: [doc/new_features/nullable_pointer_types.md](doc/new_features/nullable_pointer_types.md)

### 14. Compiler performance (compiler internals)

Type-checker indexes to keep compile times fast with Better features (operator overload resolution, extension methods, enums, and more).

Full spec: [doc/new_features/compiler_performance.md](doc/new_features/compiler_performance.md)

### 15. gopls (IDE support)

Language-server support for Better syntax — completion, diagnostics, and signature help for overloads and extensions.

Full spec: [doc/new_features/gopls.md](doc/new_features/gopls.md)

See also the [quick reference table](doc/new_features/new_features.md#quick-reference) in the feature index.

---

Unless otherwise noted, the Go source files are distributed under the BSD-style license found in the LICENSE file.

## Download and install

### Step 1: Install upstream Go (bootstrap)

Building Better requires a working **upstream Go toolchain** (Go **1.24.6 or later**). Download an official binary release from Google:

**[https://go.dev/dl/](https://go.dev/dl/)**

Follow the install guide for your platform: **[https://go.dev/doc/install](https://go.dev/doc/install)**

Verify the bootstrap toolchain:

```bash
go version
```

On Windows (PowerShell):

```powershell
go version
```

### Step 2: Build Better

Clone or copy this repository, then follow **[doc/new_docs/installation.md](doc/new_docs/installation.md)** for the full build process. A short summary is in [Build from source](#build-from-source-bootstrap) below.

After the build, add Better’s `bin` directory to your `PATH` and set `GOROOT` to the fork root.

---

## Build from source (bootstrap)

Better follows the upstream [Installing Go from source](https://go.dev/doc/install/source) process: a bootstrap Go compiler builds this tree, then the result becomes your new `GOROOT`.

**Full instructions:** [doc/new_docs/installation.md](doc/new_docs/installation.md)

You need **`GOROOT_BOOTSTRAP`** pointing at upstream Go (not Better) — typically the install from [go.dev/dl](https://go.dev/dl/). Set it explicitly if `go` on your `PATH` is missing or already points at Better.

| Platform | Build command | Notes |
| -------- | ------------- | ----- |
| **Linux** | `./make.bash` | Run from `$GOROOT/src`. Use `./all.bash` to build and run tests. |
| **macOS** | `./make.bash` | Same as Linux. Do not use `make.bash` on Windows. |
| **Windows** | `make.bat` | Run from `%GOROOT%\src`. Requires MinGW — see below. Use `all.bat` to build and test. |

Replace `/path/to/better/go` with the absolute path to this repository’s `go` directory.

### Linux and macOS

```bash
export GOROOT_BOOTSTRAP=$(go env GOROOT)
export GOEXPERIMENT=genericmethods   # optional; generic methods

cd /path/to/better/go/src
./make.bash          # build toolchain only
# ./all.bash         # build + run tests (long)

export GOROOT=/path/to/better/go
export PATH=$GOROOT/bin:$PATH
go version
```

Add the `export` lines to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make Better permanent.

### Windows

Windows builds need a C compiler on `PATH` (MinGW **`gcc`**) unless you set `CGO_ENABLED=0`. Upstream documents the full setup on the **[Go Wiki: WindowsBuild](https://go.dev/wiki/WindowsBuild)** page; see also [doc/new_docs/installation.md](doc/new_docs/installation.md).

**Install MinGW (summary):**

1. Download the MinGW installer from [SourceForge — mingw-get-inst](https://sourceforge.net/projects/mingw/files/OldFiles/mingw-get-inst/) ([WindowsBuild](https://go.dev/wiki/WindowsBuild)).
2. Enable **C Compiler**, **MinGW Developer Toolkit**, and **MSYS Basic System**.
3. Add `C:\MinGW\bin` (and MSYS bin if needed) to **`PATH`**.
4. Verify: `gcc --version`.

Install upstream Go from [go.dev/dl](https://go.dev/dl/), then:

```powershell
$env:GOROOT_BOOTSTRAP = (go env GOROOT)
$env:GOEXPERIMENT = "genericmethods"   # optional

cd C:\path\to\better\go\src
.\make.bat           # build toolchain only
# .\all.bat          # build + run tests (long)

$env:GOROOT = "C:\path\to\better\go"
$env:PATH = "$env:GOROOT\bin;$env:PATH"
go version
```

Add `GOROOT` and update `PATH` in System Environment Variables to keep Better across sessions.

### Bootstrap details

- **`GOROOT_BOOTSTRAP`** must contain `bin/go` (or `bin\go.exe` on Windows) from upstream Go ≥ 1.24.6.
- If unset, the scripts search common locations (`$HOME/go1.24.6`, `$HOME/sdk/go1.24.6`, etc.) and any other `go` on `PATH` whose `GOROOT` is not Better.
- **`make.bash` / `make.bat`** only build the toolchain. **`all.bash` / `all.bat`** also run the full test suite.
- For IDE support, build [gopls](doc/new_features/gopls.md) from the `go_tools` repository against Better’s `GOROOT`.

---

## Contributing

Go is the work of thousands of contributors. We appreciate your help!

To contribute upstream, read [https://go.dev/doc/contribute](https://go.dev/doc/contribute). For Better, see [`doc/new_features/`](doc/new_features/new_features.md) for the feature index and design docs.