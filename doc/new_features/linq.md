# Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics. With `import "linq"`, chains run on `iter.Seq[T]` via **extension methods**. When the first receiver in a chain is `[]T`, the compiler calls optimized private slice implementations inside the `linq` package.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (deferred iteration until materialization)
- Minimal allocations; iterators and pipelines avoid unnecessary intermediate slices
- Public API is unified on `iter.Seq[T]`; use `linq.From(slice)` or slice extension syntax to start a chain

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

```go
import "linq"

nums := []int{1, 2, 3, 4, 5}
out := nums.Where(n => n < 5).Select(n => n + 1).ToList()
```

Slice receivers on the first call in a chain use compiler specialization to call unexported `*Slice` fast paths (for example `whereSlice`) instead of wrapping with `slices.Values`.

## Coverage vs .NET `System.Linq.Enumerable`

All **68** core `System.Linq.Enumerable` methods are implemented, including `FullJoin`.

Go also provides generators not in `Enumerable`: `From`, `Empty`, `Range`, `Repeat`, `InfiniteSequence`, `Sequence`, and `SelectBy` (alias for `Select`). `ThenBy` / `ThenByDescending` live on `Ordered[T]` (the `IOrderedEnumerable` role).

Go renames a few .NET names to match `iter.Seq[T]`: `AsSeq` (not `AsEnumerable`), `Len` / `LongLen` (not `Count` / `LongCount`), and `TryGetSeqLen` (not `TryGetNonEnumeratedCount`).

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

These methods exist but **not all** .NET overloads are covered:

**Predicate / filtering terminals** — use `Where(pred).First()` etc. as a workaround:

| Method | Missing overload |
|--------|------------------|
| `Any` | `Any()` without predicate |
| `Len`, `LongLen` | `Len(pred)`, `LongLen(pred)` (.NET: `Count`, `LongCount`) |
| `First`, `Last`, `Single` | `*(pred)` |
| `FirstOrDefault`, `LastOrDefault`, `SingleOrDefault` | `*(pred)`, `*(pred, default)` |

**Defaults and aggregates:**

| Method | Missing overload |
|--------|------------------|
| `DefaultIfEmpty` | No-arg version (zero value when empty) |
| `Aggregate` | `Aggregate(seed, accFn, resultSelector)` |
| `AggregateBy` | Factory-seed overload without explicit seed |

**Index-aware operators:**

| Method | Missing overload |
|--------|------------------|
| `Where` | `Where(fn(T, index) bool)` |
| `Select` | `Select(fn(T, index) U)` |
| `SelectMany` | Indexed collection selector; result-selector variants |

**Numeric selectors:**

| Method | Missing overload |
|--------|------------------|
| `Sum`, `Average` | `Sum(selector)`, `Average(selector)` — use `Select(selector).Sum()` |

**Custom equality** — Go uses `comparable` / `==` only; no `IEqualityComparer` overloads for:

`Contains`, `Distinct`, `DistinctBy`, `Except`, `ExceptBy`, `Intersect`, `IntersectBy`, `Union`, `UnionBy`, `SequenceEqual`, `CountBy`, `AggregateBy`, join key comparison.

**Indexing:**

| Method | Missing overload |
|--------|------------------|
| `ElementAt`, `ElementAtOrDefault` | `Index`-typed arguments (from-end / `^n` indexing) |

**SelectMany shapes:**

| Shape | Status |
|-------|--------|
| `fn func(T) iter.Seq[U]` | Implemented |
| `fn func(T) []U` | Internal only; not public |
| Collection + result selector | Not implemented |
| Indexed selectors | Not implemented |

**FullJoin** — the tuple-returning overload (`(TOuter?, TInner?)` pairs without a result selector) is not implemented; use `FullJoin` with `resultFn` and explicit `defaultOuter` / `defaultInner` zero values.
