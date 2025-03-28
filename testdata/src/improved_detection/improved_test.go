package improved_detection

import "testing"

func TestImprovedDetection(t *testing.T) {
	// Test the test-only function
	result := TestOnlyFunctionWithComment()
	if result != "test function with comment" {
		t.Errorf("Expected 'test function with comment', got '%s'", result)
	}

	// Test functions that are used in regular code through improved detection
	if ComplexUsage() != "complex usage" {
		t.Errorf("Expected 'complex usage', got '%s'", ComplexUsage())
	}

	if ReflectivelyUsed() != "reflectively used" {
		t.Errorf("Expected 'reflectively used', got '%s'", ReflectivelyUsed())
	}

	if StringUsed() != "string used" {
		t.Errorf("Expected 'string used', got '%s'", StringUsed())
	}

	if ActuallyUsedInCode() != "actually used" {
		t.Errorf("Expected 'actually used', got '%s'", ActuallyUsedInCode())
	}

	// Test the complex usage pattern function
	UseComplexPattern()
}
