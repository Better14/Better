// errorcheck


package main

enum Section {
	MatrixOps { size int, operators []string }
}

func main() {
	_ = MatrixOps{size: 2, operators: nil} // ERROR "undefined: MatrixOps"
}
