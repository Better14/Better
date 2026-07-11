// run

package main

import (
	"person"
	"personext"
)

func main() {
	a := person.Person{Name: "Ada"}
	if a.Hello() != "Hi, my name is Ada" {
		panic("extension call failed")
	}
}
