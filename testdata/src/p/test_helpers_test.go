package p

import (
	"testing"
)

func TestUsingTestHelpers(t *testing.T) {
	// Use test helpers
	helper := NewTestHelper()
	if helper.Field1 != "test" {
		t.Errorf("Expected Field1 to be 'test', got %s", helper.Field1)
	}

	data := SetupTestData()
	if len(data) != 3 {
		t.Errorf("Expected 3 test data items, got %d", len(data))
	}

	mockDB := NewMockDB()
	mockDB.Data["key"] = "value"
	if mockDB.Data["key"] != "value" {
		t.Error("Expected mockDB to store value correctly")
	}

	// Use assertion helper
	AssertEqual(t, "test", helper.Field1)

	// Cleanup
	CleanupTestData()
}
