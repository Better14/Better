# If and Switch Expressions

## If Expressions

Go supports `if` as an expression that evaluates to a value.

```go
a := if 5 < 6 { 1 } else { 2 }
```

Both branches must be expressions with compatible types. The result type is the common type of the branch expressions.

## Switch Expressions

Go supports `switch` as an expression that evaluates to a value.

```go
a := switch x {
case 1:
	"one"
case 2:
	"two"
default:
	"other"
}
```

Each case arm must be an expression (or a single expression after `:`). All arms must have compatible types. The result type is the common type of the case expressions.
