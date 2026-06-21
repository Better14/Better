# Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics. With `import "linq"`, chains run on `iter.Seq[T]` via **extension methods**. When the first receiver in a chain is `[]T`, the compiler calls optimized private slice implementations inside the `linq` package.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (deferred iteration until materialization)
- Minimal allocations; iterators and pipelines avoid unnecessary intermediate slices
- Public API is unified on `iter.Seq[T]`; call LINQ methods directly on slices and other supported collections — no `linq.From` wrapper needed

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

// grouping (lazy until enumerated) — struct literal, not a C# tuple
byMod := nums.GroupBy(n => n % 3).Select(g => struct {
	Key   int
	Count int
}{g.Key, g.Count()}).ToList()

// distinct after transform
unique := nums.Select(n => n / 2).Distinct().OrderBy(n => n).ToList()

// first match or default
found := users.Where(u => u.Active).Select(u => u.Email).FirstOrDefault()
```

Predicate and projection arguments are typically single-expression lambdas using `=>`; parameter types are inferred from the LINQ method signature.

```go
import "linq"

nums := []int{1, 2, 3, 4, 5}
out := nums.Where(n => n < 5).Select(n => n + 1).ToList()
```

Slice receivers on the first call in a chain use compiler specialization to call unexported `*Slice` fast paths (for example `whereSlice`) instead of wrapping with `slices.Values`.

## Coverage vs .NET `System.Linq.Enumerable`

All **68** core `System.Linq.Enumerable` methods are implemented, including `FullJoin`.

Go also provides sequence generators not in `Enumerable`: `Empty`, `Range`, `Repeat`, `InfiniteSequence`, `Sequence`, and `SelectBy` (alias for `Select`). `ThenBy` / `ThenByDescending` live on `Ordered[T]` (the `IOrderedEnumerable` role). Use `linq.From` only when you need an explicit `iter.Seq[T]` value (for example, to pass into a function parameter typed as `iter.Seq[T]`); slice chains do not require it.

Go renames a few .NET names to match `iter.Seq[T]`: `AsSeq` (not `AsEnumerable`) and `TryGetSeqLen` (not `TryGetNonEnumeratedCount`).

## Not implemented

### Other namespaces on the .NET `IEnumerable<T>` page

These extension families from the Microsoft docs are **not** in the Go `linq` package:

| Namespace | Examples |
|-----------|----------|
| `System.Linq.ParallelEnumerable` | `AsParallel` |
| `System.Linq.Queryable` | `AsQueryable` |
| `System.Linq.AsyncEnumerable` | `ToAsyncEnumerable` |
| `System.Collections.Immutable` | `ToImmutableArray`, `ToImmutableList`, `ToImmutableDictionary`, … |
| `System.Collections.Frozen` | `ToFrozenDictionary`, `ToFrozenSet` |
| `System.Data` | `CopyToDataTable` |
| `System.Xml.Linq` | `Ancestors`, `Descendants`, `Elements`, … |

### Missing overloads and variants

Most common .NET overload variants are implemented (predicate terminals, indexed operators, selector `Sum`/`Average`, `SelectMany` result-selector shapes, `DefaultIfEmpty()`, and three-argument `Aggregate`).

**Custom equality (`IEqualityComparer`)** — intentionally omitted. Go LINQ uses `comparable` and `==` only; there are no `IEqualityComparer` overloads for `Contains`, `Distinct`, `DistinctBy`, `Except`, `ExceptBy`, `Intersect`, `IntersectBy`, `Union`, `UnionBy`, `SequenceEqual`, `CountBy`, `AggregateBy`, or join key comparison.

| C# approach | Go LINQ equivalent |
|-------------|-------------------|
| `Distinct(seq, comparer)` | `Distinct(seq)` when `T comparable`, or `DistinctBy(seq, keyFn)` |
| `GroupBy` with custom key equality | `GroupBy(seq, keyFn)` with `K comparable` |
| Case-insensitive string distinct | `DistinctBy(s => strings.ToLower(s))` |

**Why `==` is enough:** In Go, `comparable` includes structs whose fields are all comparable, and `==` compares them field-wise — the common LINQ case. C# often needs `IEqualityComparer` because reference types default to identity equality and custom rules (compare by ID, case-insensitive strings) are expressed via `Equals`/`GetHashCode` overrides or explicit comparers. Go expresses those rules with `*By` key projections: compare `keyFn(x) == keyFn(y)` on a `comparable` key instead of plugging in a comparer object.

**When `==` is not enough:** If `T` is not `comparable` (e.g. a struct containing a slice), use `DistinctBy` / `ExceptBy` / `UnionBy` / `GroupBy` with a `comparable` key. If equality should ignore some fields, project to the fields that matter (`DistinctBy(u => u.ID)`). Float `NaN` behaves like ordinary Go `==` (and cannot be a map key).

**Indexing:**

| Method | Missing overload |
|--------|------------------|
| `ElementAt`, `ElementAtOrDefault` | `Index`-typed arguments (from-end / `^n` indexing) |

**SelectMany shapes:**

| Shape | Status |
|-------|--------|
| `fn func(T) iter.Seq[U]` | Implemented |
| `fn func(T) []U` | Use `collectionFn func(T) []C` overload with result selector |
| Collection + result selector | Implemented |
| Indexed selectors | Implemented |

**FullJoin** — the tuple-returning overload (`(TOuter?, TInner?)` pairs without a result selector) is not implemented; use `FullJoin` with `resultFn` and explicit `defaultOuter` / `defaultInner` zero values.
