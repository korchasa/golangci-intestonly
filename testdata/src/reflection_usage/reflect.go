package reflection_usage

import (
	"fmt"
	"reflect"
)

// UsedViaReflection is detected through reflection analysis
func UsedViaReflection() string {
	return "reflected"
}

// DirectlyUsed is a function that's directly called
func DirectlyUsed() string {
	return "direct"
}

// OnlyInTests is a function only used in tests
func OnlyInTests() string { // want "function 'OnlyInTests' is only used in tests"
	return "tests only"
}

// TestFunction looks like a test function but is used in reflection
func TestFunction(s string) string {
	return "test: " + s
}

// UseReflection demonstrates using reflection to call functions
func UseReflection() {
	// Get function by name using reflection
	funcValue := reflect.ValueOf(UsedViaReflection)
	result := funcValue.Call(nil)
	fmt.Println(result[0].String())

	// Lookup function by name on a struct
	funcs := FunctionContainer{}
	method := reflect.ValueOf(funcs).MethodByName("TestFunction")
	if method.IsValid() {
		args := []reflect.Value{reflect.ValueOf("hello")}
		resultValues := method.Call(args)
		fmt.Println(resultValues[0].String())
	}

	// Using a function map (simpler reflection)
	s := SystemWithReflection{}
	s.InvokeFunction("DirectlyUsed")
}

// FunctionContainer holds functions that might be called via reflection
type FunctionContainer struct{}

// TestFunction is a method that might be called via reflection
func (FunctionContainer) TestFunction(s string) string {
	return TestFunction(s)
}

// SystemWithReflection simulates a system that uses reflection
type SystemWithReflection struct{}

// InvokeFunction invokes a function by name
func (s SystemWithReflection) InvokeFunction(name string) interface{} {
	functions := map[string]interface{}{
		"DirectlyUsed": DirectlyUsed,
	}

	if fn, ok := functions[name]; ok {
		if f, ok := fn.(func() string); ok {
			return f()
		}
	}
	return nil
}
