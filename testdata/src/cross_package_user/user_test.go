package cross_package_user

import (
	"testing"

	"cross_package_ref"
)

func TestCrossPackageRefs(t *testing.T) {
	// Use the exported test-only function
	result := cross_package_ref.ExportedTestFunc()
	if result != "exported test function" {
		t.Errorf("Expected 'exported test function', got '%s'", result)
	}

	// Use the test-only type
	testType := cross_package_ref.TestOnlyType{
		Name: "test name",
	}
	if testType.Name != "test name" {
		t.Errorf("Expected name 'test name', got '%s'", testType.Name)
	}

	// Use the test-only constant
	if cross_package_ref.TestOnlyConst != "test-only constant" {
		t.Errorf("Expected 'test-only constant', got '%s'", cross_package_ref.TestOnlyConst)
	}

	// Use the test-only variable
	if cross_package_ref.TestOnlyVar != "test-only variable" {
		t.Errorf("Expected 'test-only variable', got '%s'", cross_package_ref.TestOnlyVar)
	}

	// Also test the regular exported functions
	usedResult := UseExportedFunc()
	if usedResult != "Using: exported used function" {
		t.Errorf("Expected 'Using: exported used function', got '%s'", usedResult)
	}

	typeResult := UseType()
	if typeResult != 42 {
		t.Errorf("Expected 42, got %d", typeResult)
	}

	constResult := UseConst()
	if constResult != "Constant: used constant" {
		t.Errorf("Expected 'Constant: used constant', got '%s'", constResult)
	}

	varResult := UseVar()
	if varResult != "Variable: used variable" {
		t.Errorf("Expected 'Variable: used variable', got '%s'", varResult)
	}
}
