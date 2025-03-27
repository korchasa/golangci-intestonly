package p

// This function is only used in test files
// Should be reported by the linter
func helperFunction() string { // want "identifier \"helperFunction\" is only used in test files but is not part of test files"
	return "helper"
}
