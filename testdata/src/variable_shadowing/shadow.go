package variable_shadowing

// TopLevelTestOnlyFunction is only used in tests
func TopLevelTestOnlyFunction() string { // want "function 'TopLevelTestOnlyFunction' is only used in tests"
	return "top level"
}

// TestFunction has a parameter that shadows a global function
func TestFunction(TopLevelTestOnlyFunction func() string) string {
	return "shadowed: " + TopLevelTestOnlyFunction()
}

// NormalFunction uses shadowing but is used in production
func NormalFunction() string {
	// Local variable shadows the function
	var TopLevelTestOnlyFunction = "not a function"
	return "normal: " + TopLevelTestOnlyFunction
}

// OnlyInTestsWithShadowing is used only in tests and involves shadowing
func OnlyInTestsWithShadowing() string { // want "function 'OnlyInTestsWithShadowing' is only used in tests"
	// This shadow shouldn't affect detection
	TopLevelTestOnlyFunction := func() string {
		return "shadowed impl"
	}
	return TopLevelTestOnlyFunction()
}

// UsedInCodeWithShadow uses a shadow in production code
func UsedInCodeWithShadow() {
	// Create a local function with the same name
	TopLevelTestOnlyFunction := func() string {
		return "local shadow"
	}

	// Use the local shadow
	_ = TopLevelTestOnlyFunction()

	// Use the normal function
	_ = NormalFunction()
}
