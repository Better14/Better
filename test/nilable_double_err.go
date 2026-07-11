// errorcheck


// Applying ? twice is invalid.

package main

func main() {
	var _ (int?)? = nil // ERROR "invalid nilable type"
}
