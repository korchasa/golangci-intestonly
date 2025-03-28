package interfaces

import "testing"

func TestInterfaces(t *testing.T) {
	// Test using the test-only interface and implementation
	var testInterface TestOnlyInterface
	testImpl := TestOnlyImplementation{Data: "test data"}
	testInterface = testImpl

	result := testInterface.TestMethod()
	if result != "test data" {
		t.Errorf("Expected TestMethod to return 'test data', got '%s'", result)
	}

	// Test using the regular interface and implementation in the test
	var regularInterface RegularInterface
	regularImpl := NewRegularImplementation("regular data")
	regularInterface = regularImpl

	regularResult := regularInterface.RegularMethod()
	if regularResult != "regular data" {
		t.Errorf("Expected RegularMethod to return 'regular data', got '%s'", regularResult)
	}

	// Test the test-only method on regular implementation
	testOnlyResult := regularImpl.TestOnlyMethod()
	if testOnlyResult != "test: regular data" {
		t.Errorf("Expected TestOnlyMethod to return 'test: regular data', got '%s'", testOnlyResult)
	}

	// Test using the regular interface function
	useResult := UseRegularInterface(regularImpl)
	if useResult != "Using: regular data" {
		t.Errorf("Expected UseRegularInterface to return 'Using: regular data', got '%s'", useResult)
	}
}
