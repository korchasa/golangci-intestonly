package p

import "testing"

func TestFalsePositives(t *testing.T) {
	// Test function usage
	if commonFunction() != "common" {
		t.Error("unexpected function result")
	}

	// Test type usage
	c1 := CommonType{Field: "common"}
	if c1.Field != "common" {
		t.Error("unexpected type field value")
	}

	// Test constant usage
	if commonConstant != "common" {
		t.Error("unexpected constant value")
	}
}
