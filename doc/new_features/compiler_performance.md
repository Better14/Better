# Compiler performance

This document describes **quadratic or otherwise expensive lookup paths** introduced in the fork’s type checker (`cmd/compile/internal/types2`) by new language features. It is aimed at compiler maintainers and anyone debugging slow compiles.

Upstream Go already has a few intentionally quadratic algorithms (switch duplicate detection, union term lists, unification stacks, inference). Those are noted at the end for context but are **not** fork additions.

---

## Fixed: binary operator overload on long chains (no overloads)

### Symptom

Compiling packages with very long **left-associative binary operator chains** could appear to hang. A concrete example was `vendor/golang.org/x/text/unicode/norm` (`tables17.0.0.go`), which builds a large string via hundreds of `+` concatenations in `recompMapPacked`.

### Cause

For every binary operation, `binary()` in `expr.go` called `tryBinaryOperatorOverload` **before** the normal path. The probe type-checked both operands with `check.expr`, even when the package defined **no** overload for that operator. When overload resolution failed, `binary()` type-checked the same operands again.

On a chain of length **n**, each `+` node re-walked the entire left subtree. Total work was **O(n²)** in the number of operators.

### Fix

`tryBinaryOperatorOverload` and `tryUnaryOperatorOverload` return immediately when `len(check.operatorFuncs(name)) == 0`, so packages without operator overloads pay no probe cost.

**Location:** `types2/operator.go`

---

## Still present: binary operator overload when overloads exist

If a package defines **any** overload for an operator (e.g. `func +(l, r *Matrix) *Matrix`), the probe still runs on **every** use of that operator in the package.

When resolution **fails** (operands do not match any overload — e.g. built-in `string + string` in a package that also defines `+` on a custom type), the operands are type-checked twice:

1. `tryBinaryOperatorOverload` — `check.expr` on `lhs` and `rhs`
2. `binary()` — `check.expr` on `lhs` and `rhs` again after the probe returns `false`

On a left-associative chain of length **n**, this is again **O(n²)** in expression size.

**Locations:** `types2/operator.go` (`tryBinaryOperatorOverload`), `types2/expr.go` (`binary`)

**Mitigation ideas (not implemented):** defer operand type-checking in the probe until an overload candidate is known to apply; or pass pre-checked operands from `binary()` into the probe so subtrees are never walked twice.

---

## O(k²) at package init: overload duplicate detection

When overload suffixes are assigned, `checkOverloadDuplicates` compares every pair of candidates with the same name — **O(k²)** where **k** is the number of overloads for that name.

In practice **k** is tiny (handful of overloads per operator or method), so this is not a compile-time concern for normal code.

**Location:** `types2/overload.go` (`checkOverloadDuplicates`, called from `assignOverloadSuffixes`)

---

## O(V²) or O(cases × V): enum variant lookup

Enum variants are resolved by **linear scan** over `enumTyp.variants`:

- `enumVariantByName` — match by variant name
- `enumVariantByObj` — match by object identity

`enumCasePattern` calls `enumVariantByObj` once per switch case. With **V** variants and **C** case patterns, worst-case cost is **O(C × V)**; if **C ≈ V**, that is **O(V²)**.

Enums are expected to have a modest number of variants, so this is usually negligible. A map from name or object to variant would make lookup **O(1)** per case.

**Location:** `types2/enum.go`

---

## Linear scans per use site (not O(n²) in file size)

These paths are fork additions but scale with **scope or import count**, not with nested expression depth. They can still matter in large packages with many selector calls.

### Extension method resolution

`extensionCandidates` scans **all names** in the current package scope and in **every imported** package scope for each extension call probe. It does not use a method-name index.

Cost per selector call: **O(|imports| × |scope|)** in the worst case.

**Location:** `types2/extension.go` (`extensionCandidates`, called from `tryExtensionCall`)

Early exits in `tryExtensionCall` (skip package-qualified calls, skip when an instance method or func-typed field exists) reduce how often the full scan runs.

### Index operator fallback

`operatorFuncsForRecv` looks up `[]` and `[]=` overloads on a receiver type. When the per-package map misses, `operatorFuncsInPackage` scans all of `pkg.scope.Names()`.

Cost: **O(|scope|)** per lookup.

**Location:** `types2/operator.go` (`operatorFuncsInPackage`)

### Function and method overload selection

`selectOverloadEx` and `selectOperatorFunc` try each overload candidate against each argument — **O(k × args)** per call, where **k** is the number of candidates.

**Location:** `types2/call.go`, `types2/operator.go`

### Index operator probe: duplicate evaluation of base

`indexExpr` already type-checks `e.X` via `check.exprOrType`. If builtin indexing does not apply, `tryIndexOperatorOverload` type-checks `e.X` again with `check.expr`.

This is **duplicate work per index expression**, not quadratic in chain length unless many nested custom `[]` operators force repeated re-evaluation of the same subtrees.

**Location:** `types2/index.go`, `types2/operator.go` (`tryIndexOperatorOverload`)

---

## Summary table

| Path | Complexity | When it matters | Status |
| ---- | ---------- | --------------- | ------ |
| Binary operator probe, no overloads in package | O(n²) on operator chain length **n** | Long `+` chains (e.g. string tables) | **Fixed** |
| Binary operator probe, overloads declared | O(n²) when probe fails, then builtin path runs | Packages with `+` overloads and many builtin `+` uses | **Open** |
| Overload duplicate check | O(k²), small **k** | Package init only | Acceptable |
| Enum variant by name/obj | O(V) per lookup; O(C × V) per switch | Large enums with many cases | Acceptable today |
| Extension candidates | O(\|imports\| × \|scope\|) per probe | Many extension calls in huge scopes | Linear scan |
| `operatorFuncsInPackage` fallback | O(\|scope\|) | Custom `[]` on types from other packages | Linear scan |
| Overload selection | O(k × args) | Many overloads per name | Per call |
| Index operator re-check of base | 2× work on `e.X` | Custom `[]` on non-builtin types | Per index |

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
