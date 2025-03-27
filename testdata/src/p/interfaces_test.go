package p

import "testing"

func TestInterfaces(t *testing.T) {
	// Test interface implementation
	var i TestInterface = &TestInterfaceImpl{}
	if i.InterfaceMethod() != "interface" {
		t.Error("unexpected interface method result")
	}

	// Test type alias usage
	if aliasFunction() != "alias" {
		t.Error("unexpected alias function result")
	}

	// Test access modifiers
	t1 := &testPrivateType{field: "test"}
	if t1.privateMethod() != "private" {
		t.Error("unexpected private method result")
	}
	if t1.PublicMethod() != "public" {
		t.Error("unexpected public method result")
	}
}
