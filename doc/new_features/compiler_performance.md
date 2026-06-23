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
| `pkgHasCallOverloads` | — | `bool` | `assignOverloadSuffixes` (once per check) |
| `pkgHasOperatorOverloads` | — | `bool` | same |
| `pkgHasExtensions` | — | `bool` | same |
| `pkgHasEnums` | — | `bool` | same |
| `operatorOverloadsByName` | operator name | `[]*Func` | lazy when `pkgHasOperatorOverloads` |
| `extensionMethodByName` | method name | `bool` | lazy when `pkgHasExtensions` |

Implementation lives primarily in `types2/perf_index.go`. Package-level indexes are stored on `Package`; per-checker indexes, feature flags, and resolve caches are on `Checker`.

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

## Fixed: call-overload probe rescans imports on every call

### Symptom

Compiling large third-party packages with many function calls (e.g. `go.mongodb.org/mongo-driver/bson/bsoncodec`) could take **20+ minutes** on the fork while vanilla Go finished in seconds.

### Cause

`callExpr` called `overloadCandidatesForCall` on **every** function and method call. That function called `hasCallOverloads()`, which was **not cached** and, on each invocation, scanned the current package’s overload maps and **every import** (calling `EnsurePackageOperatorIndexes` per import) looking for names with more than one overload candidate.

For a package with **C** call sites and **I** imports, total work was **O(C × I)** even when the package defined no overloads at all.

### Fix

1. **`initForkFeatureCaches`** (in `assignOverloadSuffixes`) computes `pkgHasCallOverloads` once by scanning the current package and imports a single time.
2. **`hasCallOverloads()`** returns the cached flag — **O(1)** per call site.
3. **`overloadCandidatesForCall`** still does map lookup by name when overloads exist; the expensive import scan happens at most once per type-check pass.

**Locations:** `types2/overload.go`, `types2/perf_index.go`, `types2/check.go`

---

## Fixed: operator-overload probe rescans imports on every operator

### Cause

`applyBinaryOperatorOverload`, `tryUnaryOperatorOverload`, and `tryIncDecOperatorOverload` called `operatorOverloads(name)` on every binary/unary operation. That function scanned the current package and **every import** for each operator token (`+`, `-`, `==`, etc.), even in packages with no operator overloads.

On a file with **n** operators, work was **O(n × I)** per operator kind encountered.

### Fix

1. **`pkgHasOperatorOverloads`** is computed once in `initForkFeatureCaches`.
2. Operator probes return immediately when the flag is false.
3. When true, **`operatorOverloadsByName`** memoizes the merged candidate list per operator name (at most one import scan per distinct operator, not per use site).

**Locations:** `types2/operator.go`, `types2/perf_index.go`

---

## Fixed: extension-method name probe on every selector call

### Cause

`tryExtensionCall` ran on every selector call expression and called `extensionMethodExists(method)`, which looped over the current package and **all imports** on every probe, even when no package in the import graph declared extension methods.

For **S** selector calls and **I** imports, work was **O(S × I)**.

### Fix

1. **`pkgHasExtensions`** is computed once in `initForkFeatureCaches`.
2. **`callExpr`** skips `tryExtensionCall` entirely when the flag is false.
3. When extensions may exist, **`extensionMethodByName`** memoizes the yes/no result per method name.

**Locations:** `types2/call.go`, `types2/extension.go`, `types2/perf_index.go`

---

## Fixed: method call chains re-evaluate receiver (O(n²))

### Symptom

Long left-nested method call chains such as `rb.RegisterTypeDecoder(...).RegisterTypeDecoder(...)...` or `u.Child(0).Child(0)...` could make the fork compiler appear to hang. `default_value_decoders.go` (`RegisterDefaultDecoders`) and `test/torture.go` (`ChainUNoAssert`) were concrete examples.

### Cause

In `callExpr`, after type-checking `call.Fun` (a selector whose receiver `X` may be another call), argument checking called `exprOrType` on `sel.X` again to build `methodRecv` for lambda/iter hint substitution. Each call in a chain of length **n** re-walked the entire left spine, giving **O(n²)** `rawExpr` work.

When call overloads exist, `recvBaseNameFromExpr` also called `rawExpr` on the full receiver before `Types[sel.X]` was consulted.

### Fix

1. **`typedOperand`** / **`typeOfExpr`** reuse recorded type information from `Types[e]` or syntax node type info (`StoreTypesInSyntax`) instead of re-running `exprOrType` on `sel.X` when checking call arguments.
2. **`recvBaseNameFromExpr`** uses the same lookup before falling back to `rawExpr`.

