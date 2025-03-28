package test_helpers

// AssertEqual is a test helper function that should be excluded
func AssertEqual(a, b interface{}) bool { // want "function 'AssertEqual' is only used in tests"
	return a == b
}

// MockData is a test mock that should be detected
type MockData struct { // want "type 'MockData' is only used in tests"
	ID   int
	Name string
}

// FakeClient is a test fake that should be detected
type FakeClient struct { // want "type 'FakeClient' is only used in tests"
	URL string
}

// Get is a method on a test-only type
func (f FakeClient) Get() string { // want "method 'Get' is only used in tests"
	return "data from " + f.URL
}

// TestSetup looks like a test helper
func TestSetup() string { // want "function 'TestSetup' is only used in tests"
	return "test environment"
}

// testCleanup looks like a test helper
func testCleanup() { // want "function 'testCleanup' is only used in tests"
	// Cleanup resources
}

// UsedFunction is actually used in regular code
func UsedFunction() string {
	return "used"
}

// RegularFunction uses the normal function
func RegularFunction() {
	_ = UsedFunction()
}
