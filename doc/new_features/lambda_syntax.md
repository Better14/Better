# Lambda Syntax (`=>`)

For single-expression functions, Go supports arrow lambda syntax. Parameter types are inferred from context.

Before:

```go
func(a, b, c int) int { return a + b + c }
```

After:

```go
(a, b, c) => a + b + c
```

If the function body requires more than one expression or any statement, use the standard `func` syntax:

```go
func(a, b, c int) {
	sum := a + b + c
	return sum
}
```

### Notes

- `=>` lambdas are limited to a single expression body.
- Parameter types are inferred when the lambda appears in a typed context (e.g. assignment, argument, return).
- Multi-statement or multi-expression bodies must use `func(...) { ... }`.