**Locations:** `types2/call.go` (`typedOperand`, `callExpr`), `types2/overload.go` (`recvBaseNameFromExpr`)

---

## Fixed: enum variant probe on every call and composite literal

### Cause

`tryEnumVariantCall` ran at the start of **every** `callExpr`, including selector forms that call `enumTypeExpr` on the receiver. `tryEnumCompositeLit` and enum-aware `Name` handling in `exprInternal` also ran even when no package in the import graph declared enum types.

For packages with **C** call sites and no enums, work was **O(C)** in unnecessary enum probes (selector calls still evaluated receivers via `enumTypeExpr` before bailing out).

### Fix

1. **`pkgHasEnums`** is computed once in `initForkFeatureCaches` by scanning `objMap`, the current package scope, and imports.
2. **`callExpr`**, **`compositeLit`**, and enum-aware **`Name`** handling in `exprInternal` skip enum probes when the flag is false.

**Locations:** `types2/call.go`, `types2/expr.go`, `types2/literals.go`, `types2/perf_index.go`

---

## Fixed: method overload duplicate detection at package init

### Cause

`checkMethodOverloadDuplicates` compared every pair of method overload candidates with the same receiver and name using `identicalMethodSig` — **O(k²)** where **k** is the number of overloads for that method.

### Fix

`checkMethodOverloadDuplicates` uses a map keyed by `methodOverloadDupKey(name, sig)` — canonical receiver base, parameter suffix, result suffix, and variadic flag — **O(k)** per method name instead of **O(k²)** pairwise comparison.

**Location:** `types2/overload.go`

---

## Summary table

| Path | Complexity (before) | Status |
| ---- | ------------------- | ------ |
| Binary operator probe, no overloads in package | O(n²) on chain length | **Fixed** |
| Binary operator probe, overloads declared | O(n²) when probe fails | **Fixed** |
| Package-level overload duplicate check | O(k²) | **Fixed** (O(k)) |
| Method overload duplicate check | O(k²) | **Fixed** (O(k)) |
| Enum variant by name/obj | O(V) per lookup | **Fixed** (O(1)) |
| Extension candidates | O(\|imports\| × \|scope\|) | **Fixed** (O(\|imports\|)) |
| `operatorFuncsInPackage` fallback | O(\|scope\|) per lookup | **Fixed** (O(1) after index) |
| Overload selection (exact arg types) | O(k × args) per call | **Indexed** (see below) |
| Overload selection (assignability fallback) | O(k × args) per call | **Open** (memoized after first resolve) |
| Index operator re-check of base | 2× work on `e.X` | **Fixed** |
| `hasCallOverloads` per call site | O(I) per call | **Fixed** (O(1) flag) |
| `operatorOverloads` per operator use | O(I) per op | **Fixed** (O(1) flag; O(I) once per op name) |
| `extensionMethodExists` per selector | O(I) per call | **Fixed** (O(1) flag; O(I) once per method name) |
| Method call chain receiver re-check | O(n²) on chain length | **Fixed** (reuse `Types[sel.X]`) |
| Enum variant probe per call/lit | O(1)–O(expr) per site | **Fixed** (O(1) flag) |

---

## Overload selection: index hit vs assignability fallback

Overload resolution at a call site is a **two-tier** lookup:

1. **Exact-type index (O(1))** — `overloadBySig` maps `name + "·" + overloadParamSuffix(sig)` → `*Func` for each declared overload. At the call site, `lookupOverloadByArgTypes` builds the same suffix from the **actual argument operand types** (`operandTypesSuffix`) and looks up that key in `overloadBySig` (checker-local and per-import package maps built in `buildCheckerIndexes` / `ensurePackageOverloadBySig`).

2. **Assignability fallback (O(k × args))** — when the index misses, `selectOperatorFunc` / `selectOverloadEx` scan every candidate overload and run `assignableTo` on each parameter. This is required when argument types are not identical to the declared parameter types but still assignable (untyped constants, named types vs underlying types, interface targets, variadic `...` expansion, etc.). The suffix keys use `briefType`, which is a lossy string form — it cannot encode full assignability rules.

3. **Memoization** — when the fallback finds a unique match, `overloadResolveCache` stores `(candidate set, arg suffix) → *Func` so the same call shape is not rescanned in one type-check pass.

**Why not map every assignability case?** A perfect map would need keys for every assignable (arg type, param type) pair, or would require precomputing assignability edges between all types in scope. The exact-type index covers the common case (overload declared for `int`, call with `int`). The fallback handles the rest; **k** (overload count) is typically small (2–5).

**Location:** `types2/perf_index.go` (`lookupOverloadByArgTypes`), `types2/operator.go` (`selectOperatorFunc`), `types2/call.go` (`selectOverloadEx`).

