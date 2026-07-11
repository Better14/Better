# Panics and stack traces

Bow extends panic behavior so failures are easier to diagnose in production: **unrecovered panics print a stack trace to stderr**, and panic values that are structured errors include their captured traces.

## Unrecovered panics

When a panic is not recovered by `recover()`, the runtime:

1. Runs deferred functions (`defer`) in LIFO order.
2. Prints each panic value in the chain (`panic: …`) to **stderr**.
3. Prints a **full stack trace** for the goroutine (and often other goroutines) to **stderr**.
4. Exits with status 2.

This matches upstream Go’s fatal panic path (`runtime.fatalpanic` → `printpanics` → `tracebackothers`). Bow keeps that behavior and extends what appears in the panic message when the value is a structured error.

Example:

```
panic: nil pointer dereference

goroutine 1 [running]:
example.com/pkg.(*Connection).Subroute(...)
    /path/connection.go:293
main.main()
    /path/main.go:42
...
```

## Stack traces on panic values

### `panic(err)` with `*errors.Base`

When the panic argument implements structured errors (`*errors.Base`, `*errors.Error`, or types embedding `Base`), the runtime formats the value with **`String()`** (full message + inner chain + stack frames) before printing:

```go
err := errors.New("connection lost")
panic(err) // stderr shows message and stack captured at errors.New
```

Prefer `errors.New` / `errors.Wrap` for panic values you expect to inspect in logs.

### Compiler-inserted panics

Panics inserted by the compiler capture a stack at the **failure site**, including:

- **Nil receiver method calls** ([nil_receivers.md](nil_receivers.md))
- **Nil channel receive** ([fixed_weird_behaviors.md](fixed_weird_behaviors.md))
- **Nil pointer dereference**
- **Nil unwrap** of `T?`, `T!`, or `*T?` (force cast / `!.value`)
- **Bounds / type assertion** failures

These use the same `gopanic` path as `panic(...)` in source.

## `recover()`

`recover()` inside a deferred function still returns the panic value and suppresses the fatal traceback for that goroutine. The stack trace is **not** printed to stderr when recovery succeeds.

If a recovered panic is **re-panicked**, the full chain is printed on the final unrecovered exit.

## Exceptions — system and runtime faults

Some failures use **`runtime.throw`** instead of user-visible `panic` / `gopanic`. These are intentionally abrupt and may **not** produce the usual panic message and stack format:

| Kind | Typical handling |
| ---- | ---------------- |
| Out of memory (`throw("out of memory")`) | Process abort via `throw`; **no full user stack** by default — see below |
| Panic on system stack | `throw` after brief message |
| Panic during `malloc`, with locks held, or in `preemptoff` | `throw` — unsafe to unwind |
| `fatal` / `fatalthrow` internal errors | May print a **limited** traceback via `dopanic_m`; not the full `fatalpanic` path |

User code cannot recover from `throw`. Treat these as process-level faults, not catchable panics.

Ordinary **`panic(...)`** and compiler nil checks always go through **`gopanic`** and get the standard stderr stack trace when unrecovered.

### Why not always trace `throw` / OOM?

`throw` is deliberately **nosplit** and runs on the **system stack** with minimal allocation. A full traceback walks every frame, resolves symbols, and may **allocate memory** — risky when the process is already out of memory, corrupting the stack, or holding runtime locks. `fatalthrow` may still print a short traceback when safe; expanding that to match every `gopanic` dump would add **crash-path cost only** (not steady-state), but can **worsen OOM** or fail entirely in the worst cases.

**Summary:** Full stacks on normal panics are **low cost** (already implemented). Requiring full stacks on every `throw`/`OOM` is **moderate-to-high risk** at exit time and offers little benefit when the process is about to die anyway.

## Logging and tests

- **`log.Fatal` / `log.Panic`** — log and exit; structured errors use `String()` for full detail.
- **`testing` package** — test panics are caught by the test harness; stderr trace appears when running `go test -v` or on failure.

For production services, unrecovered panics should be rare; prefer `T!` / `!` and explicit error returns. Use `panic` for invariant violations and “impossible” branches after strict pointers and no-nil-receiver semantics.

## Related docs

- [Structured errors](errors.md) — `errors.New`, stack capture, `%+v`
- [No nil receivers](nil_receivers.md)
- [Nilable types](nilable_types.md) — `?.`, force cast panics
- [Result types](result_types.md) — `!.value` panics
