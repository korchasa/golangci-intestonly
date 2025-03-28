package reflection_usage

import (
	"reflect"
	"testing"
)

func TestReflectionUsage(t *testing.T) {
	// Test directly calling functions
	if UsedViaReflection() != "reflected" {
		t.Errorf("Expected 'reflected', got '%s'", UsedViaReflection())
	}

	if DirectlyUsed() != "direct" {
		t.Errorf("Expected 'direct', got '%s'", DirectlyUsed())
	}

	// Test the test-only function
	if OnlyInTests() != "tests only" {
		t.Errorf("Expected 'tests only', got '%s'", OnlyInTests())
	}

	if TestFunction("input") != "test: input" {
		t.Errorf("Expected 'test: input', got '%s'", TestFunction("input"))
	}

	// Test using reflection in tests too
	funcValue := reflect.ValueOf(UsedViaReflection)
	result := funcValue.Call(nil)
	if result[0].String() != "reflected" {
		t.Errorf("Expected 'reflected' via reflection, got '%s'", result[0].String())
	}

	// Test the FunctionContainer
	funcs := FunctionContainer{}
	testResult := funcs.TestFunction("hello")
	if testResult != "test: hello" {
		t.Errorf("Expected 'test: hello', got '%s'", testResult)
	}

	// Test the reflection system
	s := SystemWithReflection{}
	invokeResult := s.InvokeFunction("DirectlyUsed")
	if invokeResult.(string) != "direct" {
		t.Errorf("Expected 'direct', got '%s'", invokeResult)
	}

	// Test the UseReflection function
	UseReflection()
}
