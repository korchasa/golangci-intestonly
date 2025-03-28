package false_negatives

import "testing"

func TestFalseNegatives(t *testing.T) {
	// Test all functions to ensure they have test coverage too

	// Test IndirectlyUsedFunction
	indirect := IndirectlyUsedFunction()
	if indirect != "indirect" {
		t.Errorf("Expected 'indirect', got '%s'", indirect)
	}

	// Test UsedViaReflection
	reflection := UsedViaReflection()
	if reflection != "reflection" {
		t.Errorf("Expected 'reflection', got '%s'", reflection)
	}

	// Test UsedInStringLiteral
	literal := UsedInStringLiteral()
	if literal != "string literal" {
		t.Errorf("Expected 'string literal', got '%s'", literal)
	}

	// Test DynamicallyUsed
	dynamic := DynamicallyUsed()
	if dynamic != "dynamic" {
		t.Errorf("Expected 'dynamic', got '%s'", dynamic)
	}

	// Test the Main function that uses all these indirectly
	Main()
}
