package basic_test_cases

// TestOnlyFunction is a function that is only used in tests
func TestOnlyFunction() string { // want "function 'TestOnlyFunction' is only used in tests"
	return "test"
}

// UsedFunction is a function that is used in both test and non-test code
func UsedFunction() string {
	return "used"
}

// anotherTestOnlyFunction is an unexported function only used in tests
func anotherTestOnlyFunction() int { // want "function 'anotherTestOnlyFunction' is only used in tests"
	return 42
}

// TestOnlyType is a type that is only used in tests
type TestOnlyType struct { // want "type 'TestOnlyType' is only used in tests"
	Value string
}

// TestOnlyConstant is a constant that is only used in tests
const TestOnlyConstant = "test constant" // want "const 'TestOnlyConstant' is only used in tests"

// TestOnlyVariable is a variable that is only used in tests
var TestOnlyVariable = "test variable" // want "variable 'TestOnlyVariable' is only used in tests"

// normalFunction is used in the package itself
func normalFunction() string {
	return UsedFunction()
}
