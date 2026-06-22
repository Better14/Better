# Enums

Go supports Rust-style **algebraic enums** (tagged unions): each variant is one of several named forms, with optional payloads and optional explicit discriminants.

`enum` is a **contextual keyword**: it starts a top-level enum declaration (`enum Name { ... }`) but remains a normal identifier elsewhere (parameter names, range variables, struct fields, etc.). This keeps third-party code that uses `enum` as an identifier compatible with BetterGo.

### Declaration

```go
enum SomeEnum {
	Value1
	Value2(String)
	Value3(int)
	Value4 = 3
}
```

- **Unit variant** — `Value1` carries no data.
- **Tuple variant** — `Value2(String)`, `Value3(int)` attach one or more payload types (tuple variants).
- **Explicit discriminant** — `Value4 = 3` assigns a fixed numeric tag (for C/interop or stable layout); variants without `=` get auto-incremented tags where applicable.

Struct-style variants (named fields) are also supported:

```go
enum Message {
	Quit
	Write { text string, bytes int }
	ChangeColor { r, g, b uint8 }
}
```

### Construction

When the expected type is known, **omit the enum name** and use the variant alone:

```go
var a SomeEnum = Value1
var b SomeEnum = Value2("hello")
c := Value3(42)              // type inferred from context
func open(mode Mode = Read)  // default argument
```

The qualified form is always valid:

```go
a := SomeEnum.Value1
b := SomeEnum.Value2("hello")
c := SomeEnum.Value3(42)
d := SomeEnum.Value4

m := Message.Write{ text: "hi", bytes: 5 }
```

### Switching on enums

Use a `**switch` statement** or `**switch` expression** to branch on the active variant and bind payloads. There is no `match` keyword.

**Exhaustiveness** — if a switch has no `default` case, the compiler requires every variant to be covered; missing any variant is a **compile-time error**. A `default` case is **allowed** and satisfies exhaustiveness (you may switch on a subset of variants and handle the rest in `default`).

Inside a `switch` on an enum, **case labels omit the enum type name** — write `case Value1:` not `case SomeEnum.Value1:`.

Switch statement:

```go
switch v {
case Value1:
	fmt.Println("value1")
case Value2(s):
	fmt.Println(s)
case Value3(n):
	fmt.Println(n)
case Value4:
	fmt.Println("value4")
}
```

Switch expression:

```go
n := switch v {
case Value1:
	0
case Value2(s):
	len(s)
case Value3(n):
	n
case Value4:
	3
}

desc := switch m {
case Quit:
	"quit"
case Write { text }:
	text
case ChangeColor { r, g, b }:
	int(r) + int(g) + int(b)
}
```

**Ignoring struct variant fields** — when you do not need the payload, use an empty struct pattern `{}` or list `_` for each field you want to discard:

```go
enum Color {
	Red { r, g, b uint8 }
	Green
	Blue(int)
}

// Ignore all fields (preferred when you only care about the variant):
switch c {
case Red {}:
	fmt.Println("red")
case Green:
	fmt.Println("green")
case Blue(_):
	fmt.Println("blue")
}

// Same meaning, but list every field and ignore each one explicitly:
switch c {
case Red { _, _, _ }:
	fmt.Println("red")
case Green:
	fmt.Println("green")
case Blue(_):
	fmt.Println("blue")
}
```

**Struct pattern field counts** — a case pattern cannot list more entries than the variant has fields. For `Red { r, g, b uint8 }`, at most three field slots may appear: `Red {}`, `Red { _, _, _ }`, and `Red { r, _, _ }` are valid; `Red { _, _, _, _ }` is a compile error. An empty pattern `Red {}` always matches the variant without binding any fields.

When you list every field in a struct pattern, bindings are **positional**: names must match the variant field at that index, or use `_` to skip binding. A pattern that lists only `_` must name every field (`{ _, _, _ }` for three fields). Partial patterns match by field name — for example, `Write { text }` binds only `text`, and `Red { _, g, _ }` binds only `g`.

Invalid (compile error — missing `Value4` and no `default`):

```go
// switch v {
// case Value1:
// case Value2(s):
// case Value3(n):
// } // ERROR: switch on SomeEnum is not exhaustive
```

Valid ( `default` satisfies exhaustiveness):

```go
switch v {
case Value1:
	fmt.Println("value1")
default:
	fmt.Println("other")
}
```

### Methods and generics

Enums may have methods and type parameters:

```go
enum Option[T] {
	None
	Some(T)
}

func (o Option[int]) IsSome() bool {
	switch o {
	case Some(_):
		return true
	case None:
		return false
	}
}
```

### Notes

- Enums are distinct from `int?` nullable types and from `iota` constant groups.
- Variant names live in the enum’s namespace. Use unqualified names when the type is known (`var a SomeEnum = Value1`, `case Value2(s):`, `mode Mode = Read`) or the qualified form (`SomeEnum.Value2`) anywhere.
- Memory layout is implementation-defined; explicit discriminants (`Value4 = 3`) document ABI intent.
- Unit `enum` variants may appear in default arguments when the default is a compile-time constant variant (e.g. `mode Mode = Read`). Tuple and struct variants are not valid defaults.
- When new variants are added to an enum, every non-`default` `switch` on that type must be updated or the build fails (exhaustiveness checking).
