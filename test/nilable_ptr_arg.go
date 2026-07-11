// run

package main

func take(p *string?) {}

func main() {
	s := "id"
	take(&s)
}
