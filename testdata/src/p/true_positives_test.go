package p

import "testing"

func TestTruePositives(t *testing.T) {
	// Test function usage
	if testOnlyFunction() != "test" {
		t.Error("unexpected function result")
	}

	// Test type usage
	t1 := TestOnlyType{Field: "test"}
	if t1.Field != "test" {
		t.Error("unexpected type field value")
	}

	// Test constant usage
	if testOnlyConstant != "test" {
		t.Error("unexpected constant value")
	}
}
