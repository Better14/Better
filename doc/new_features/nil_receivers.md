# No nil receivers

Upstream Go allows calling a method on a **nil pointer receiver** — the method runs, and the author may defensively check `if c == nil { … }` inside the body. Bow changes this: a direct method call on a nil pointer **panics at the call site**, not inside the method.

## Motivation

Defensive nil-receiver branches are easy to get wrong, hide real bugs, and encourage chained calls like `mgr.Connection(host).Subroute(path)` that silently rely on in-method guards. Failing fast at the call site makes “forgot to check for nil” a loud panic with a stack trace, and keeps method bodies focused on non-nil logic.

## Behavior

In Bow, pointer-receiver method calls always panic at the call site when the receiver is nil:

```go
var c *Connection
c.Subroute("api") // panics here — Subroute's body never runs with c == nil
```

| Form | Nil behavior |
| ---- | ------------ |
| `ptr.Method(args)` — pointer receiver | **Panic at call site** if `ptr` is nil |
| `iface.Method(args)` — interface value | Panic at call site if interface is nil (unchanged) |
| `T.Method` — method value | Panic when the value is **invoked** with a nil receiver (unchanged) |
| `(*T).Method` — method expression | No check until invoked; nil first argument panics at that call |
| Value receiver `t.Method()` | Unchanged (receiver is a value, not a nil pointer) |

### Compiler implementation

Before lowering `x.M(…)` to `T.M(x, …)`, the compiler inserts a nil check on the pointer receiver (`OCHECKNIL`). The panic uses the runtime `gopanic` path and includes a stack trace (see [Panics and stack traces](panics.md)).

This is always on in Bow — there is no `go.mod` toggle.

## Migration patterns

### Remove in-method nil-receiver guards

**Before (upstream / defensive):**

```go
func (c *Connection) Subroute(s string) *Subroute {
	if c == nil {
		return nil
	}
	return &Subroute{Connection: c, route: s}
}
```

**After (Bow):**

```go
func (c *Connection) Subroute(s string) *Subroute {
	return &Subroute{Connection: c, route: s}
}
```

[modernize](https://github.com/Bow5/modernize) removes `if recv == nil { return … }` / `panic(…)` guards in pointer-receiver methods when `remove_nil_receiver_guards` is enabled (default).

### Optional call sites — use `?.`

When nil is **expected** and the call should be skipped, use null-conditional syntax at the **call site** only where it matches the old `if recv == nil { return nil }` behavior:

```go
// Subroute used to guard nil receiver and return nil — equivalent to:
sub := mgr.Connection(host)?.Subroute(path)

// Process has no nil-receiver guard — keep regular syntax (nil conn panics at call site):
res := mgr.Connection(host).Process()
```

Chained calls — add `?.` per link only when **that** method had a nil-receiver guard:

```go
// MethodA and MethodB both had `if m == nil { return nil }`:
return a?.MethodA()?.MethodB()

// Only MethodB had the guard:
return a.MethodA()?.MethodB()
```

See [`?.` / `??`](nilable_types.md) and [`*T?`](nilable_pointer_types.md). modernize applies these rewrites only for methods with nil-receiver guards (`optional_method_chains`, default on).

### Explicit checks

When you need a custom error instead of a panic or optional short-circuit:

```go
c := mgr.Connection(host)
if c == nil {
	return errors.New("unable to find connection for %s", host)
}
return c.Subroute(path)
```

## Relationship to nilable pointers

- **No nil receivers** — pointer method calls panic at the call site.
- **`*T` / `*T?`** — compile-time: which pointer types may hold `nil`.

A function may return `*Connection?` when nil means “not found”. Callers choose:

- `c.Subroute(p)` — panic if `c` is nil (strict call),
- `c?.Subroute(p)` — skip call, propagate nil,
- `if c == nil { … }` — custom handling.

Do **not** use `return new(T)` or `return nil` inside methods to mean “receiver was nil” — that path is unreachable for direct calls in Bow.

## Related docs

- [Panics and stack traces](panics.md)
- [Nilable pointer types](nilable_pointer_types.md)
- [Nilable types (`T?`, `?.`)](nilable_types.md)
- [modernize `remove_nil_receiver_guards`](https://github.com/Bow5/modernize)
