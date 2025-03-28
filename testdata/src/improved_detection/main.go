// Package improved_detection provides test cases for improved usage detection.
package improved_detection

// ReflectionTarget is a struct that might be used via reflection.
// want "identifier .ReflectionTarget. is only used in test files but is not part of test files"
type ReflectionTarget struct {
	Field string
}

// ReflectionMethod is a method that might be called via reflection.
// want "identifier .ReflectionMethod. is only used in test files but is not part of test files"
func (r *ReflectionTarget) ReflectionMethod() string {
	return r.Field
}

// EmbeddedType is a type that might be embedded in test-only structs.
// want "identifier .EmbeddedType. is only used in test files but is not part of test files"
type EmbeddedType struct {
	CommonField string
}

// EmbeddedMethod is a method on an embedded type.
// want "identifier .EmbeddedMethod. is only used in test files but is not part of test files"
func (e *EmbeddedType) EmbeddedMethod() string {
	return e.CommonField
}

// RegistryPattern demonstrates a common registry pattern where a function
// might register itself in a global map.
// want "identifier .RegistryPattern. is only used in test files but is not part of test files"
func RegistryPattern() {
	// This would typically register something, but for test purposes it's empty
}

// ShadowedIdentifier is a name that might be shadowed in local scopes.
// want "identifier .ShadowedIdentifier. is only used in test files but is not part of test files"
const ShadowedIdentifier = "global"

// InitFunction is used in an init() function in tests.
// want "identifier .InitFunction. is only used in test files but is not part of test files"
func InitFunction() {
	// Do some initialization
}

// Type checking:
type testInterface interface {
	TestMethod() string
}

// TestInterfaceImplementation implements the testInterface.
// This should only be detected if actually used.
// want "identifier .TestInterfaceImplementation. is only used in test files but is not part of test files"
type TestInterfaceImplementation struct{}

// TestMethod implements the testInterface.
// want "identifier .TestMethod. is only used in test files but is not part of test files"
func (t *TestInterfaceImplementation) TestMethod() string {
	return "test"
}
