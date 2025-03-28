package nested_structures

import "testing"

func TestNestedStructures(t *testing.T) {
	// Test using the inner struct that is only used in tests
	outer := OuterType{
		Field: "test",
		Inner: struct{ Value int }{
			Value: 42,
		},
	}
	if outer.Inner.Value != 42 {
		t.Errorf("Expected Inner.Value to be 42, got %d", outer.Inner.Value)
	}

	// Test nested struct with test-only nested field
	nested := NestedStruct{
		Name: "test nested",
		UsedNested: struct{ ID int }{
			ID: 123,
		},
		TestOnlyNested: struct{ Value string }{
			Value: "test only",
		},
	}
	if nested.TestOnlyNested.Value != "test only" {
		t.Errorf("Expected TestOnlyNested.Value to be 'test only', got '%s'", nested.TestOnlyNested.Value)
	}

	// Also test the GetID method
	if nested.GetID() != 123 {
		t.Errorf("Expected GetID() to return 123, got %d", nested.GetID())
	}

	// Test the test-only method
	outerWithMethod := NewOuter()
	testMethodResult := outerWithMethod.TestOnlyMethod()
	if testMethodResult != "test: value" {
		t.Errorf("Expected TestOnlyMethod to return 'test: value', got '%s'", testMethodResult)
	}

	// Also test the regular method
	regularMethodResult := outerWithMethod.RegularMethod()
	if regularMethodResult != "value" {
		t.Errorf("Expected RegularMethod to return 'value', got '%s'", regularMethodResult)
	}
}
