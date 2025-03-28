package false_negatives

import "fmt"

// IndirectlyUsedFunction appears to be test-only but is used indirectly
func IndirectlyUsedFunction() string {
	return "indirect"
}

// UsedViaReflection is used via reflection in non-test code
func UsedViaReflection() string {
	return "reflection"
}

// UsedInStringLiteral is referenced in a string literal in non-test code
func UsedInStringLiteral() string {
	return "string literal"
}

// DynamicallyUsed is used dynamically through function variables
func DynamicallyUsed() string {
	return "dynamic"
}

// Main function that uses all these functions in non-obvious ways
func Main() {
	// Indirect usage
	fn := getFunction()
	result := fn()
	fmt.Println(result)

	// Dynamic usage
	dynamicFn := getFunctionByName("DynamicallyUsed")
	if dynamicFn != nil {
		fmt.Println(dynamicFn())
	}

	// String literal usage - should prevent detection
	code := `
		// This code uses UsedInStringLiteral() in a comment
		function callIt() {
			return UsedInStringLiteral();
		}
	`
	fmt.Println(code)

	// Reflection is handled in getFunctionByName
}

// Helper functions
func getFunction() func() string {
	return IndirectlyUsedFunction
}

// Use function name in string, simulating reflection
func getFunctionByName(name string) func() string {
	if name == "DynamicallyUsed" {
		return DynamicallyUsed
	}
	if name == "UsedViaReflection" {
		return UsedViaReflection
	}
	return nil
}
