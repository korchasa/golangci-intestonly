// Package implicit_usage demonstrates functionality used implicitly
package implicit_usage

import (
	"fmt"
	"strings"
)

// ImplicitlyUsedFunction is a function that is referenced by string literals
// in non-test code but only called explicitly in tests
func ImplicitlyUsedFunction() string {
	return "I'm implicitly used in non-test code"
}

// UseImplicitFunctions uses the function name in a string but doesn't call it
func UseImplicitFunctions() {
	// This function doesn't explicitly call ImplicitlyUsedFunction
	// but it mentions it in a string literal, which should be detected
	functionName := "ImplicitlyUsedFunction"
	fmt.Println("The function", functionName, "is used implicitly")

	// We also test with the function name in a larger string
	docs := `
	API Documentation:

	ImplicitlyUsedFunction() - Returns a string indicating implicit usage
	`

	if strings.Contains(docs, "ImplicitlyUsedFunction") {
		fmt.Println("Function is documented")
	}
}
