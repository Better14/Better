# Interface == nil and typed nil

Bow changes interface nil comparisons so a **typed nil** stored in an interface compares **equal to `nil`**.

## Upstream

Assigning a typed nil pointer to an interface sets the interface’s type slot but leaves the data word nil:

```go
var p *MyError = nil
var err error = p
fmt.Println(err == nil) // false in upstream Go
```

Code that relied on `err != nil` for a typed nil concrete pointer misbehaves — the classic `return err` footgun when the returned `error` is a nil `*MyType`.

## Bow

`iface == nil` and `iface != nil` treat an interface as nil when **either** the type slot is unset **or** the data word is nil (typed nil):

```go
var p *T = nil
var i SomeInterface = p
fmt.Println(i == nil) // true
```

The compiler lowers comparisons to `tab == nil || data == nil` (and the dual for `!= nil`).

## Migration

Search for interface-typed `== nil` / `!= nil` checks. After this change:

- Branches that assumed typed nil is “non-nil” may now take the nil path.
- `if err != nil` after `return err` from a function that returns a typed nil pointer now behaves as many authors expected.

[modernize](https://github.com/Bow5/modernize) labels candidate sites with:

```go
//FIXME: Make sure still works after interface == nil change.
```

Review each labeled comparison manually; the tool does not rewrite logic.

## Related docs

- [Fixed weird behaviors](fixed_weird_behaviors.md)
- [Weird behaviors backlog](weird_behaviors.md)
