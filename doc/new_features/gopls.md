# gopls (IDE support)

[gopls](https://pkg.go.dev/golang.org/x/tools/gopls) is the Go language server used by VS Code, Cursor, and other editors. In this fork it lives in a separate repository (`go_tools`) and must be built against this fork’s **GOROOT** so it uses the extended **`go/parser`**, **`go/ast`**, and **`go/types`** packages.

The **compiler** (`cmd/compile`) uses a different stack: **`cmd/compile/internal/syntax`** → **`types2`**. Features that exist only there compile fine but may not appear correctly in the IDE until they are ported to `go/types` (and to `go/parser` / `go/ast` where new syntax is needed).

---

## Building gopls

Use a **GOROOT** pointing at this fork and a **GOPATH separate from GOROOT**:

```bash
export GOROOT=/path/to/fork          # e.g. /root/go
export GOPATH=/path/to/gopath        # must not equal GOROOT
export GOEXPERIMENT=genericmethods
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

cd /path/to/go_tools/gopls
go build -o $GOPATH/bin/gopls .
```

Configure your editor to run this `gopls` binary and to set **`GOROOT`** (and **`GOEXPERIMENT=genericmethods`** if needed) for the language server process.

The fork’s **`go`** tool must be built first (`src/make.bash` in GOROOT).

---

## Two type-checking stacks

| Stack | Parser | Type checker | Used by |
| ----- | ------ | ------------ | ------- |
| **Compiler** | `cmd/compile/internal/syntax` | `cmd/compile/internal/types2` | `compile`, `go build` |
| **IDE** | `go/parser` → `go/ast` | `go/types` | gopls |

gopls type-checks via `types.NewChecker` and reads overload metadata from `types.Info` (`FuncOverloads`, `MethodOverloads`, `CallOverloads`). It does **not** call `types2` directly.

---

## Feature support in gopls

### Supported (parse + typecheck)

These are implemented in **`go/parser`**, **`go/ast`**, and **`go/types`**. gopls can parse, type-check, and provide diagnostics/completion for them when built with this GOROOT.

| Feature | AST / API | Document |
| ------- | --------- | -------- |
| Result types `T!` | `ast.ResultTypeExpr` | [result_types.md](result_types.md) |
| Error propagation `expr!.value`, `expr!.field` | `ast.TryExpr` | [result_types.md](result_types.md) |
| If expressions | `ast.IfExpr` | [expressions.md](expressions.md) |
| Switch expressions | `ast.SwitchExpr` | [expressions.md](expressions.md) |
| Lambda `=>` | `ast.LambdaExpr` | [lambda_syntax.md](lambda_syntax.md) |
| Default function arguments | `ast.Field.Default` | [default_arguments.md](default_arguments.md) |
| Function and method overloading | `types.Info.FuncOverloads`, `MethodOverloads`, `CallOverloads` | [overloading.md](overloading.md) |

**Overload signature help:** when a call has multiple overload candidates, gopls lists all matching signatures (see `gopls/internal/golang/signature_help.go` and marker test `testdata/signature/overload.txt` in `go_tools`).

**Tests in `go_tools`:** `gopls/internal/cache/parsego/parse_test.go` (syntax nodes), `gopls/internal/cache/overload_test.go` (overload typecheck).

### Not supported in gopls (compiler-only today)

These are implemented in **`cmd/compile/internal/syntax`** and **`types2`** but are **not** in `go/parser` / `go/types`. gopls will report **parse or type errors** even when `go build` succeeds.

| Feature | Typical gopls behavior | Document |
| ------- | ---------------------- | -------- |
| Enums `enum E { … }` | `expected declaration, found enum` | [enums.md](enums.md) |
| Nullable types `T?`, `?.`, `??` | not parsed | [nullable_types.md](nullable_types.md) |
| Extension methods | ordinary method lookup only | [extension_methods.md](extension_methods.md) |
| Operator overloading | built-in operators only | [operator_overloading.md](operator_overloading.md) |

**Standard library extensions** ([LINQ](linq.md), [structured errors](errors.md), [data structures](data_structures.md), etc.) compile with the fork toolchain; IDE support is the same as for normal Go packages once the **language** syntax type-checks. No separate gopls plugin is required for stdlib APIs.

---

## Summary table

| Syntax / feature | `go build` | gopls parse | gopls typecheck |
| ---------------- | ---------- | ----------- | --------------- |
| `int!`, `expr!.value` | yes | yes | yes |
| `if` / `switch` expressions | yes | yes | yes |
| `(a, b) => expr` | yes | yes | yes |
| Default args | yes | yes | yes |
| Overloading | yes | yes | yes (+ signature help) |
| `enum E { … }` | yes | no | no |
| `T?`, `?.`, `??` | yes | no | no |
| Extension methods | yes | yes* | no |
| Operator overloading | yes | yes* | no |

\*Source that uses only standard Go syntax may parse; fork-specific resolution does not run in gopls.

---

## Porting work (future)

To get full IDE support for compiler-only features:

1. **`go/ast` + `go/parser`** — new nodes and parsing for enum, nullable types, etc.
2. **`go/types`** — port checking logic from `types2` (extensions, operators, enums, optionals).
3. **`go_tools` / gopls** — semantic tokens, completion, and analyzers as needed.

Until then, use **`go build`** / **`go test`** as the source of truth for compiler-only syntax.

---

## Related reading

- [new_features.md](new_features.md) — index of language extensions
- [Compiler performance](compiler_performance.md) — type checker optimizations in `types2` (compiler path)
