# New `?` Result Syntax

This document describes the new result-type and postfix-unwrap syntax based on `?`.

## Overview

Go now supports a result shorthand:

- `(T, error)` can be written as `T?`
- `expr?` unwraps a `(T, error)` expression

The goal is to reduce boilerplate for error propagation while preserving the same runtime behavior as explicit `if err != nil { ... }` checks.

## Function Result Type Shorthand

You can declare a function returning `(int, error)` as `int?`.

Before:

```go
func myFunc() (int, error) {
	a, err := myFunc2()
	if err != nil {
		return 0, err
	}
	return a, nil
}
```

After:

```go
func myFunc() int? {
	a := myFunc2()?
	return a, nil
}
```

`int?` is semantically equivalent to `(int, error)`.

## Postfix `?` Operator

`expr?` can be used when `expr` has type `(T, error)` (or `T?`).

Behavior:

1. Evaluate `expr`
2. If `err != nil`, return early from the current function with:
   - zero value of the function's value result
   - the error
3. Otherwise, yield the unwrapped `T` value

This is similar to Rust's `?` operator in spirit.

## Conceptual Representation

A `T?` can be thought of as a pair:

```go
struct TValueOrErr {
	value T
	err   error
}
```

For example, `int?` corresponds conceptually to:

```go
struct intOrErr {
	value int
	err   error
}
```

This is a conceptual model for documentation; syntax-level behavior is defined by compiler lowering and type checking.

## Additional Intended Usage Patterns

The requested usage patterns are:

```go
var a := int?
if a.err != nil { a.value }
if a.err == nil { a.err } else { a.value }
var a, err := int?   // destructure into value and error
a?.someProperty      // early-return if error, then access property
```

These examples describe the intended ergonomics around result-style values and chained unwrapping.

## Notes

- `T?` should be treated as the canonical shorthand for `(T, error)`.
- `expr?` should only be used in contexts where early-returning an error is valid for the enclosing function's signature.

## Postfix `!` Force-Unwrap Operator

Go now also supports postfix `!` as a force-unwrap operator.

`expr!` evaluates a result-like expression and:

1. Panics with the error if the error is non-nil
2. Otherwise returns the underlying value

This is useful when failure is unexpected and should be fatal, similar to a forced unwrap.

### `(T, error)` Example

```go
func someFunc() (int, error) {
	// ...
}

var a = someFunc()! // if err != nil: panic(err); otherwise a is int
```

Equivalent behavior:

```go
tmp, err := someFunc()
if err != nil {
	panic(err)
}
var a = tmp
```

### `T?` Example

```go
func someFunc2() int? {
	// ...
}

var b = someFunc2()! // if result err != nil: panic(err); otherwise b is int
```

Equivalent conceptual behavior:

```go
tmp, err := someFunc2()
if err != nil {
	panic(err)
}
var b = tmp
```

### Notes

- `expr!` accepts values of type `(T, error)` and `T?`.
- `expr!` does not early-return; it panics on error.
- `expr!` should be used sparingly and only when panic-on-error is desired.

## Function and Method Overloading

Go now supports function and method overloading.

Multiple declarations with the same name are valid as long as their parameter lists differ by type and/or arity.

For example:

```go
func myFunc(a int) {
}

func myFunc(a int64) {
}

func myFunc(a int, b int) {
}
```

All of the above declarations are valid.

### Overload Resolution

At a call site, the compiler resolves the overload by matching argument count and argument types.

- `myFunc(10)` resolves to `func myFunc(a int)`
- `myFunc(int64Value)` resolves to `func myFunc(a int64)`
- `myFunc(1, 2)` resolves to `func myFunc(a int, b int)`

### Notes

- Overloading applies to both package-level functions and methods.
- Overload sets must be unambiguous for all valid calls.

## If Expressions

Go supports `if` as an expression that evaluates to a value.

```go
a := if 5 < 6 { 1 } else { 2 }
```

