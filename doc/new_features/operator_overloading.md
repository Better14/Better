# Operator Overloading

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
| `a[i]` or `a[i, j]` (multi-index) | `func []` `(a T, indices ...IndexTypes) U` |
| `a[i] = v`                        | `func []=` `(a T, indices ..., v V)` |
| `a[i:j]` / `a[i:j:k]`             | Optional: `func [:]` `(a T, i, j, k ...) U` if slice syntax is overloaded for the type (otherwise use ordinary methods) |


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
| `?.`, `??`, `!.` | Nilable / result syntax (see [nilable_types.md](nilable_types.md), [result_types.md](result_types.md)) |
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
