package improved_detection

import (
	"reflect"
	"testing"
)

// Global registry for testing
var registry = make(map[string]interface{})

// Test setup with init
func init() {
	// Use the InitFunction in an init context
	InitFunction()
}

func TestReflection(t *testing.T) {
	// Use reflection to access a type
	target := &ReflectionTarget{Field: "test value"}

	// Get type info via reflection
	rt := reflect.TypeOf(target)
	t.Logf("Type name: %s", rt.Elem().Name())

	// Get method via reflection
	method, _ := rt.MethodByName("ReflectionMethod")
	result := method.Func.Call([]reflect.Value{reflect.ValueOf(target)})
	t.Logf("Method result: %s", result[0].String())
}

func TestEmbedding(t *testing.T) {
	// Use embedding to test the embedded type
	type TestStruct struct {
		EmbeddedType // Embedded
		ExtraField   string
	}

	test := &TestStruct{
		EmbeddedType: EmbeddedType{CommonField: "common value"},
		ExtraField:   "extra value",
	}

	// Access embedded method
	if test.EmbeddedMethod() != "common value" {
		t.Error("Embedded method failed")
	}
}

func TestRegistry(t *testing.T) {
	// Register the function in a map
	registry["test_function"] = RegistryPattern
}

func TestShadowing(t *testing.T) {
	// Local variable with same name as global
	ShadowedIdentifier := "local"

	// Still referencing the global indirectly
	if ShadowedIdentifier != "local" {
		// This is a dummy condition just to use the global
		t.Log(ShadowedIdentifier)
	}
}

func TestInterfaceUsage(t *testing.T) {
	// Use the interface implementation
	var ti testInterface = &TestInterfaceImplementation{}

	if ti.TestMethod() != "test" {
		t.Error("Interface method implementation failed")
	}
}
