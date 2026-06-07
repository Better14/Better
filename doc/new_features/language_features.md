# Language features

Miscellaneous language semantics in this fork that differ from or extend classic Go behavior.

## Per-iteration loop variables

In `for` loops with a **`range`** clause, each iteration binds **fresh** loop variables. Closures created in the loop body capture the **current iteration’s** values, not a single shared variable reused across iterations.

This applies to all loop variables introduced by `range` (index/value, key/value, or a single value).

### Example

```go
options := []string{"a", "b", "c"}
funcs := make([]func(), len(options))

for i, v := range options {
	funcs[i] = func() {
		fmt.Println(v)
	}
}

for _, f := range funcs {
	f()
}
// Output:
// a
// b
// c
```

On classic Go (before Go 1.22), the same program typically printed `c` three times because `i` and `v` were reused for every iteration.

### Scope

- **`for i, v := range x`** — `i` and `v` are per-iteration.
- **`for k, v := range m`** — `k` and `v` are per-iteration (independent of map iteration order below).
- **`for v := range x`** — `v` is per-iteration.
- **`for i := range n`** — `i` is per-iteration.

Body-scoped redeclarations (`for i, v := range x { i, v := … }`) still create inner bindings as today.

### Notes

- This is the **default** in this fork; you do not need `GOEXPERIMENT=loopvar` or a language version flag.
- If a closure must share one variable across iterations, declare it outside the loop and assign explicitly inside the body.

## Deterministic map iteration order

Map iteration order is **stable and predictable** within a program run: `for k := range m` and `for k, v := range m` visit entries in **insertion order** — the order keys were **first** inserted into the map.

### Rules

| Operation | Effect on iteration order |
| --------- | ------------------------- |
| First insert of key `k` | `k` is appended at the end |
| Update value for existing `k` | Position unchanged |
| Delete `k`, then insert `k` again | `k` moves to the end (new insertion) |
| `clear(m)` | Empty map; new inserts start fresh |

Example:

```go
m := map[string]int{}
m["c"] = 3
m["a"] = 1
m["b"] = 2
m["a"] = 10 // update in place

for k := range m {
	fmt.Println(k)
}
// Output: c, a, b  (insertion order of first appearance)
```

### Rationale

Upstream Go randomizes map iteration to discourage reliance on order. This fork chooses **insertion order** because it is easy to reason about, matches common expectations from other languages, and avoids surprising test flakes that depend on map order.

### Performance and implementation

The runtime may record insertion order alongside existing map storage (for example an append-only key list or links between entries). The goal is **minimal overhead** on ordinary `m[k] = v` and `delete(m, k)` — not a separate ordered map type in the language.

- **Not serialized** — order is not guaranteed across processes, `encoding/gob`/`json` round-trips, or map copies unless documented otherwise by those APIs.
- **Concurrent maps** — iteration order follows the same rules for a single map value; concurrent writers still require synchronization as today.
- **Mutations during `range`** — behavior matches classic Go: new keys may or may not appear in the current iteration; deleted keys are skipped. Order rules apply to the map’s logical insertion sequence.

### Relation to loop variables

Map **iteration order** (which key comes first) is separate from **loop variable capture** (which `k`/`v` a closure sees). Both features apply together in:

```go
for k, v := range m {
	go func() {
		fmt.Println(k, v) // this iteration's k and v; keys visited in insertion order
	}()
}
```

## Quick reference

| Topic | Behavior |
| ----- | -------- |
| `for i, v := range x` in closure | Captures this iteration’s `i`, `v` |
| `for k := range m` | Keys in insertion order |
| Update `m[k]` | Keeps `k`’s position |
| Re-insert after delete | `k` at end |
