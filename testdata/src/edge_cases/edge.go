package edge_cases

import "fmt"

// EmptyFunction does nothing but is only called in tests
func EmptyFunction() { // want "function 'EmptyFunction' is only used in tests"
}

// functionWithUnusedParam has an unused parameter and is only used in tests
func functionWithUnusedParam(unused string) string { // want "function 'functionWithUnusedParam' is only used in tests"
	return "test"
}

// TestHelperFunc looks like a test helper but is only used in tests
func TestHelperFunc(t interface{}) bool { // want "function 'TestHelperFunc' is only used in tests"
	return true
}

// GenerateTestData sounds like test-only code and is only used in tests
func GenerateTestData() []string { // want "function 'GenerateTestData' is only used in tests"
	return []string{"test1", "test2"}
}

// testSetup has a test-like name but is used in regular code
func testSetup() string {
	return "setup"
}

// Used in both test and non-test code through string formatting
func FormatName(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Function that uses the non-test function
func RegularFunction() {
	fmt.Println(testSetup())
	fmt.Println(FormatName("World"))
}
