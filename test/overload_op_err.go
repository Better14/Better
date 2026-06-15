// errorcheck

package p

type T int

func ==(a, b T) bool { return a == b } // ERROR "operator == requires paired operator !="
