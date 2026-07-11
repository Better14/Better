// run

package main

func (i int) Twice() int { return i + i }

func main() {
	x := 21
	if x.Twice() != 42 {
		panic("extension int failed")
	}
}
