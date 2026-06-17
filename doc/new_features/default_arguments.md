# Default Function Arguments

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

See [Function and method overloading](overloading.md) for general overload rules.

#### Default arguments and ambiguous overloads

Defaults let a caller omit trailing parameters, so an overload with optional parameters can match the **same argument count** as another overload. Such an overload set is **ill-formed** and must be rejected at compile time.

**Rule:** When type-checking an overload set, the compiler considers every argument count and type tuple that any overload could accept (counting omitted trailing parameters filled from defaults). If more than one overload matches the same tuple equally well, the set is invalid. The compiler reports **`ambiguous overload set for f`** on the conflicting declaration(s). No call site is required; the declarations themselves do not compile.

Invalid (**compile error** — ambiguous overload set):

```go
func myFunc(a int, b int = 5) {
} // ERROR: ambiguous overload set for myFunc

func myFunc(a int) {
} // ERROR: ambiguous overload set for myFunc (conflicts with myFunc(int, int = 5))
```

Why: a one-argument call `myFunc(5)` would match both overloads equally:

- `func myFunc(a int)` — `a = 5`
- `func myFunc(a int, b int = 5)` — `a = 5`, `b` uses its default

Because that call would be ambiguous, the overload set is rejected when the package is type-checked. If a call site were reached anyway (for example in incomplete code), it would also be a **compile error** (`ambiguous overloaded call`).

The same rule applies to methods and to any overloads whose effective arities overlap once defaults are applied.

Valid (no arity overlap after defaults):

```go
func myFunc(a int) {}

func myFunc(a int, b int) {} // second argument required; no default on b

myFunc(5)    // func myFunc(a int)
myFunc(5, 6) // func myFunc(a int, b int)
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

Use an `enum` type for mode defaults (see [Enums](enums.md)). When the parameter type is known, omit the enum name and write the unit variant alone (`mode Mode = Read`). The qualified form (`Mode.Read`) is also valid. Pointer parameters use `nil` where a nullable default is intended.

```go
enum Mode {
	Read
	Write
	Both
}

func Open(path string, mode Mode = Read) {
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

#### Ambiguous overloads with default arguments (compile-time error)

See [Default arguments and ambiguous overloads](#default-arguments-and-ambiguous-overloads) above. The overload set below does not compile:

```go
// func myFunc(a int, b int = 5) {} // ERROR: ambiguous overload set for myFunc
// func myFunc(a int) {}            // ERROR: ambiguous overload set for myFunc
// myFunc(5)                        // would also ERROR: ambiguous overloaded call
```

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
- Overload sets that become ambiguous because of default arguments are **compile errors at declaration time** (`ambiguous overload set for f`); see [Default arguments and ambiguous overloads](#default-arguments-and-ambiguous-overloads).
- Not valid in upstream Go.
