
package main

import "errors"

func panicStructuredError() {
	err := errors.New("boom")
	panic(err)
}

func init() {
	register("panicStructuredError", panicStructuredError)
}
