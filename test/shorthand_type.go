// run

package main

struct Person {
	Name string
	Age  int
}

interface Stringer {
	String() string
}

func (p Person) String() string {
	return p.Name
}

func main() {
	p := Person{Name: "Ada", Age: 42}
	var s Stringer = p
	if s.String() != "Ada" {
		panic(s.String())
	}
	println("ok")
}
