package p

// Test case for functions used through reflection
func reflectionFunction() string { // want "identifier \"reflectionFunction\" is only used in test files but is not part of test files"
	return "reflection"
}

// Test case for functions used through type assertions
type TestType struct {
	Field string
}

func (t *TestType) testMethod() string { // want "identifier \"testMethod\" is only used in test files but is not part of test files"
	return "type assertion"
}
