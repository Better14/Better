// run

package main

func main() {
	ch := make(chan int, 1)
	ch <- 1
	close(ch)
	go func() {
		for v in ch {
			_ = v
		}
	}()
}
