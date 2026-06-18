# Go Language Extensions

This directory documents language and library extensions planned or implemented in this fork. Each major feature has its own file.

## Type system

- [Result types (`T!`)](result_types.md) — `(T, error)` shorthand and `!.value` / `!.field` error propagation
- [Nullable types (`T?`)](nullable_types.md) — optional values, `?.`, and `??`
- [Nullable pointer types (`*T` / `*T?`)](nullable_pointer_types.md) — non-nullable vs nullable pointers (proposed)
- [Enums](enums.md) — algebraic enums, variants, and exhaustive switching
- [Operator overloading](operator_overloading.md) — user-defined operators via operator methods

## Functions and syntax

- [Function and method overloading](overloading.md)
- [If and switch expressions](expressions.md)
- [Lambda syntax (`=>`)](lambda_syntax.md)
- [Default function arguments](default_arguments.md)
- [Proposed syntax changes (not implemented)](syntax.md)

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

## Quick reference

| Syntax | Document |
| ------ | -------- |
| `int!`, `expr!.value` | [result_types.md](result_types.md) |
| `int?`, `expr?.field`, `expr ?? fallback` | [nullable_types.md](nullable_types.md) |
| `*T`, `*T?` (with `nullable_pointers` in go.mod) | [nullable_pointer_types.md](nullable_pointer_types.md) |
| `enum E { … }` | [enums.md](enums.md) |
| `func f(a int, b int = 1)` | [default_arguments.md](default_arguments.md) |
| `(x, y) => x + y` | [lambda_syntax.md](lambda_syntax.md) |
| `nums.Where(…).Select(…)` | [linq.md](linq.md) |
| `errors.New`, `err.Wrap`, `*errors.Error` | [errors.md](errors.md) |
| `t.FormatCustom("yyyy-MM-dd hh:mm:ss tt")`, `time.ParseCustom(…)` | [library_changes.md](library_changes.md) |
| `func (p Person) Hello()` (foreign receiver) | [extension_methods.md](extension_methods.md) |
| Proposed `struct T { … }`, `interface I { … }` (not implemented) | [syntax.md](syntax.md) |
