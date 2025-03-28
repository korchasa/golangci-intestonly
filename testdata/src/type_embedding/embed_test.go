package type_embedding

import "testing"

// TestType embeds the TestOnlyEmbedded type
type TestType struct {
	TestOnlyEmbedded // Embedded type only used in tests
	Name             string
}

// TestMethod shows usage of the embedded type's method
func (t TestType) GetEmbeddedValue() string {
	return t.TestMethod() // Uses method from embedded type
}

func TestTypeEmbedding(t *testing.T) {
	// Test the test-only embedded type directly
	testEmbedded := TestOnlyEmbedded{TestField: "embedded test value"}
	embeddedResult := testEmbedded.TestMethod()
	if embeddedResult != "embedded test value" {
		t.Errorf("Expected 'embedded test value', got '%s'", embeddedResult)
	}

	// Test the test type that embeds the test-only embedded type
	testType := TestType{
		TestOnlyEmbedded: TestOnlyEmbedded{TestField: "from test type"},
		Name:             "Test",
	}

	embeddedMethodResult := testType.TestMethod() // Directly accessing embedded method
	if embeddedMethodResult != "from test type" {
		t.Errorf("Expected 'from test type', got '%s'", embeddedMethodResult)
	}

	getValueResult := testType.GetEmbeddedValue()
	if getValueResult != "from test type" {
		t.Errorf("Expected 'from test type', got '%s'", getValueResult)
	}

	// Also test regular types with embedding
	usedType := UsedType{
		UsedEmbedded: UsedEmbedded{Field: "used field"},
		ID:           123,
	}

	usedMethodResult := usedType.UsedMethod() // Directly accessing embedded method
	if usedMethodResult != "used field" {
		t.Errorf("Expected 'used field', got '%s'", usedMethodResult)
	}

	getFieldResult := usedType.GetField()
	if getFieldResult != "used field" {
		t.Errorf("Expected 'used field', got '%s'", getFieldResult)
	}
}
