# Go Language Extensions

This directory documents language and library extensions planned or implemented in this fork. Each major feature has its own file.

## Type system

- [Result types (`T!`)](result_types.md) — `(T, error)` shorthand and `!.value` / `!.field` error propagation
- [Nilable types (`T?`)](nilable_types.md) — optional values, `?.`, and `??`
- [Nilable pointer types (`*T` / `*T?`)](nilable_pointer_types.md) — non-nilable vs nilable pointers; `nilable_pointers` in go.mod and `//go:nilable_pointers` regions
- [No nil receivers](nil_receivers.md) — pointer method calls panic at the call site; use `?.` for optional chains
- [Enums](enums.md) — algebraic enums, variants, and exhaustive switching
- [Operator overloading](operator_overloading.md) — user-defined operators via operator methods

## Functions and syntax

- [Function and method overloading](overloading.md)
- [If and switch expressions](expressions.md)
- [Lambda syntax (`=>`)](lambda_syntax.md)
- [Default function arguments](default_arguments.md)
- [Struct and interface shorthand (`struct T { … }`, `interface I { … }`)](syntax.md)
- [Panics and stack traces](panics.md) — stderr traceback on unrecovered panic; structured error panics

## Generics and methods

- [Generic methods (Go 1.27)](generic_methods.md)
- [Extension methods](extension_methods.md)

## Standard library

- [Library changes](library_changes.md) — stdlib API extensions (e.g. .NET `FormatCustom` date/time)
- [Structured errors (`errors.Error`)](errors.md) — message, stack trace, inner error chain; **most stdlib custom error types migrated**
- [Built-in LINQ](linq.md)
- [Data structures](data_structures.md) — `list`, `set`, `queue`, `stack`, heaps, trees

## Performance

- [Compiler performance](compiler_performance.md) — Changes to prevent O(n^2) lookups and expensive type-checker paths in fork features

## Tooling

- [gopls (IDE support)](gopls.md) — what the language server supports vs the compiler-only path

## Review backlog

- [Weird upstream Go behaviors](weird_behaviors.md) — surprising semantics we may want to change or lint (nil interfaces, maps, defer, etc.)

## Quick reference

| Syntax | Document |
| ------ | -------- |
| `int!`, `expr!.value` | [result_types.md](result_types.md) |
| `int?`, `expr?.field`, `expr ?? fallback` | [nilable_types.md](nilable_types.md) |
| `*T`, `*T?` | [nilable_pointer_types.md](nilable_pointer_types.md) |
| `nilable_pointers enable` in go.mod | [nilable_pointer_types.md](nilable_pointer_types.md#project-default-nilable_pointers-in-gomod) |
| `//go:nilable_pointers enable` / `end` | [nilable_pointer_types.md](nilable_pointer_types.md#file-overrides-gonilable_pointers) |
| `enum E { … }` | [enums.md](enums.md) |
| `func f(a int, b int = 1)` | [default_arguments.md](default_arguments.md) |
| `(x, y) => x + y` | [lambda_syntax.md](lambda_syntax.md) |
| `nums.Where(…).Select(…)` | [linq.md](linq.md) |
| `errors.New`, `err.Wrap`, `*errors.Error` | [errors.md](errors.md) |
| `t.FormatCustom("yyyy-MM-dd hh:mm:ss tt")`, `time.ParseCustom(…)` | [library_changes.md](library_changes.md) |
| `func (p Person) Hello()` (foreign receiver) | [extension_methods.md](extension_methods.md) |
| `struct T { … }`, `interface I { … }` | [syntax.md](syntax.md) |
| `nil_receiver_panic enable` in go.mod | [nil_receivers.md](nil_receivers.md) |
| `obj?.method()`, `inner()?.Method()` | [nil_receivers.md](nil_receivers.md), [nilable_types.md](nilable_types.md) |
