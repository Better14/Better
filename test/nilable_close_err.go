// errorcheck


//go:nilable_pointers enable

package main

func main() {
	var ch chan int?
	close(ch) // ERROR "without nil check"
	if ch != nil {
		close(ch)
	}
}
