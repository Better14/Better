// run

// Receiving from a nil channel panics in Bow instead of blocking forever.

package main

func main() {
	var ch chan int
	defer func() {
		if recover() == nil {
			panic("expected panic from nil channel receive")
		}
	}()
	<-ch
}
