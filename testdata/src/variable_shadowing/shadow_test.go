package variable_shadowing

import "testing"

func TestVariableShadowing(t *testing.T) {
	// Test using the top level function directly
	result := TopLevelTestOnlyFunction()
	if result != "top level" {
		t.Errorf("Expected 'top level', got '%s'", result)
	}

	// Test TestFunction with a custom function parameter that shadows the global
	customResult := TestFunction(func() string {
		return "custom function"
	})
	if customResult != "shadowed: custom function" {
		t.Errorf("Expected 'shadowed: custom function', got '%s'", customResult)
	}

	// Test with the test-only shadowing function
	shadowResult := OnlyInTestsWithShadowing()
	if shadowResult != "shadowed impl" {
		t.Errorf("Expected 'shadowed impl', got '%s'", shadowResult)
	}

	// Also test normal functions
	normalResult := NormalFunction()
	if normalResult != "normal: not a function" {
		t.Errorf("Expected 'normal: not a function', got '%s'", normalResult)
	}

	// Call production code function that uses shadowing
	UsedInCodeWithShadow()
}
