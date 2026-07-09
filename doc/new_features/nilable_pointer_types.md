# Nilable Pointer Types

**Proposed.** Not implemented yet.

This document specifies *nilable pointer types* (NPT): a compile-time null-safety layer for Go pointer types. Runtime behavior is unchanged; the feature is entirely static analysis plus diagnostics.

NPT is separate from but complementary to [nilable value types (`T?`)](nilable_types.md).

## Setting

NPT is controlled by a `nilable_pointers` directive in `go.mod`:

```go
module example.com/myapp

go 1.27

nilable_pointers enable   // or: disable | warn
```

| Value | Meaning |
| ----- | ------- |
| `disable` | Legacy behavior; `*T` may be `nil`; `*T?` not used (default for existing modules) |
| `warn` | NPT on; all violations are warnings |
| `enable` | NPT on; **definite** violations are compile errors; **flow-analysis** violations are warnings (see [Diagnostics](#diagnostics)) |

### Severity under `enable`

Not every NPT diagnostic is equally certain. Under `enable`:

- **Compile errors** — violations the compiler knows are wrong without control-flow inference, such as assigning or returning `nil` for a non-nilable `*T`.
- **Warnings** — violations that depend on null-state analysis across branches, calls, or initialization (e.g. dereferencing a `*T?` without a check, passing `*T?` where `*T` is required). These may include false positives or require refactors the analyzer cannot prove are safe.

**Eventual goal:** as null-state analysis matures, more diagnostics move from warnings to compile errors under `enable`, until `enable` treats all NPT violations as compile errors. Until then, `warn` remains the migration mode where everything is a warning.

`go mod tidy` and module graph tools should read this directive so builds are reproducible and importers know which nullability rules apply.

File- or package-level overrides are supported for migration:

```go
//go:nilable_pointers disable
```

Precedence: **file > package > go.mod**.

## Overview

Today every pointer can be `nil`, and the compiler does not distinguish “this must point to something” from “this might be absent.” When `nilable_pointers` is `warn` or `enable`, NPT lets callers and implementers state that intent in the type system and get warnings or errors when code violates it.

| Today (`disable`) | With NPT on (`warn` / `enable`) |
| ----------------- | ------------------------------- |
| `var a *MyStruct` — may be `nil`; no annotation | `var a *MyStruct` — **must not** be `nil` |
| (same syntax) | `var a *MyStruct?` — **may** be `nil` |

When the context is **disabled** (default), pointer types behave exactly as in current Go: every `*T` may be `nil`, and `*T?` is not a distinct type (or is rejected as redundant).

When the context is **enabled** (`warn` or `enable`), pointer annotations apply:

- `*T` — non-nilable pointer; assigning or passing `nil` where `*T` is expected is a compile error under `enable` (a warning under `warn`).
- `*T?` — nilable pointer; holding `nil` is allowed; dereferencing without a nil check is a warning under `enable` (see [Diagnostics](#diagnostics)).

The `?` suffix attaches to the pointer type as a whole (`*MyStruct?`), consistent with `int?` for value types. It is not the same as `*int?` (pointer to nilable `int`), which remains “pointer to `int?`” when both features are in use.

`T?` adds optional values for types that cannot hold `nil` today (e.g. `int?`). NPT adds the opposite annotation for types that *can* hold `nil` today (pointers, and optionally other reference kinds): express which pointers **must not** be `nil` and which **may** be.

### Scope (initial proposal)

**In scope for v1:**

- Plain pointers: `*T`, `*T?`
- Pointer fields in struct definitions
- Function parameters and results
- Local variables and assignments
- Method receivers (`func (p *MyStruct) …` vs `func (p *MyStruct?) …`)

**Out of scope for v1 (may follow later):**

- Slices, maps, channels, functions, and interfaces as separate nilable/non-nilable reference kinds (more surface area; may follow later)
- Changing the meaning of `nil` at runtime
- Automatic insertion of nil checks in generated code

## Express intent with annotations

### Declarations

```go
// NPT enabled in this module (nilable_pointers enable or warn)

var required *MyStruct = &MyStruct{} // ok
var required *MyStruct = nil         // compile error with enable; warning with warn

var optional *MyStruct? = nil        // ok
var optional *MyStruct? = &MyStruct{} // ok
```

### Struct fields

```go
struct Person {
	Name       string
	MiddleName *string?  // optional middle name
	Parent     *Person   // every person has a parent pointer (non-null by contract)
}
```

Implementations must respect the annotations: use `MiddleName` only after a nil check; use `Parent` directly when the field is non-nilable.

### Function signatures

```go
func Find(id int) *User? {
	// may return nil when not found
}

func MustFind(id int) *User {
	// caller expects non-nil; returning nil is a compile error with enable, warning with warn
}
```

Call sites:

```go
u := Find(1)
if u != nil {
	fmt.Println(u.Name) // ok: u is *User in this branch
}

m := MustFind(1)
fmt.Println(m.Name)   // ok: m is known non-null when Find succeeds
```

Non-nilable parameters reject `nil` at the call site:

```go
func Process(u *User) { … }

var maybe *User?
Process(maybe)        // warning: *User? not assignable to *User without unwrap

if maybe != nil {
	Process(maybe)    // ok in then-branch: maybe is *User
}
```

## Null-state analysis

The compiler tracks a *null-state* for each pointer-typed expression:

- **not-null** — known not `nil`
- **maybe-null** — might be `nil`

Two mechanisms update state: **assignments** and **nil checks**.

```go
var p *MyStruct? = lookup()

// warning: dereference of possibly nil pointer
_ = p.Field

p = &MyStruct{}
_ = p.Field // ok: p is not-null after non-nil assignment

if p != nil {
	_ = p.Field // ok: not-null in then-branch
}
```

Analysis follows control flow: `if`, `for`, `switch`, early `return`, and [pattern matching](expressions.md) where applicable. It does not trace into arbitrary function bodies unless contracts are described (see [API contracts](#api-contracts) below).

### Relationship to `?.` and `??`

Existing [nilable operators](nilable_types.md) apply to NPT:

- `p?.Field` — safe when `p` is `*T?`; short-circuits on `nil`
- `p ?? fallback` — when `p` is `*T?`, use `p` if non-`nil`, else `fallback`

For non-nilable `*T`, `?.` is unnecessary (dereference is always allowed by annotation); the compiler may warn on redundant `?.`.

## Assignability: `*T` is not `*T?`

`*T` and `*T?` are distinct types in an enabled nilable context, similar to `T` vs `T?` and `T` vs `T!`.

| From | To | Allowed |
| ---- | -- | ------- |
| `*T?` | `*T?` | yes |
| `*T` | `*T?` | yes (non-nilable is a subset) |
| `*T?` | `*T` | no, unless proven not-null (branch, unwrap) |
| `*T` | `*T` | yes |
| `nil` | `*T?` | yes |
| `nil` | `*T` | compile error with `enable`; warning with `warn` |

Unwrapping `*T?` to `*T`:

**Nil check** — use in a branch where analysis proves not-null:

```go
if p != nil {
	useRequired(p) // p is *MyStruct here
}
```

**Explicit non-null assertion** — when the programmer knows more than the analyzer (see [Null-forgiving](#null-forgiving) below).

**Null coalescing** — supply a non-nil fallback:

```go
useRequired(p ?? &MyStruct{})
```

## Null-forgiving

Sometimes a value is provably non-`nil` to the author but not to the compiler (e.g. after a map lookup keyed by convention, or immediately after `new`). This fork already uses `!` for [result types](result_types.md), so NPT needs a distinct spelling for null-forgiving assertions.

**Proposed:** reuse the force-unwrap cast pattern from nilable value types:

```go
var p *MyStruct? = lookup()
useRequired((*MyStruct)(p)) // panics at run time if p == nil; silences the NPT diagnostic
```

Alternatively, a dedicated null-forgiving operator may be added later if a lighter spelling is needed; it must not collide with `expr!` / `T!`.

Use non-null assertions sparingly; each one is a place the compiler no longer protects you.

## API contracts

Parameter and return annotations are not always enough. A helper may accept `*T?` but guarantee a non-null result when the input is non-null. Optional analysis attributes on stdlib and user APIs (e.g. indicating that an argument is not-null when a predicate returns true) can describe these contracts. Exact attribute names and placement are TBD; the compiler honors them for flow analysis across calls.

## Diagnostics

When NPT is on (`nilable_pointers` is `warn` or `enable`):

| Situation | Category | `warn` | `enable` (v1) | `enable` (goal) |
| --------- | -------- | ------ | ------------- | --------------- |
| Assign `nil` to `*T` | definite | warning | compile error | compile error |
| Return `nil` from function declared `*T` | definite | warning | compile error | compile error |
| Pass `nil` for a non-nilable `*T` parameter | definite | warning | compile error | compile error |
| Pass `*T?` where `*T` is required | flow | warning | warning | compile error |
| Dereference `*T?` without check | flow | warning | warning | compile error |
| Leave non-nilable struct field uninitialized | flow | warning | warning | compile error |

**Definite** violations do not depend on control-flow inference — the source itself assigns or returns `nil` for a non-nilable type. **Flow** violations depend on null-state analysis; they are warnings under `enable` in v1 because the analyzer may be incomplete or produce debatable results.

Over time, flow diagnostics should be promoted to compile errors under `enable` as analysis improves. `warn` stays available for modules that are not ready for any build failures.

**Runtime:** unchanged. A non-nilable `*T` that holds `nil` at run time still panics on dereference (or behaves as today); NPT does not insert checks.

## Known pitfalls

Static analysis has known blind spots:

### Default struct values

A struct with a non-nilable pointer field left at the zero value has `nil` in that field at run time without a warning if the struct itself is created with `var s MyStruct` or `new(MyStruct)`:

```go
struct Node {
	Next *Node // non-nilable by annotation
}

var n Node
_ = n.Next.Value // warning if analysis tracks fields; zero-value init may need `//go:nilable_pointers disable` or required-field syntax later
```

Mitigation options for a future revision: `required` field markers, constructors, or warnings on struct literals with missing non-nilable fields.

### Pointer arrays and slices

`make([]*MyStruct, n)` produces elements that are `nil` until assigned. Elements typed as `*MyStruct` in an enabled context should warn if used before initialization, or the element type should be `*MyStruct?` until proven set.

## Migration

1. Add `nilable_pointers disable` explicitly to existing `go.mod` files (or omit; default is disable).
2. Enable per module with warnings first: `nilable_pointers warn`.
3. Fix warnings module-by-module; use `//go:nilable_pointers disable` on generated or legacy files.
4. Tighten to `nilable_pointers enable` when definite violations (nil assignment/return) are clean; flow warnings may remain.
5. As analysis improves, `enable` will promote more flow diagnostics to compile errors without changing the directive name.
6. New modules created from a template default to `enable`.

Libraries consumed with NPT off keep today’s behavior. When both consumer and provider use NPT, exported APIs should annotate nilable vs non-nilable pointer parameters and results in docs and signatures.

## Interaction with other fork features

| Feature | Interaction |
| ------- | ----------- |
| [`T?` value types](nilable_types.md) | Orthogonal. `int?` vs `*int?` vs `*int` are three distinct concepts. |
| [`T!` result types](result_types.md) | `!` remains error propagation, not null-forgiving. |
| [`?.` / `??`](nilable_types.md) | Apply to `*T?`; encouraged over raw dereference. |
| [Default arguments](default_arguments.md) | Non-nilable `*T` parameters cannot default to `nil`; use `*T?` with default `nil` if optional. |
| [gopls](gopls.md) | Should surface null-state, annotations, and diagnostics in the IDE. |

## Summary

| Syntax | NPT off | NPT on |
| ------ | ------- | ------ |
| `*MyStruct` | may be `nil` | must not be `nil` (non-nilable) |
| `*MyStruct?` | N/A or same as `*MyStruct` | may be `nil` (nilable) |
| `p?.Field` | valid on any pointer | idiomatic for `*T?` |
| `p ?? fallback` | valid | unwrap `*T?` to `*T` with default |

NPT is compile-time only, opt-in via [`nilable_pointers` in `go.mod`](#setting), with optional file-level overrides for gradual migration.
