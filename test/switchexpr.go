// run

package main

import "fmt"

func main() {
	x := 2
	a := switch x {
	case 1:
		"one"
	case 2:
		"two"
	default:
		"other"
	}
	if a != "two" {
		panic(fmt.Sprintf("switch expr: got %q", a))
	}
	fmt.Println("ok")
}
