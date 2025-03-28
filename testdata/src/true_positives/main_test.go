package true_positives

import "testing"

func TestTruePositives(t *testing.T) {
	// Using exported test-only function
	result := TestOnlyExportedFunction()
	if result != "test export" {
		t.Errorf("Expected 'test export', got '%s'", result)
	}

	// Using unexported test-only function
	unexpResult := testOnlyUnexportedFunction()
	if unexpResult != "test unexport" {
		t.Errorf("Expected 'test unexport', got '%s'", unexpResult)
	}

	// Using exported test-only type
	testType := ExportedTestType{
		Name: "Test",
		Age:  30,
	}
	if testType.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", testType.Name)
	}

	// Using method on test-only type
	methodResult := testType.Method()
	if methodResult != "Test" {
		t.Errorf("Expected 'Test', got '%s'", methodResult)
	}

	// Using unexported test-only type
	unexpType := unexportedTestType{
		field: "unexported",
	}
	if unexpType.field != "unexported" {
		t.Errorf("Expected 'unexported', got '%s'", unexpType.field)
	}

	// Using test-only constant
	if TestOnlyConst != "test constant" {
		t.Errorf("Expected 'test constant', got '%s'", TestOnlyConst)
	}

	// Using test-only variable
	if testOnlyVar != 42 {
		t.Errorf("Expected 42, got %d", testOnlyVar)
	}

	// Also using normal function to make sure it's not reported
	normal := NormalFunction()
	if normal != "normal" {
		t.Errorf("Expected 'normal', got '%s'", normal)
	}
}
