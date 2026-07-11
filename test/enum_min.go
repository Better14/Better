// run

package main

enum SomeEnum {
	Value1
	Value2(string)
}

func main() {
	var msg SomeEnum = Value2("hello")
	switch v := msg; v {
	case Value1:
	case Value2(s):
		println(s)
	}
}
