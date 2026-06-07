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

See [Function and method overloading](overloading.md) for general overload rules. An overload set must be unambiguous: if more than one overload matches a call equally well, it is a **compile error** (`ambiguous overloaded call`).

#### Default arguments and ambiguous overloads

Defaults let a caller omit trailing parameters, so an overload with optional parameters can match the **same argument count** as another overload. When that happens, the call is ambiguous.

Invalid (ambiguous overload set — one-argument calls do not resolve uniquely):

```go
func myFunc(a int, b int = 5) {
}

func myFunc(a int) {
}
```

For `myFunc(5)`, both overloads match:

- `func myFunc(a int)` — `a = 5`
- `func myFunc(a int, b int = 5)` — `a = 5`, `b` uses its default

That call is a **compile error** (`ambiguous overloaded call`). Two-argument calls are fine: `myFunc(5, 10)` resolves only to the two-parameter overload.

The same rule applies to methods and to overloads that differ only in how many trailing parameters have defaults. Overloads must not overlap in arity once defaults are applied at the call site.

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

Use an `enum` type for mode defaults (see [Enums](enums.md)). Pointer parameters use `nil` where a nullable default is intended.

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

#### Ambiguous overloads with default arguments (compile-time error)

See [Default arguments and ambiguous overloads](#default-arguments-and-ambiguous-overloads) above. Overloads whose optional parameters overlap another overload’s arity are rejected at ambiguous call sites:

```go
// func myFunc(a int, b int = 5) {}
// func myFunc(a int) {}
// myFunc(5) // ERROR: ambiguous overloaded call
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
- Not valid in upstream Go.
