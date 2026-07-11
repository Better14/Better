# Weird upstream Go behaviors (review backlog)

Upstream Go has several surprising or non-standard behaviors. Some we have already changed (or partially addressed); the rest are candidates to look at for Bow.

This is a brainstorming list, not a commitment.

## Already addressed (or related)

| Behavior | Notes |
| -------- | ----- |
| Nil pointer receivers run methods | Call-site panic — [nil_receivers.md](nil_receivers.md) |
| Nil channel receive blocks forever | Receive-site panic — [fixed_weird_behaviors.md](fixed_weird_behaviors.md) |
| Silent nil chains / defensive nil-receiver guards | Prefer `?.` / `*T?` — [nilable types](nilable_types.md), [nilable pointers](nilable_pointer_types.md) |
| Weak panic diagnostics | [Panics and stack traces](panics.md) |
| `(T, error)` / nil-error footguns | [Result types](result_types.md), [structured errors](errors.md) |

## Nil / zero-value semantics

- **Typed nil in interfaces** — `var p *T; var i error = p` → `i != nil`. Classic `return err` bug when `err` is a nil concrete pointer.
- **Nil map: read OK, write panics** — asymmetric; easy to forget initialization.
- **Nil slice** — `len` / `cap` / `range` / `append` fine; indexing panics.
- **Nil channel in `select`** — case is ignored (useful but surprising); direct receive panics in Bow — see [fixed_weird_behaviors.md](fixed_weird_behaviors.md).
- **`close(nil chan)` panics** — inconsistent with “nil channel is inert in select.”
- **Range over nil** slice/map/chan — zero iterations, no panic (usually fine; still surprising next to other nil panics).

## Interfaces and equality

- **Interface equality** depends on dynamic type *and* value; two different typed nils are not equal.
- **Structs with slices/maps/funcs are not comparable** even when those fields are nil.
- **`NaN` map keys** — insertable, often un-look-up-able (`NaN != NaN`).
- **Comparing interfaces that hold floats** — NaN makes equality weird.

## Control flow

- **`defer` arguments evaluated immediately**, body later — easy to capture the wrong loop variable / value (loop-var issue mostly fixed in Go 1.22+; defer timing remains).
- **`defer` can mutate named return values** — powerful and easy to misuse.
- **`switch` with no expression** — sugar for if/else; fine, but unusual vs other languages.
- **`fallthrough` ignores the next case’s condition** — always falls into the next body.
- **Labeled `break` / `continue`** out of `select` / `switch` — rarely used, easy to misread.

## Memory and representation

- **`append` may or may not reallocate** — aliasing after append is a classic footgun.
- **`string` ↔ `[]byte` conversion copies**; indexing a string yields bytes, ranging yields runes.
- **Empty `struct{}` is size 0** — many can share one address.
- **Copying `sync.Mutex` (or structs containing one) by value** — compiles; locking then silently wrong.
- **Method values capture the receiver** — `f := p.M` keeps `p` alive / copies value receivers.

## Concurrency

- **Starting a goroutine establishes no happens-before** by itself.
- **Concurrent map write → panic** (detected); data races elsewhere are undefined.
- **Send on closed channel panics**; receive from closed returns zero + `ok == false`.
- **No built-in cancellation of a plain goroutine** — must use context/channels by convention.

## Types and constants

- **Untyped constants** vs typed values — shift/overflow rules differ and surprise people.
- **Integer division truncates toward zero**.
- **Implicit numeric conversions are very limited** — often good; still a frequent complaint vs other languages.
- **Unused local / import is a compile error** — intentional strictness; sometimes annoying in WIP code.

## Possible directions (open questions)

For each item above, options roughly look like:

1. **Keep upstream semantics** (document only).
2. **Lint / modernize rewrite** (no language change).
3. **Per-file `//go:` directive** (like `//go:nilable_pointers`).
4. **Hard language change** in Bow (with migration path).

Highest-impact leftovers to consider first:

1. Typed nil in interfaces / error returns  
2. Nil map write vs read asymmetry (or required make)  
3. `defer` + named returns clarity  
4. `append` aliasing / copy-on-grow warnings  
5. Mutex-by-value detection (compiler or vet)

## Related docs

- [No nil receivers](nil_receivers.md)
- [Fixed weird behaviors](fixed_weird_behaviors.md)
- [Nilable types](nilable_types.md)
- [Nilable pointer types](nilable_pointer_types.md)
- [Result types](result_types.md)
- [Panics and stack traces](panics.md)
