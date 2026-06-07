# Function and Method Overloading

Go now supports function and method overloading.

Multiple declarations with the same name are valid as long as their parameter lists differ by type and/or arity.

For example:

```go
func myFunc(a int) {
}

func myFunc(a int64) {
}

func myFunc(a int, b int) {
}
```

All of the above declarations are valid.

### Overload Resolution

At a call site, the compiler resolves the overload by matching argument count and argument types.

- `myFunc(10)` resolves to `func myFunc(a int)`
- `myFunc(int64Value)` resolves to `func myFunc(a int64)`
- `myFunc(1, 2)` resolves to `func myFunc(a int, b int)`

### IDE support (gopls)

When you type a call such as `myFunc(`, signature help lists **every** overload signature. Diagnostics for invalid calls include the available overload signatures in the error message.

### Errors

The type checker reports:

- **Redeclared** — two overloads with the same parameter types (`redeclared function f` / `redeclared method m`).
- **No matching overload** — no overload accepts the argument types at a call site.
- **Ambiguous overload set for f** — two or more overloads of the same name can match the same argument count and types (including when [default arguments](default_arguments.md) make effective arities overlap). Reported when type-checking the conflicting declaration(s); the overload set is ill-formed and the package does not compile.
- **Ambiguous overloaded call** — more than one overload fits equally well at a call site (for example, two overloads taking different named types with the same underlying type and an untyped constant argument). Also reported for calls that would be ambiguous if an ill-formed overload set were not already rejected.

Example of **ambiguous overload set** (compile error on the declarations):

```go
func myFunc(a int, b int = 5) {} // ERROR: ambiguous overload set for myFunc
func myFunc(a int) {}            // ERROR: ambiguous overload set for myFunc
```

See [Default arguments and ambiguous overloads](default_arguments.md#default-arguments-and-ambiguous-overloads).

### Notes

- Overloading applies to both package-level functions and methods.
- Overload sets must be unambiguous for all argument counts and types any member could accept (after applying default arguments). Ill-formed sets are compile errors at declaration time, not only at call sites.
- The blank identifier `_` is not overloadable; duplicate `func _()` declarations are still invalid.
