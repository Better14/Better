// errorcheck

package p

type T int

func ==(a, b T) bool { return a == b }
// ERROR "requires paired operator"
