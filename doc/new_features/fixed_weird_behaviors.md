# Fixed upstream weird behaviors

Bow deliberately changes some surprising upstream Go behaviors. Each item below documents what upstream did, what Bow does instead, and how to migrate.

## Nil channel receive blocks forever → panic

### Upstream

Receiving from a `nil` channel **blocks forever**:

```go
var ch chan int
<-ch        // hangs
x := <-ch   // hangs
x, ok := <-ch // hangs
for v := range ch { _ = v } // hangs on first receive
```

This is easy to mistake for a deadlock. A `nil` channel in `select` is silently ignored (the case never runs), which is inconsistent with direct receive.

### Bow

Any **executed** receive from a `nil` channel **panics** at the receive site with a nil exception (`invalid memory address or nil pointer dereference`) and a stack trace (see [panics.md](panics.md)).

```go
var ch chan int
<-ch // panics here — does not block
```

| Form | Bow behavior |
| ---- | ------------ |
| `<-ch` | Panic if `ch` is nil |
| `x = <-ch` | Panic if `ch` is nil |
| `x, ok = <-ch` | Panic if `ch` is nil |
| `for v := range ch` | Panic on first receive if `ch` is nil |
| `select { case x := <-ch: … }` | Unchanged: nil channel cases are never selected (no receive runs) |
| Send on nil channel | Unchanged: still blocks forever |
| `close(nil)` on strict `chan T` | Unchanged: still panics at run time |
| `close(ch)` on nilable `chan T?` without nil check | Compile error with NPT `enable` (see [Nilable pointer types](nilable_pointer_types.md)) |

The compiler inserts `OCHECKNIL` on the channel expression before lowering to `chanrecv1` / `chanrecv2`. There is no `go.mod` toggle — this is always on in Bow.

### Migration

Use an explicit nil check or optional channel type when absence is expected:

```go
// NPT enabled: declare nilable channels where nil is valid
var ch chan int?
if ch != nil {
	v := <-ch // narrowed to chan int in then-branch
	_ = v
}

// Or guard explicitly
if ch == nil {
	return errors.New("no channel")
}
v := <-ch

// close requires a nil check on nilable channels
if ch != nil {
	close(ch)
}
```

With [nilable pointer types](nilable_pointer_types.md), `chan T` is non-nilable under NPT; use `chan T?` when the channel itself may be absent.

## Related docs

- [Weird behaviors backlog](weird_behaviors.md) — candidates not yet changed
- [Nilable pointer types](nilable_pointer_types.md) — `chan T` vs `chan T?`
- [Panics and stack traces](panics.md)
- [No nil receivers](nil_receivers.md)
