package string_reference

import "testing"

func TestStringReferences(t *testing.T) {
	// Test the functions directly
	if ReferencedInString() != "referenced in string" {
		t.Errorf("Expected 'referenced in string', got '%s'", ReferencedInString())
	}

	if UsedNormally() != "used normally" {
		t.Errorf("Expected 'used normally', got '%s'", UsedNormally())
	}

	if OnlyInTestString() != "only in test string" {
		t.Errorf("Expected 'only in test string', got '%s'", OnlyInTestString())
	}

	if TestOnlyFunc() != "test only" {
		t.Errorf("Expected 'test only', got '%s'", TestOnlyFunc())
	}

	// Test the string reference
	stringRef := UseFunctionInString()
	if stringRef != "This function calls ReferencedInString() to get data" {
		t.Errorf("Expected string reference, got '%s'", stringRef)
	}

	// Also reference OnlyInTestString in a string
	testStringRef := "This test uses OnlyInTestString() for testing"
	if len(testStringRef) == 0 {
		t.Error("String reference shouldn't be empty")
	}

	// Test normal function usage
	normalResult := UseNormalFunc()
	if normalResult != "used normally" {
		t.Errorf("Expected 'used normally', got '%s'", normalResult)
	}
}
