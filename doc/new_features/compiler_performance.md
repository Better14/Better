# Compiler performance

This document describes **quadratic or otherwise expensive lookup paths** introduced in the fork’s type checker (`cmd/compile/internal/types2`) by new language features, and the **dictionary indexes** added to fix them. It is aimed at compiler maintainers and anyone debugging slow compiles.

Upstream Go already has a few intentionally quadratic algorithms (switch duplicate detection, union term lists, unification stacks, inference). Those are noted at the end for context but are **not** fork additions.

---

## Index overview

Fork feature lookups that previously scanned lists or scopes now use maps built once per package (or once per type-check pass) and reused across files.

| Index | Key | Value | Built |
| ----- | --- | ----- | ----- |
| `operatorExact` | operator name → `(leftType, rightType)` | `*Func` | `assignOverloadSuffixes` / lazy for imports |
| `operatorUnaryExact` | operator name → operand type | `*Func` | same |
| `overloadBySig` | `name + "·" + paramTypeSuffix` | `*Func` | same |
| `operatorFuncIndex` | operator / overload base name | `[]*Func` | lazy per imported package |
| `extensionByName` | method name | `[]*Func` | lazy per import; incremental in `indexExtensionFunc` |
| `variantByName` / `variantByObj` | variant name / `Object` | `*EnumVariant` | `NewEnum` |
| `overloadResolveCache` | `(cand, nargs, argTypeSuffix)` | `*Func` | per file check, on first unambiguous resolution |

Implementation lives primarily in `types2/perf_index.go`. Package-level indexes are stored on `Package`; per-checker indexes and the resolve cache are on `Checker`.

---

## Fixed: binary operator overload on long chains (no overloads)

### Symptom

Compiling packages with very long **left-associative binary operator chains** could appear to hang. A concrete example was `vendor/golang.org/x/text/unicode/norm` (`tables17.0.0.go`), which builds a large string via hundreds of `+` concatenations in `recompMapPacked`.

### Cause

For every binary operation, `binary()` in `expr.go` called `tryBinaryOperatorOverload` **before** the normal path. The probe type-checked both operands with `check.expr`, even when the package defined **no** overload for that operator. When overload resolution failed, `binary()` type-checked the same operands again.

On a chain of length **n**, each `+` node re-walked the entire left subtree. Total work was **O(n²)** in the number of operators.

### Fix

`applyBinaryOperatorOverload` returns immediately when `len(check.operatorFuncs(name)) == 0`, so packages without operator overloads pay no probe cost.

**Location:** `types2/operator.go` (`applyBinaryOperatorOverload`)

---

## Fixed: binary operator overload when overloads exist

### Cause

When a package defined overloads for an operator, the old probe type-checked operands **before** `binary()` type-checked them again on failure — **O(n²)** on long chains of built-in uses (e.g. `string + string` alongside a custom `+` overload).

### Fix

1. **`binary()` evaluates operands once**, then calls `applyBinaryOperatorOverload` with the pre-checked operands (`types2/expr.go`).
2. **Exact-type index:** at package init, `buildCheckerIndexes` builds `operatorExact` and `operatorUnaryExact` (`map[operator][operandTypes]→*Func`) in `types2/perf_index.go`. Resolution tries the index first, then falls back to assignability scanning only when needed.

**Locations:** `types2/expr.go`, `types2/operator.go`, `types2/perf_index.go`

---

## Fixed: overload duplicate detection at package init

### Cause

`checkOverloadDuplicates` compared every pair of overload candidates with the same name — **O(k²)** where **k** is the number of overloads for that operator or method.

### Fix

`checkOverloadDuplicates` uses a map keyed by `name + "·" + overloadParamSuffix(sig)` — **O(k)** per overload name instead of **O(k²)** pairwise comparison.

**Location:** `types2/overload.go`

---

## Fixed: enum variant lookup by linear scan

### Cause

`enumVariantByName` and `enumVariantByObj` scanned the `variants` slice — **O(V)** per lookup, **O(C × V)** per enum switch with **C** cases.

### Fix

