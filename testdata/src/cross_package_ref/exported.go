package cross_package_ref

// ExportedTestFunc is exported but only used in tests from another package
func ExportedTestFunc() string { // want "function 'ExportedTestFunc' is only used in tests"
	return "exported test function"
}

// ExportedUsedFunc is exported and used in regular code from another package
func ExportedUsedFunc() string {
	return "exported used function"
}

// TestOnlyType is only used in tests from another package
type TestOnlyType struct { // want "type 'TestOnlyType' is only used in tests"
	Name string
}

// UsedType is used in regular code from another package
type UsedType struct {
	ID int
}

// TestOnlyConst is only used in tests
const TestOnlyConst = "test-only constant" // want "const 'TestOnlyConst' is only used in tests"

// UsedConst is used in regular code
const UsedConst = "used constant"

var TestOnlyVar = "test-only variable" // want "variable 'TestOnlyVar' is only used in tests"

var UsedVar = "used variable"
