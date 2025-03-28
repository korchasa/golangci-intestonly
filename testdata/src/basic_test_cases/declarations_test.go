package basic_test_cases

import "testing"

func TestUsage(t *testing.T) {
	// Use the test-only function
	result := TestOnlyFunction()
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Use anotherTestOnlyFunction
	answer := anotherTestOnlyFunction()
	if answer != 42 {
		t.Errorf("Expected 42, got %d", answer)
	}

	// Use TestOnlyType
	testType := TestOnlyType{Value: "test value"}
	if testType.Value != "test value" {
		t.Errorf("Expected 'test value', got '%s'", testType.Value)
	}

	// Use TestOnlyConstant
	if TestOnlyConstant != "test constant" {
		t.Errorf("Expected 'test constant', got '%s'", TestOnlyConstant)
	}

	// Use TestOnlyVariable
	if TestOnlyVariable != "test variable" {
		t.Errorf("Expected 'test variable', got '%s'", TestOnlyVariable)
	}

	// Also use the normal function that's used in both test and non-test code
	used := UsedFunction()
	if used != "used" {
		t.Errorf("Expected 'used', got '%s'", used)
	}
}
