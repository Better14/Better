package forin

func f(items []string) {
	for _, item := range items { // want "range loop can use for-in syntax"
		_ = item
	}
	for i, item := range items { // want "range loop can use for-in syntax"
		_, _ = i, item
	}
	for k := range items { // want "range loop can use for-in syntax"
		_ = k
	}
}
