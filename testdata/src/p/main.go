package p

// Main function to use the "false positive" identifiers
func Main() {
	// Use common function from false_positives.go
	_ = commonFunction()

	// Use CommonType from false_positives.go
	_ = CommonType{Field: "main"}

	// Use commonConstant from false_positives.go
	_ = commonConstant

	// Use TestInterface from interfaces.go
	var ti TestInterface = &TestInterfaceImpl{}
	_ = ti.InterfaceMethod()

	// Use TestAlias from interfaces.go
	var alias TestAlias = aliasFunction()
	_ = alias

	// Use testPrivateType from interfaces.go
	tp := &testPrivateType{field: "main"}
	_ = tp.privateMethod()
	_ = tp.PublicMethod()
}
