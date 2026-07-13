// run

package main

func routePattern() string {
	return "{key-id:.*}"
}

func main() {
	_ = routePattern()
}