Both branches must be expressions with compatible types. The result type is the common type of the branch expressions.

## Switch Expressions

Go supports `switch` as an expression that evaluates to a value.

```go
a := switch x {
case 1:
	"one"
case 2:
	"two"
default:
	"other"
}
```

Each case arm must be an expression (or a single expression after `:`). All arms must have compatible types. The result type is the common type of the case expressions.

## Lambda Syntax (`=>`)

For single-expression functions, Go supports arrow lambda syntax. Parameter types are inferred from context.

Before:

```go
func(a, b, c int) int { return a + b + c }
```

After:

```go
(a, b, c) => a + b + c
```

If the function body requires more than one expression or any statement, use the standard `func` syntax:

```go
func(a, b, c int) {
	sum := a + b + c
	return sum
}
```

### Notes

- `=>` lambdas are limited to a single expression body.
- Parameter types are inferred when the lambda appears in a typed context (e.g. assignment, argument, return).
- Multi-statement or multi-expression bodies must use `func(...) { ... }`.

## Default Function Arguments

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

### Notes

- Default values are compile-time constants or constant expressions evaluable at compile time (same spirit as C# constant defaults).
- Default arguments are not supported on `=>` lambdas; use a named `func` or a wrapper.
- Not valid in upstream Go.

## Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (e.g. deferred iteration until materialization)
- Minimal allocations; iterators and pipelines should avoid unnecessary intermediate slices

Step-by-step example:

```go
nums := []int{1, 2, 3, 4, 5}
evens := nums.Where(n => n%2 == 0)
doubled := evens.Select(n => n * 2)
first := doubled.First()
```

### Chained one-liners

Pipelines compose left-to-right; each stage is lazy until a terminal operator (`First`, `ToList`, `Sum`, etc.) runs.

```go
nums := []int{1, 2, 3, 4, 5, 6, 7, 8}

// filter → map → first
firstEvenDouble := nums.Where(n => n%2 == 0).Select(n => n * 2).First()

// filter → order → take
topThree := nums.Where(n => n > 2).OrderByDescending(n => n).Take(3).ToList()

// map → aggregate
sumOfSquares := nums.Select(n => n * n).Sum()

// skip → take → map
page := nums.Skip(10).Take(20).Select(n => fmt.Sprintf("%d", n)).ToList()

// any / all over a chain
hasLargeEven := nums.Where(n => n%2 == 0).Any(n => n > 100)
allPositive := nums.Select(n => n - 1).All(n => n >= 0)

// strings: filter → project → join
names := []string{"alice", "", "bob", "carol"}
line := names.Where(s => len(s) > 0).Select(s => strings.ToUpper(s)).Aggregate((a, b) => a + ", " + b)

// grouping (lazy until enumerated)
byMod := nums.GroupBy(n => n % 3).Select(g => (g.Key, g.Count())).ToList()

// distinct after transform
unique := nums.Select(n => n / 2).Distinct().OrderBy(n => n).ToList()

// first match or default
found := users.Where(u => u.Active).Select(u => u.Email).FirstOrDefault()
```

Predicate and projection arguments are typically single-expression lambdas using `=>`; parameter types are inferred from the LINQ method signature.

LINQ extensions are provided as methods on supported sequence types (slices, arrays, and other iterable types as defined by the standard library).

## Data Structures

Standard Go does not provide a built-in `set` type, queue/stack abstractions, or a generic binary tree. It does ship lower-level building blocks:

| Need | Standard library today |
|------|-------------------------|
| Growable sequence | `[]T` + `append` (must reassign: `s = append(s, x)`) |
| Set-like membership | `map[T]struct{}` (manual; no literal, no set algebra) |
| Doubly linked list | `container/list` (not typed; not a dedicated queue/stack API) |
| Min-heap | `container/heap` (you implement `heap.Interface`; min-heap only) |
| Max-heap | `container/heap` with inverted `Less` |
| Binary tree | Not in the standard library |

This fork adds first-class container types in the standard library (or as built-in generic types) with literal syntax where noted below.

### List

A `list` is a growable, ordered sequence (like a slice) with methods that mutate in place. You no longer need to reassign the result of `append`:

Before (slice):

```go
var nums []int
nums = append(nums, 1)
nums = append(nums, 2, 3)
```

After (`list`):

```go
nums := list.New[int]()
nums.Append(1)
nums.Append(2, 3)   // variadic; same as multiple appends

names := list.Of("a", "b", "c")  // from values
names.Append("d")
```

**Literal syntax** (optional; mirrors slice literals):

```go
a := list[int]{1, 2, 3}
b := list[int]{}           // empty
```

Conversion and indexing:

```go
s := a.ToSlice()           // []int — share or copy per implementation
c := list.FromSlice(s)

x := a.At(0)               // or a[0] if indexer syntax is enabled
a.Set(1, 99)
n := a.Len()
a.Insert(1, 42)            // insert at index
a.RemoveAt(1)
last, ok := a.Pop()        // remove and return last element
a.Clear()
```

`Append` returns nothing (or returns `*list[T]` for chaining, e.g. `nums.Append(1).Append(2)`). Slices remain valid and interoperate via `ToSlice` / `FromSlice`; use `list` when you want method-style growth without `s = append(s, x)`.

LINQ methods (`Where`, `Select`, `OrderBy`, etc.) are defined on `list[T]` the same as on slices.

### LinkedList

`LinkedList[T]` is a doubly linked, ordered sequence. Unlike `list[T]` (slice-backed), inserting or removing in the middle does not shift a backing array; growth does not trigger slice reallocation.

Standard Go’s `container/list` is untyped (`Value any`) and uses external `*list.Element` handles. `LinkedList[T]` is generic and keeps nodes internal unless you opt into cursor APIs.

```go
ll := linkedlist.New[int]()
ll.Append(1)           // tail
ll.Append(2, 3)
ll.Prepend(0)          // head

ll.PushFront(99)       // aliases for head/tail
ll.PushBack(100)
v := ll.PopFront()     // 99
w := ll.PopBack()      // 100

n := ll.Len()
first := ll.Front()    // *int or (int, bool) — peek head value
last := ll.Back()

ll.InsertAt(2, 42)     // by index; O(n) walk; prefer node APIs when hot
ll.RemoveAt(1)
ll.RemoveValue(42)     // first matching element

for _, x := range ll.Iterate() {  // or range ll
	_ = x
}

s := ll.ToSlice()      // snapshot in linked order
other := linkedlist.FromSlice([]int{4, 5, 6})
```

**When to use which**

| | `list[T]` | `LinkedList[T]` |
|---|-----------|-----------------|
| Backing | Dynamic array (slice) | Doubly linked nodes |
| Index access `At(i)` | O(1) | O(n) |
| Append / pop at end | O(1) amortized | O(1) |
| Insert / remove at front | O(n) shift | O(1) |
| Insert / remove in middle | O(n) shift | O(1) with node cursor; O(n) by index |
| Memory | Contiguous; less overhead per element | Pointer per node; extra allocations |

Use `LinkedList[T]` for frequent front/middle edits, stable iterators while mutating elsewhere (with cursor API), or algorithms that splice sublists. Use `list[T]` for index-heavy work and cache-friendly sequential access.

LINQ methods apply on `LinkedList[T]` via iteration or after `ToSlice()`, depending on implementation.

### Set

A `set` is an unordered collection of unique elements. Elements must be `comparable`.

**Literal syntax** (mirrors `[]T{...}` but uses `{}` instead of `[]`):

```go
a := {}string{"foo", "bar", "bar", "baz"}
// len(a) == 3; duplicate "bar" is dropped at construction
```

Yes — `{}T{ ... }` is the intended literal form for this extension. It is not valid in upstream Go. An empty set is written `{}string{}`.

Other construction and use:

```go
b := set.Make[int]()           // empty set
b.Add(1)
b.Add(2)
b.Contains(1)                  // true
b.Delete(2)

c := {}int{1, 2, 3}
d := c.Union({}int{3, 4, 5})   // {}int{1, 2, 3, 4, 5}
e := c.Intersect({}int{2, 99}) // {}int{2}

for v := range a.Values() {    // or range a
	_ = v
}
```

Set operations: `Add`, `Delete`, `Contains`, `Len`, `Union`, `Intersect`, `Difference`, `Subset`, `Equal`, and iteration. LINQ-style methods (`Where`, `Select`, etc.) apply when converting via `.Values()` or when defined on the set type.

### Queue (FIFO)

```go
q := queue.New[int]()
q.Enqueue(1)
q.Enqueue(2)
v := q.Dequeue()   // 1
n := q.Len()
front := q.Peek()  // 2; does not remove
```

Use for breadth-first traversal, work queues, and ordered processing. Not a substitute for channels when you need concurrency safety; use channels for goroutine communication.

### Stack (LIFO)

```go
s := stack.New[int]()
s.Push(10)
s.Push(20)
v := s.Pop()       // 20
top := s.Peek()    // 10
```

Use for depth-first traversal, undo buffers, and expression parsing.

### Min-heap and max-heap

Typed heaps replace the boilerplate of implementing `heap.Interface` by hand.

```go
h := minheap.New[int]()   // smallest int at top
h.Push(5)
h.Push(1)
h.Push(3)
x := h.Pop()              // 1

mh := maxheap.New[string]()
mh.Push("a")
mh.Push("z")
mh.Pop()                  // "z"
```

Supports `Push`, `Pop`, `Peek`, `Len`, `Fix` (after changing a stored element), and optional `PushHeap`/`PopHeap` batch patterns. `container/heap` remains available for custom ordering; prefer `minheap` / `maxheap` when the element type is the priority key.

### Binary tree

A generic ordered binary tree (typically BST) for keyed lookup and ordered traversal:

```go
t := tree.New[int, string]()  // key int, value string
t.Insert(2, "two")
t.Insert(1, "one")
v, ok := t.Search(2)          // "two", true
t.Inorder(func(k int, v string) { /* sorted by k */ })
t.Delete(1)
```

Operations: `Insert`, `Search`, `Delete`, `Min`, `Max`, `Inorder`, `Preorder`, `Postorder`, `Len`, `Height`. Balanced variants (e.g. AVL/red-black) may be provided as `tree.NewBalanced` or a separate `balancetree` package depending on implementation.

### Summary

| Type | Literal | Standard Go equivalent |
|------|---------|-------------------------|
| `list` | `list[T]{...}` | `[]T` + `append` |
| `LinkedList` | — (`linkedlist.New`) | `container/list` (`any`, manual `Element`) |
| `set` | `{}T{...}` | `map[T]struct{}` |
| `queue` | — (`queue.New`) | slice + mutex, or `LinkedList` discipline |
| `stack` | — (`stack.New`) | slice, or `container/list` |
| `minheap` / `maxheap` | — | `container/heap` + custom `Less` |
| `tree` | — | third-party or hand-rolled |

---

## Optional Features

The following features are planned or under consideration; they are not required for the core language extensions above.

### Operator Overloading

Types may define operators via special method syntax:

```go
func (m *Matrix) +{ /* m + other */ }
func (m *Matrix) +={ /* m += other */ }
func (m *Matrix) *{ /* m * other */ }
func (m *Matrix) []{ /* indexer: m[key] */ }
```

Supported operators and their method forms include (non-exhaustive):

- Arithmetic: `+`, `-`, `*`, `/`, `%`, and compound forms (`+=`, `-=`, etc.)
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=` (where defined)
- Indexing: `[]` for get/set via indexer methods

Overload resolution selects the receiver type’s operator method when the corresponding built-in operator is used with that type.