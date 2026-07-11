// run

package main

var add func(int, int) int = (a, b) => a + b

func main() {
	if add(1, 2) != 3 {
		panic(add(1, 2))
	}
}
