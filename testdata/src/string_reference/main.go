// Package string_reference demonstrates functions that appear in string literals
package string_reference

import (
	"fmt"
	"reflect"
)

// StringRefFunction is a function that is referenced in a string literal
// and only explicitly called in tests
func StringRefFunction() string {
	return "I appear in strings but am only called in tests"
}

// CommentRefFunction is only used in comments and tests
func CommentRefFunction() string {
	return "I'm only used in comments and tests"
}

// ReflectionRefFunction is referenced via reflection
func ReflectionRefFunction() string {
	return "I'm used via reflection"
}

// UseStringsAndReflection mentions functions in strings and uses reflection
func UseStringsAndReflection() {
	// The following calls use strings and reflection
	functionName := "StringRefFunction"
	fmt.Println("Function name:", functionName)

	// This calls the function via reflection
	reflectFunc := reflect.ValueOf(ReflectionRefFunction)
	result := reflectFunc.Call(nil)[0].String()
	fmt.Println("Result from reflection:", result)

	// This references CommentRefFunction in a comment but doesn't call it:
	// Example usage: CommentRefFunction()
}