---

## Open issues (known quadratic / expensive paths not yet fixed)

These paths are **intentionally left as-is** today because **N** is small in practice, or because a full fix would require a much larger assignability index. They are listed so maintainers know where compile time can still grow if **N** is unusually large.

### Fork / types2

| Path | File | Complexity | What **N** is | Notes |
| ---- | ---- | ---------- | ------------- | ----- |
| Overload selection assignability fallback | `operator.go`, `call.go` | O(k × args) | Overloads for one name × parameters | Index hit is O(1); fallback when arg types ≠ declared param types. Memoized via `overloadResolveCache` after first resolve. |
| Expression switch duplicate values | `stmt.go` `caseValues` | O(k²) per value bucket | Cases sharing same underlying constant | Linear when values are distinct (typical). Worst case: many typed variants of one constant on an interface switch. |
| Extension candidate import loop | `extension.go` | O(I) per resolution | Imports when extensions exist | Each step is `extensionByName[method]` (O(1)), not a scope scan. |
| Enum presence scan | `perf_index.go` `packageHasEnums` | O(\|scope\|) once per import | Names in package scope | Runs once in `initForkFeatureCaches`, not per call site. |

### Upstream types2 (inherited)

| Path | File | Complexity | What **N** is | Notes |
| ---- | ---- | ---------- | ------------- | ----- |
| Union term overlap | `union.go` | O(T²) | Terms in one `\|` constraint | Hard cap 100 terms. |
| Interface identity / unify stack | `unify.go`, `predicates.go` | O(depth²) | Nested interface pairs on stack | depth usually 0–3. |
| Generic type inference phase 2 | `infer.go` | O(n²) | Type params on one generic func | n usually &lt; 5. |

### SSA / backend (post type-check)

| Path | File | Complexity | Notes |
| ---- | ---- | ---------- | ----- |
| Dominator computation | `ssa/dom.go` | O(n²) worst case | TODO in comment; mitigated in practice. |
| Prove pass loop | `ssa/prove.go` | Can be quadratic | TODO at line ~2216. |
| Regalloc live-value map | `ssagen/ssa.go` | O(n²) on large entry blocks | Many live values in one block. |
| Rangefunc rewrite | `rangefunc/rewrite.go` | O(depth²) | Nesting depth of range-over-func. |

Other SSA passes (`cse`, `copyelim`, `fuse`, `deadstore`) document quadratic risks and include explicit guards.

---

## Upstream quadratic algorithms (not fork additions)

The fork inherits several algorithms that are already documented as quadratic in upstream `types2`. These are **local** to a single switch, constraint, interface comparison, or generic call — they do not scale with file length or import count the way the fork’s operator-overload bug did. **Type switch duplicate detection was changed to O(1) map lookup; the rest are unchanged.**

| Area | File | What **N** is | Typical **N** | Status |
| ---- | ---- | ------------- | ------------- | ------ |
| Duplicate case values in expression switch | `stmt.go` | Cases per underlying value bucket | 1–2 (5–30 cases total) | Upstream (O(k²) per bucket; O(C) when values distinct) |
| Duplicate types in type switch | `stmt.go` | Type cases in one switch | 3–15 | **Fixed** (O(1) lookup) |
| Union term overlap | `union.go` | Terms in one `\|` constraint | 2–8 (max 100) | Upstream |
| Interface unify stack scan | `unify.go`, `predicates.go` | Recursion depth comparing interfaces | 0–3 | Upstream |
| Generic type inference | `infer.go` | Type parameters on one generic func | 1–4 (< 5) | Upstream |

---

### Duplicate case values in expression switch

**Location:** `caseValues` in `types2/stmt.go`

Duplicate detection runs only for **constant** int, float, and string cases. Non-constant cases skip the inner check entirely.

The checker keeps a map keyed by **underlying Go value**, not by case index:

```go
seen valueMap  // map[underlying value] → prior cases with that value
```

For each new constant case, the inner loop scans only `seen[val]` — cases that already used the **same** underlying value (same `1`, same `"foo"`, etc.):

```257:270:src/cmd/compile/internal/types2/stmt.go
		if val := goVal(v.val); val != nil {
			// look for duplicate types for a given value
			// (quadratic algorithm, but these lists tend to be very short)
			for _, vt := range seen[val] {
				if Identical(v.typ(), vt.typ) {
					// ... duplicate case error ...
				}
			}
			seen[val] = append(seen[val], valueType{v.Pos(), v.typ()})
		}
```

**Example — distinct constants (linear, not quadratic):**

```go
switch x {
case 1:
case 2:
case 3:
// ... through 10
}
```

