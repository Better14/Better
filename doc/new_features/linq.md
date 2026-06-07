# Built-in LINQ

Go includes built-in LINQ-style query operations that mirror C# naming and semantics. Prefer **[Extension Methods](extension_methods.md)** for new code; `import "linq"` currently enables legacy method desugaring on slices and `linq.Lazy[T]`.

- Same method names as C# (`Where`, `Select`, `OrderBy`, `GroupBy`, `First`, `ToList`, etc.)
- Lazy evaluation where applicable (e.g. deferred iteration until materialization)
- Minimal allocations; iterators and pipelines should avoid unnecessary intermediate slices
- Target model: extensions on `iter.Seq[T]` with `[]T` adaptation; see [Extension Methods](extension_methods.md)

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
evens := nums.Where(n => n%2 == 0)       // desugared to linq.Where
doubled := evens.Select(n => n * 2)    // linq.LazySelectBy on Lazy[int]
first := doubled.First()                 // terminal: materializes one element
```

LINQ extensions are provided as methods on supported sequence types (slices, arrays, `linq.Lazy[T]`, and other iterable types as defined by the standard library). Under the hood, slice/array calls desugar to `linq` functions; named types such as `Lazy[T]` may define real receiver methods where supported.

