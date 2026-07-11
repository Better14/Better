// errorcheck


// T? is not assignable to T without an explicit unwrap.

package main

func myPrint(a int) {}

func main() {
	var a int?
	myPrint(a) // ERROR "cannot use"
}
