# Data Structures

Standard Go does not provide a built-in `set` type, queue/stack abstractions, or a generic binary tree. It does ship lower-level building blocks:


| Need                | Standard library today                                           |
| ------------------- | ---------------------------------------------------------------- |
| Growable sequence   | `[]T` + `append` (must reassign: `s = append(s, x)`)             |
| Set-like membership | `map[T]struct{}` (manual; no literal, no set algebra)            |
| Doubly linked list  | `container/list` (not typed; not a dedicated queue/stack API)    |
| Min-heap            | `container/heap` (you implement `heap.Interface`; min-heap only) |
| Max-heap            | `container/heap` with inverted `Less`                            |
| Binary tree         | Not in the standard library                                      |


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


|                           | `list[T]`                             | `LinkedList[T]`                      |
| ------------------------- | ------------------------------------- | ------------------------------------ |
| Backing                   | Dynamic array (slice)                 | Doubly linked nodes                  |
| Index access `At(i)`      | O(1)                                  | O(n)                                 |
| Append / pop at end       | O(1) amortized                        | O(1)                                 |
| Insert / remove at front  | O(n) shift                            | O(1)                                 |
| Insert / remove in middle | O(n) shift                            | O(1) with node cursor; O(n) by index |
| Memory                    | Contiguous; less overhead per element | Pointer per node; extra allocations  |


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
t := tree.New[int, string]()  // *tree.BinaryTree[int, string]
t.Insert(2, "two")
t.Insert(1, "one")
v, ok := t.Search(2)          // "two", true
t.Inorder(func(k int, v string) { /* sorted by k */ })
t.Delete(1)
```

Operations: `Insert`, `Search`, `Delete`, `Min`, `Max`, `Inorder`, `Preorder`, `Postorder`, `Len`, `Height`. Balanced variants (e.g. AVL/red-black) may be provided as `tree.NewBalanced` or a separate `balancetree` package depending on implementation.

### Summary


| Type                  | Literal              | Standard Go equivalent                     |
| --------------------- | -------------------- | ------------------------------------------ |
| `list`                | `list[T]{...}`       | `[]T` + `append`                           |
| `LinkedList`          | — (`linkedlist.New`) | `container/list` (`any`, manual `Element`) |
| `set`                 | `{}T{...}`           | `map[T]struct{}`                           |
| `queue`               | — (`queue.New`)      | slice + mutex, or `LinkedList` discipline  |
| `stack`               | — (`stack.New`)      | slice, or `container/list`                 |
| `minheap` / `maxheap` | —                    | `container/heap` + custom `Less`           |
| `tree`                | —                    | third-party or hand-rolled                 |


---
