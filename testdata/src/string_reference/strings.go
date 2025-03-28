package string_reference

// ReferencedInString is a function that is referenced in a string literal
func ReferencedInString() string {
	return "referenced in string"
}

// UsedNormally is a function that is used normally
func UsedNormally() string {
	return "used normally"
}

// OnlyInTestString is a function that is referenced in a string literal only in tests
func OnlyInTestString() string { // want "function 'OnlyInTestString' is only used in tests"
	return "only in test string"
}

// TestOnlyFunc is only used in tests
func TestOnlyFunc() string { // want "function 'TestOnlyFunc' is only used in tests"
	return "test only"
}

// UseFunctionInString uses a function name in a string
func UseFunctionInString() string {
	// This string contains a reference to ReferencedInString which should
	// prevent it from being reported as test-only
	return "This function calls ReferencedInString() to get data"
}

// UseNormalFunc uses a function directly
func UseNormalFunc() string {
	return UsedNormally()
}
