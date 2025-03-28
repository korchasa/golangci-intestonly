package complex_detection

import (
	"testing"
)

func TestGlobalAccess(t *testing.T) {
	// Check direct access to global identifiers
	t.Logf("Global variable: %s", GlobalVariable)
	t.Logf("Global function: %s", GlobalFunction())

	g := &GlobalType{Field: "test field"}
	t.Logf("Global type method: %s", g.GlobalMethod())
}

func TestShadowingTypes(t *testing.T) {
	// Check shadowing of types in containers
	container := &ShadowingContainer{
		GlobalVariable: "shadowed variable",
		GlobalType:     42,
	}

	// Access to shadowed fields
	t.Logf("Container fields: %s, %d", container.GlobalVariable, container.GlobalType)

	// Parallel use of global identifiers and shadowed fields
	t.Logf("Global and shadowed: %s vs %s", GlobalVariable, container.GlobalVariable)
}

func TestShadowingFunctions(t *testing.T) {
	// Check shadowing in functions
	result := ShadowingFunction("parameter")
	t.Logf("Shadowing function result: %s", result)
}

func TestNestedShadowing(t *testing.T) {
	// Check nested shadowing
	result := NestedShadowing()
	t.Logf("Nested shadowing result: %s", result)
}

func TestDirectAccessWithShadowing(t *testing.T) {
	// Local shadowing in the test
	GlobalVariable := "test local"
	GlobalFunction := func() string { return "test function" }

	// Use both local and global versions in parallel
	t.Logf("Local shadowed: %s, %s", GlobalVariable, GlobalFunction())

	// To access global versions, qualified names must be used
	// In tests this is more difficult since we're in the same package, so we use
	// a special function that accesses global versions
	global := NotShadowed()
	t.Logf("Global via function: %s", global)
}

func TestComplexShadowing(t *testing.T) {
	// Create a local variable with the same name as the global type
	GlobalType := "string value"

	// Use local and global value
	t.Logf("Local string: %s", GlobalType)

	// To access the global type, we use explicit casting
	g := &(struct {
		Field string
	}{"field value"})

	t.Logf("Struct field: %s", g.Field)

	// Complex shadowing with passing shadowed variables to other functions
	localVar := GlobalVariable // Not shadowing, but copying the global variable
	func() {
		// Shadowing within an anonymous function
		GlobalVariable := "anonymous"
		t.Logf("In anonymous function: %s, outer: %s", GlobalVariable, localVar)
	}()
}
