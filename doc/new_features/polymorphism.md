# Polymorphism

Go’s core model is **structural interface polymorphism**: a value is usable wherever its method set satisfies an interface. This fork adds **generics**, **overloading**, **extension methods**, **operator methods**, and **enums** — each is a different form of polymorphism with distinct resolution rules.

Use this page as a map. Each mechanism has its own document with syntax, limits, and examples.

## Forms of polymorphism

| Kind | What varies | How the compiler picks an implementation | Document |
| ---- | ----------- | ---------------------------------------- | -------- |
| **Interface (subtype)** | Dynamic type behind an interface value | Method set must satisfy the interface; `errors.As` / type assertions recover concrete types | (standard Go) |
| **Parametric (generics)** | Type arguments (`T`, `K`, `V`, …) | Monomorphization / instantiation at compile time | [generic_methods.md](generic_methods.md) |
| **Ad-hoc — overloads** | Parameter count and types | Overload resolution at the call site | [overloading.md](overloading.md) |
| **Ad-hoc — operators** | Operand types for built-in operators | Operator method lookup by arity and types | [operator_overloading.md](operator_overloading.md) |
| **Extension dispatch** | Receiver type (including foreign and predeclared types) | Method-call syntax desugars to the matching extension in scope | [extension_methods.md](extension_methods.md) |
| **Sum types (enums)** | Active variant in a fixed variant set | `switch` / `switch` expression; exhaustiveness checking | [enums.md](enums.md) |

## Interface polymorphism (standard Go)

An interface value holds a **dynamic type** and a **dynamic value**. Any concrete type whose method set satisfies the interface can be assigned without inheritance or registration:

```go
type Writer interface {
	Write([]byte) (int, error)
}

func log(w Writer, msg []byte) {
	w.Write(msg)
}
```

**Type assertions** and **`errors.As`** recover a concrete type from an interface when needed:

```go
var err error = errors.New("fail")
var e *errors.Error
if errors.As(err, &e) {
	fmt.Println(e.StackTrace)
}
```

Interfaces remain the primary tool for **open** polymorphism — third-party types can satisfy your interface without modifying their definitions. See [Structured errors](errors.md) for embedding `errors.Error` in custom error types while still returning `error`.

## Parametric polymorphism (generics)

**Type parameters** abstract over types at compile time. One function or type definition works for many types; the compiler generates (or reuses) specialized code per instantiation:

```go
func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
```

**Generic methods** attach behavior to generic defined types; **generic methods with extra type parameters** (e.g. `Select[F any]` on `List[E]`) mirror C# LINQ projections. See [Generic methods](generic_methods.md).

**When to prefer generics over interfaces**

- Same algorithm for many types, no shared method set → generics.
- Behavior varies by type and callers pass different implementations → interfaces (or function parameters).
- Both compose: `func Sort[S ~[]E, E cmp.Ordered]` `(s S)` uses constraints instead of `interface{}`.

## Ad-hoc polymorphism — overloading

**Function and method overloading** selects among several declarations with the **same name** by **argument count and types**:

```go
func Parse(s string) int { /* … */ }
func Parse(s string, base int) int { /* … */ }

Parse("10")      // Parse(string)
Parse("ff", 16)  // Parse(string, int)
```

Resolution is **static** (compile time). It does not use dynamic dispatch. Overload sets must be unambiguous; see [Function and method overloading](overloading.md) and [Default arguments](default_arguments.md) for interaction with default parameters.

## Ad-hoc polymorphism — operator overloading

**Operator methods** let user-defined types participate in built-in operator syntax (`+`, `==`, `[]`, …):

```go
func +(l, r Vector) Vector { /* … */ }
func ==(l, r Vector) bool  { /* … */ }

var a, b Vector
c := a + b
ok := a == b
```

Operator names are ordinary function names at package scope; the compiler rewrites `a + b` to `+(a, b)` when a matching overload exists. Compound assignment (`+=`) is lowered through the binary operator. See [Operator overloading](operator_overloading.md).

## Extension methods

**Extension methods** add call-site method syntax for types you do not own — predeclared types, slices, foreign structs, `iter.Seq[T]`, etc.:

```go
func (i int) Square() int { return i * i }

func (p person.Person) Hello() string {
	return "Hi, " + p.Name
}

x := 5.Square()
person.Person{}.Hello()
```

The compiler desugars `x.Square()` to `Square(x)` when `x`’s type has no ordinary method with that name. Extensions are **not** interface satisfaction; they depend on **imports** and **name lookup** in the calling package. See [Extension methods](extension_methods.md) and [Built-in LINQ](linq.md).

## Sum types — enum variants

**Enums** provide **closed** variant polymorphism: a value is exactly one of a fixed set of variants, with optional payloads:

```go
enum Result {
	Ok(int)
	Err(string)
}

switch r {
case Ok(n):
	fmt.Println("ok", n)
case Err(msg):
	fmt.Println("err", msg)
}
```

Without a `default` case, the compiler checks **exhaustiveness** — every variant must be handled. This is compile-time polymorphism over a known finite set, unlike interfaces (open) or overloads (by signature). See [Enums](enums.md).

## Choosing a mechanism

| Goal | Prefer |
| ---- | ------ |
| Pluggable behavior across packages and unknown future types | **Interface** |
| One implementation shared by many unrelated types | **Generics** + constraints |
| Same name, different parameter lists (C#-style) | **Overloading** |
| Natural syntax for user-defined `+`, `==`, `[]`, … | **Operator overloading** |
| Add methods to `int`, `[]T`, or types in another module | **Extension methods** |
| Fixed set of shapes with payloads (Rust-style enums) | **Enums** |

Mechanisms can combine. Example: a generic `List[E]` with receiver methods and extension-based LINQ; an `enum` variant carrying a value that satisfies an `error` interface; overloaded `Parse` functions returning a `Result` enum.

## Relation to upstream Go

| Feature | Upstream Go | This fork |
| ------- | ----------- | --------- |
| Interfaces, type assertions, `errors.As` | Yes | Yes |
| Generics (functions, types) | Yes (1.18+) | Yes |
| Generic methods on generic types | Go 1.27 | Yes |
| Function / method overloading | No | [overloading.md](overloading.md) |
| Operator overloading | No | [operator_overloading.md](operator_overloading.md) |
| Extension methods | No | [extension_methods.md](extension_methods.md) |
| Algebraic enums | No | [enums.md](enums.md) |
