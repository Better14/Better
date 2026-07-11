// errorcheck


// Test overload resolution errors.

package main

type A int
type B int

func addInt(a int) int     { return 0 }
func addInt(a int64) int64 { return 0 }

func addAmbig(a A) int { return 0 }
func addAmbig(a B) int { return 0 }

func dup(int)     {}
func dup(int) {} // ERROR "redeclared function dup|redeclared in this block"

func badNoMatch() {
	addInt("x") // ERROR "no matching overload"
}

func badAmbig() {
	addAmbig(1) // ERROR "ambiguous overloaded"
}

func main() {}
