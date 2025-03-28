package false_positives

import "testing"

func TestSharedCode(t *testing.T) {
	// Using shared function
	result := SharedFunction()
	if result != "shared" {
		t.Errorf("Expected 'shared', got '%s'", result)
	}

	// Using shared type
	helper := HelperType{
		ID:   2,
		Name: "Test",
	}
	if helper.ID != 2 {
		t.Errorf("Expected ID 2, got %d", helper.ID)
	}

	// Using shared method
	methodResult := helper.Method()
	if methodResult != "Test" {
		t.Errorf("Expected 'Test', got '%s'", methodResult)
	}

	// Using shared constant
	if CONSTANT != "shared constant" {
		t.Errorf("Expected 'shared constant', got '%s'", CONSTANT)
	}

	// Using shared variable
	if sharedVariable != "shared variable" {
		t.Errorf("Expected 'shared variable', got '%s'", sharedVariable)
	}

	// Also test UserFacing function
	user := UserFacing()
	if user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}
}
