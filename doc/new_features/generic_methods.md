# Generic Methods (Go 1.27)

Go 1.27 extends generic types with **methods whose receivers are generic**. This is the mechanism behind C#-style “attach query methods to any enumerable” — but Go expresses it through **named generic types** and (for built-in slices) **compiler desugaring**, not C# `static` extension methods.

### Methods on generic types

If the receiver base type is generic, the receiver specification must declare matching type parameters. Those parameters are in scope for the method body (like type parameters on a generic struct):

```go
type Pair[A, B any] struct {
	a A
	b B
}

func (p Pair[A, B]) Swap() Pair[B, A] {
	return Pair[B, A]{a: p.b, b: p.a}
}
```

A method on `Lazy[T]` uses the same pattern:

```go
type Lazy[T any] struct { /* iterator state */ }

func (l Lazy[T]) Where(pred func(T) bool) Lazy[T] {
	// filter l using pred; return a new lazy pipeline
}
```

Here `T` comes from the receiver `Lazy[T]` — you do **not** write a separate `[T any]` on the method when `T` is already declared by the receiver.

Invalid (receiver must be a defined type, not a bare type parameter):

```go
func (o *T) Where(pred func(T) bool) Lazy[T]  // invalid
```

### Generic methods (extra type parameters on the method)

A method may declare **additional** type parameters after the method name, like a generic function bound to a receiver. This mirrors C# `Select` projecting to a new element type:

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

Such a declaration is a **generic method**. It must be **instantiated** (explicitly or by inference) before it can be called, the same as generic functions.

### C# extension methods vs Go

In C#, `Where`, `Select`, etc. are **extension methods** on `IEnumerable<T>` — any type implementing that interface picks them up.

Go does not have `IEnumerable<T>`, but this fork adds **[Extension Methods](extension_methods.md)** — the general mechanism C# uses, adapted for Go generics and `iter.Seq[T]`. Until extensions are implemented, enumerables also use the legacy paths below:


| Enumerable shape | How methods attach | Example |
| ---------------- | ------------------ | ------- |
| Slice / array `[]T` | **Extension methods** (planned) or legacy compiler **desugar** to `linq` when `import "linq"` | `a.Where(s => s == "a")` |
| `iter.Seq[T]` | **Extension methods** on `iter.Seq[T]` | `seq.Where(pred)` |
| Named generic sequence type | Real **receiver methods** on the type | `func (l Lazy[T]) Where(…)`, `func (l list[T]) Where(…)` |
| Other iterables | Extension methods after adaptation, or convert then chain | `set.Values().Where(…)` |


Go still does **not** allow methods on `[]T` itself (slice types are not defined types). To add methods directly in library code, use a defined generic type such as `type List[T any] []T` or the std `list[T]`, `Lazy[T]`, etc.

### Limitations (Go 1.27 and this fork)

- **No methods on slice types.** Use a type alias/definition (`type List[E any] []E`), a wrapper (`Lazy[T]`), or slice LINQ desugaring (below).
- **Receiver base type** must be a defined type in the same package; it cannot be a pointer or interface type, and generic aliases have restrictions (see the language spec).
- **Generic methods with method-local type parameters** (e.g. `Select[F any]` on `Lazy[T]`) are part of the Go 1.27 language, but the compiler in this fork may **ICE** when exporting some generic methods from generic types (`internal compiler error` in `noder/writer.go`). Until that is fixed, operations that need an extra type parameter (`Select`, `OrderBy` with key type `K`, `GroupBy` with key type `K`) may remain **package functions** (`linq.Select`, `linq.LazySelectBy`, …) even when simpler methods (`Where`, `Take`, `Skip`) work as receivers.
- **Instantiation:** generic methods must be instantiated; type inference at the call site applies when the compiler can infer method type arguments from arguments (same rules as generic functions).
- **Not in upstream Go** before 1.27.

### Relation to LINQ

Built-in LINQ syntax (`nums.Where(…).Select(…).ToList()`) is intended to be implemented with **[Extension Methods](#extension-methods)** on `iter.Seq[T]` (with automatic `slices.Values` for `[]T`). Until that lands, the compiler also uses:

1. **Legacy desugaring** for slices/arrays (and continued chains on `linq.Lazy[T]`) via a fixed `linqMethods` table.
2. **Receiver methods** on `linq.Lazy[T]` and container types where supported.
3. **Package functions** in `import "linq"` as the lowering target.

See [Built-in LINQ](linq.md) for usage examples.
