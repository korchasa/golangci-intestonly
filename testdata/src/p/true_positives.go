package p

// Test case for functions only used in tests
func testOnlyFunction() string { // want "identifier \"testOnlyFunction\" is only used in test files but is not part of test files"
	return "test"
}

// Test case for types only used in tests
type TestOnlyType struct { // want "identifier \"TestOnlyType\" is only used in test files but is not part of test files"
	Field string
}

// Test case for constants only used in tests
const testOnlyConstant = "test" // want "identifier \"testOnlyConstant\" is only used in test files but is not part of test files"
