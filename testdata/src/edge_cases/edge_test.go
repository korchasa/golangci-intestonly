package edge_cases

import "testing"

func TestEdgeCases(t *testing.T) {
	// Call empty function
	EmptyFunction()

	// Call function with unused param
	result := functionWithUnusedParam("not used")
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Use test helper function
	if !TestHelperFunc(t) {
		t.Error("Expected TestHelperFunc to return true")
	}

	// Use data generator
	data := GenerateTestData()
	if len(data) != 2 {
		t.Errorf("Expected 2 items, got %d", len(data))
	}

	// Use regular function
	if testSetup() != "setup" {
		t.Errorf("Expected 'setup', got '%s'", testSetup())
	}

	// Use FormatName
	formatted := FormatName("Tester")
	if formatted != "Hello, Tester!" {
		t.Errorf("Expected 'Hello, Tester!', got '%s'", formatted)
	}
}