Each value hits an empty `seen[val]` list → **0** inner comparisons per case → **O(10)** total for 10 cases, **not** O(100).

**Example — same underlying value, different types (interface switch):**

```go
type myByte byte

func f(x any) {
    switch x {
    case byte(1):
    case myByte(1): // same underlying 1 → compare against seen[1]
    case 2, 3, 4:
    }
}
```

This matters mainly when switching on an **interface** value, where the same numeric constant can appear with different types.

**Complexity:** If **k** cases share the same underlying value, duplicate checking for that value bucket costs 0 + 1 + … + (k−1) = **O(k²)**. Across the whole switch with distinct values 1, 2, …, C, total work is **O(C)**.

The upstream comment calls this “quadratic” because of the **k²** worst case when many typed variants of one constant appear — not because C case clauses implies C² checks.

**Typical N:** 5–30 cases per switch; **k** (cases per value bucket) is usually 1–2.

---

### Duplicate types in type switch (fixed)

**Location:** `caseTypes` in `types2/stmt.go`

For each `case T:` in `switch x.(type)`, the checker detects duplicate types using a **map keyed by a structural type hash** (`typeSwitchCaseKey`), not a linear scan over prior cases.

**Example:**

```go
func f(x any) {
    switch x.(type) {
    case int:
    case string:
    case int: // error: duplicate case int
    case nil, nil:
    }
}
```

**Complexity:** **O(C)** for **C** type expressions in the switch — one map lookup and insert per type.

**Implementation:** `typeSwitchCaseKey` uses `newTypeHasher` with the checker's `Context` when available (so identical anonymous struct types from different case clauses hash the same); falls back to `TypeString` otherwise.

---

### Union term overlap

**Location:** `parseUnion` / `overlappingTerm` in `types2/union.go`

When parsing a constraint union like `int | ~string | MyType`, each new term is checked against **all previous terms** for overlap (`a|a`, `~a|A`, etc.).

**Example:**

```go
func Sum[T interface {
    int | int64 | float64 | ~int32 | MyInt
}](s []T) T { /* ... */ }
```

Each term in the union is compared to earlier terms via `overlappingTerm(terms[:i], t)`.

**Complexity:** O(T²) where **T** is the number of terms in one constraint expression.

**Typical N:** 2–8 terms (e.g. `int | int64`, `~[]byte | string`).

**Hard cap:** `maxTermCount = 100` — unions with more than 100 terms are rejected before quadratic work becomes extreme. Worst case: ~10,000 term-pair checks per union, and most packages have only a handful of generic constraints.

---

### Interface unify / identity stack scan

**Location:** `identical` in `types2/predicates.go`; unification in `types2/unify.go`

When checking whether two **interface types** are identical (or unify), the checker walks method sets recursively. To stop infinite recursion on self-referential interfaces, it keeps a linked list (`ifacePair` stack) of `(x, y)` pairs already being compared. Before recursing, it scans that stack — **O(depth²)** per comparison path.

**Example (pathological, from comments in code):**

```go
type T interface {
    m() interface{ T }
}
```

Comparing two different named interfaces that embed this pattern pushes pairs onto the stack.

**Complexity:** **N** = stack depth = number of nested interface pairs currently being compared (not package or file size).

**Typical N:** 0–3 — almost all interfaces are non-recursive (`io.Reader`, `fmt.Stringer`, etc.).

**Rare N:** 5–10 for deeply self-referential generic constraint interfaces. The code notes this is “extremely rare”; a hash map would cost more than the tiny stack scan.

---

### Generic type inference

**Location:** phase 2 of inference in `types2/infer.go`

After a generic call, the checker repeatedly unifies each type parameter with its constraint until no progress. The outer loop runs up to **n** times (at least one type argument inferred per iteration); the inner loop scans all **n** type parameters → **O(n²)**.

**Example:**

```go
func Map[T, U any](s []T, f func(T) U) []U { /* ... */ }

// inference for T and U from arguments:
Map([]int{1, 2}, func(x int) string { return strconv.Itoa(x) })
```

**Complexity:** **N** = number of type parameters on the generic function or method being instantiated.

**Typical N:** 1–4 (`Map[K,V]`, `Reduce[T]`, `New[T any]`). Comment in code: “< 5 or so.”

**Worst realistic N:** 10–20 on heavily generic library code → 100–400 inner-loop iterations per call site, still tiny compared to compiling a whole file.

---

## Related reading

- [Operator overloading](operator_overloading.md) — declaration and resolution rules
- [Function and method overloading](overloading.md) — overload selection at call sites
- [Extension methods](extension_methods.md) — extension resolution
- [Enums](enums.md) — variants and exhaustive switching
