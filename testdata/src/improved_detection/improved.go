package improved_detection

import "fmt"

// TestOnlyFunctionWithComment is a function that is only used in tests
// This comment also mentions TestOnlyFunctionWithComment for self-reference
func TestOnlyFunctionWithComment() string { // want "function 'TestOnlyFunctionWithComment' is only used in tests"
	return "test function with comment"
}

// ComplexUsage has a complex usage pattern
func ComplexUsage() string {
	return "complex usage"
}

// ReflectivelyUsed is used through reflection techniques
func ReflectivelyUsed() string {
	return "reflectively used"
}

// StringUsed is referenced in string concatenation
func StringUsed() string {
	return "string used"
}

// ActuallyUsedInCode is used in regular code through indirection
func ActuallyUsedInCode() string {
	return "actually used"
}

// UseComplexPattern demonstrates complex usage patterns
func UseComplexPattern() {
	// Use function indirectly through map lookup
	funcs := map[string]func() string{
		"complex": ComplexUsage,
	}

	fn := funcs["complex"]
	if fn != nil {
		_ = fn()
	}

	// Use function name in string concatenation
	funcName := "StringUsed"
	code := fmt.Sprintf("Function %s() returns a string value", funcName)
	_ = code

	// Complex indirection
	getActualFunc := func() func() string {
		return ActuallyUsedInCode
	}
	actualFn := getActualFunc()
	_ = actualFn()

	// Reflection-like indirection
	funcMap := make(map[string]interface{})
	funcMap["ReflectivelyUsed"] = ReflectivelyUsed
	if fn, ok := funcMap["ReflectivelyUsed"].(func() string); ok {
		_ = fn()
	}
}
