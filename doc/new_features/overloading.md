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
- **Ambiguous overloaded call** — more than one overload fits equally well (for example, two overloads taking different named types with the same underlying type and an untyped constant argument).

### Notes

- Overloading applies to both package-level functions and methods.
- Overload sets must be unambiguous for all valid calls.
- The blank identifier `_` is not overloadable; duplicate `func _()` declarations are still invalid.
