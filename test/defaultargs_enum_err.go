// errorcheck


package p

enum Mode {
	Read
	ByName(string)
}

func bad(mode Mode = ByName("x")) {} // ERROR "compile-time constant"