`NewEnum` builds `variantByName` and `variantByObj` maps. `enumVariantByName` and `enumVariantByObj` use **O(1)** map lookup.

**Location:** `types2/enum.go`

---

## Fixed: extension method candidate lookup

### Cause

`extensionCandidates` scanned every name in the current package scope and in each imported package scope for every extension call — **O(|imports| × |scope|)** per probe.

### Fix

Each `Package` keeps `extensionByName map[string][]*Func`, built once per package (lazily for imports, incrementally via `indexExtensionFunc` when extensions are lowered). `extensionCandidates` looks up by method name instead of scanning entire scopes.

**Location:** `types2/perf_index.go`, `types2/extension.go`

---

## Fixed: operator func lookup in imported packages

### Cause

When `pkg.overloadFuncs[name]` missed, `operatorFuncsInPackage` scanned all of `pkg.scope.Names()` on every lookup — **O(|scope|)** per call.

### Fix

`operatorFuncsInPackage` uses `pkg.operatorFuncIndex`, populated once from `overloadFuncs` plus a single scope scan for suffixed decls (`name·suffix`), instead of scanning scope on every lookup.

**Location:** `types2/perf_index.go`, `types2/operator.go`

---

## Fixed: overload selection by argument types

### Cause

`selectOverloadEx` and `selectOperatorFunc` tried each overload candidate against each argument — **O(k × args)** per call with no memoization.

### Fix

At package init, `buildCheckerIndexes` builds `overloadBySig map[name+paramTypes]→*Func` for package-level and method overloads. `selectOperatorFunc`, `selectOverloadEx`, and operator overload resolution consult this index before linear assignability scans.

`selectOverloadEx` also memoizes unambiguous results in `overloadResolveCache`.

**Location:** `types2/perf_index.go`, `types2/call.go`, `types2/operator.go`

---

## Fixed: index operator duplicate evaluation of base

### Cause

`indexExpr` already type-checked `e.X`, but `tryIndexOperatorOverload` called `check.expr` on `e.X` again.

### Fix

`indexExpr` passes the already type-checked receiver operand into `tryIndexOperatorOverload`, avoiding a second `check.expr` on `e.X`.

**Location:** `types2/index.go`, `types2/operator.go`

---

## Summary table

| Path | Complexity (before) | Status |
| ---- | ------------------- | ------ |
| Binary operator probe, no overloads in package | O(n²) on chain length | **Fixed** |
| Binary operator probe, overloads declared | O(n²) when probe fails | **Fixed** |
| Overload duplicate check | O(k²) | **Fixed** (O(k)) |
| Enum variant by name/obj | O(V) per lookup | **Fixed** (O(1)) |
| Extension candidates | O(\|imports\| × \|scope\|) | **Fixed** (O(\|imports\|)) |
| `operatorFuncsInPackage` fallback | O(\|scope\|) per lookup | **Fixed** (O(1) after index) |
| Overload selection | O(k × args) per call | **Indexed** + memoized |
| Index operator re-check of base | 2× work on `e.X` | **Fixed** |

---

## Upstream quadratic algorithms (not fork additions)

The fork inherits several algorithms that are already documented as quadratic in upstream `types2`:

| Area | File | Comment in code |
| ---- | ---- | --------------- |
| Duplicate case values in expression switch | `stmt.go` | “quadratic algorithm, but these lists tend to be very short” |
| Duplicate types in type switch | `stmt.go` | “quadratic algorithm, but type switches tend to be reasonably small” |
| Union term lists | `union.go` | “quadratic termlist operations”; “unions tend to be short” |
| Unification stacks | `unify.go`, `predicates.go` | “quadratic algorithm, but in practice these stacks …” |
| Type inference | `infer.go` | “O(n²) algorithm where n is the number of …” |

These are unchanged by the fork features above unless a new code path invokes them more often.

---

## Related reading

- [Operator overloading](operator_overloading.md) — declaration and resolution rules
- [Function and method overloading](overloading.md) — overload selection at call sites
- [Extension methods](extension_methods.md) — extension resolution
- [Enums](enums.md) — variants and exhaustive switching
