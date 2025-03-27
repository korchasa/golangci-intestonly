package p

import "testing"

func TestNestedStructures(t *testing.T) {
	// Test nested structure usage
	outer := &OuterStruct{
		InnerStruct: InnerStruct{Field: "test"},
	}
	if outer.Field != "test" {
		t.Error("unexpected nested field value")
	}

	// Test nested method usage
	if outer.outerMethod() != "outer" {
		t.Error("unexpected outer method result")
	}
	if outer.innerMethod() != "inner" {
		t.Error("unexpected inner method result")
	}

	// Test embedded type usage
	embedded := &EmbeddedType{}
	if embedded.embeddedMethod() != "embedded" {
		t.Error("unexpected embedded method result")
	}
}
