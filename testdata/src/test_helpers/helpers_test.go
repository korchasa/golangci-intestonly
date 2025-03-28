package test_helpers

import "testing"

func TestHelpers(t *testing.T) {
	// Use AssertEqual
	if !AssertEqual("test", "test") {
		t.Error("Expected 'test' to equal 'test'")
	}

	// Use MockData
	mock := MockData{
		ID:   1,
		Name: "Test Mock",
	}
	if mock.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", mock.ID)
	}

	// Use FakeClient
	fake := FakeClient{URL: "http://example.com"}
	result := fake.Get()
	if result != "data from http://example.com" {
		t.Errorf("Expected 'data from http://example.com', got '%s'", result)
	}

	// Use TestSetup
	setupResult := TestSetup()
	if setupResult != "test environment" {
		t.Errorf("Expected 'test environment', got '%s'", setupResult)
	}

	// Use testCleanup
	testCleanup()

	// Also use the regular function
	used := UsedFunction()
	if used != "used" {
		t.Errorf("Expected 'used', got '%s'", used)
	}
}
