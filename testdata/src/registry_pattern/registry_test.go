package registry_pattern

import "testing"

func init() {
	// Register test-only handler in init function
	RegisterHandler("test", TestOnlyHandler)
}

func TestRegistry(t *testing.T) {
	// Test the production handler
	prodHandler, ok := GetHandler("prod").(func(string) string)
	if !ok {
		t.Error("Expected to get production handler")
		return
	}

	prodResult := prodHandler("prod-data")
	if prodResult != "prod: prod-data" {
		t.Errorf("Expected 'prod: prod-data', got '%s'", prodResult)
	}

	// Test the test-only handler
	testHandler, ok := GetHandler("test").(func(string) string)
	if !ok {
		t.Error("Expected to get test handler")
		return
	}

	testResult := testHandler("test-data")
	if testResult != "test: test-data" {
		t.Errorf("Expected 'test: test-data', got '%s'", testResult)
	}

	// Also directly test the handlers
	directProdResult := UsedInProdHandler("direct")
	if directProdResult != "prod: direct" {
		t.Errorf("Expected 'prod: direct', got '%s'", directProdResult)
	}

	directTestResult := TestOnlyHandler("direct")
	if directTestResult != "test: direct" {
		t.Errorf("Expected 'test: direct', got '%s'", directTestResult)
	}

	// Test the function that uses the registry
	UseRegisteredHandler()
}
