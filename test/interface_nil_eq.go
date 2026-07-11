// run

// Test that interface values holding typed nil pointers compare equal to nil.

package main

type SomeInterface interface {
	Do()
}

type T struct{}

func (*T) Do() {}

type myErr struct{ msg string }

func main() {
	var a SomeInterface = nil
	if a != nil {
		panic("direct nil assignment")
	}

	var p *T = nil
	var b SomeInterface = p
	if b != nil {
		panic("typed nil pointer in interface")
	}

	var c SomeInterface
	if c != nil {
		panic("zero value interface")
	}

	var err error
	var me *myErr = nil
	err = me
	if err != nil {
		panic("typed nil error")
	}

	t := &T{}
	var d SomeInterface = t
	if d == nil {
		panic("non-nil interface value")
	}
}
