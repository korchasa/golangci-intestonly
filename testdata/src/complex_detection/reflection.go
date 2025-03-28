// Package complex_detection tests complex detection scenarios.
package complex_detection

import (
	"encoding/json"
)

// ComplexReflectionStruct represents a structure for testing complex reflection cases
// want "identifier .ComplexReflectionStruct. is only used in test files but is not part of test files"
type ComplexReflectionStruct struct {
	Field1 string
	Field2 int
	inner  innerStruct
}

// innerStruct - nested structure
// want "identifier .innerStruct. is only used in test files but is not part of test files"
type innerStruct struct {
	Value string
}

// DynamicMethod - method that can be called through reflection
// want "identifier .DynamicMethod. is only used in test files but is not part of test files"
func (c *ComplexReflectionStruct) DynamicMethod(arg string) string {
	return arg + c.Field1
}

// GetInnerValue gets the value from the nested structure
// want "identifier .GetInnerValue. is only used in test files but is not part of test files"
func (c *ComplexReflectionStruct) GetInnerValue() string {
	return c.inner.Value
}

// GenericReflectionHandler - handler that uses reflection to work with different types
// want "identifier .GenericReflectionHandler. is only used in test files but is not part of test files"
func GenericReflectionHandler(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data)
}

// ReflectionWrapper hides the use of reflection behind an abstraction
// want "identifier .ReflectionWrapper. is only used in test files but is not part of test files"
type ReflectionWrapper struct {
	target interface{}
}

// CallMethod calls a method through reflection
// want "identifier .CallMethod. is only used in test files but is not part of test files"
func (r *ReflectionWrapper) CallMethod(methodName string, args ...interface{}) interface{} {
	// In real code, there would be an implementation using reflect.ValueOf().MethodByName()
	return nil
}
