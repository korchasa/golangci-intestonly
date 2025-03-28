package complex_detection

// GlobalVariable - global variable that can be shadowed
// want "identifier .GlobalVariable. is only used in test files but is not part of test files"
var GlobalVariable = "global value"

// GlobalFunction - global function that can be shadowed
// want "identifier .GlobalFunction. is only used in test files but is not part of test files"
func GlobalFunction() string {
	return "global function"
}

// GlobalType - global type that can be shadowed
// want "identifier .GlobalType. is only used in test files but is not part of test files"
type GlobalType struct {
	Field string
}

// GlobalMethod - method of global type
// want "identifier .GlobalMethod. is only used in test files but is not part of test files"
func (g *GlobalType) GlobalMethod() string {
	return g.Field
}

// ShadowingContainer contains fields with the same names as global identifiers
// want "identifier .ShadowingContainer. is only used in test files but is not part of test files"
type ShadowingContainer struct {
	// Shadows global variable
	GlobalVariable string

	// Shadows global type
	GlobalType int
}

// ShadowingFunction shadows global identifiers with local variables
// want "identifier .ShadowingFunction. is only used in test files but is not part of test files"
func ShadowingFunction(GlobalVariable string) string {
	// Local variable shadows function parameter
	GlobalVariable = "local in function"

	// Shadow global function with local variable
	GlobalFunction := func() string {
		return "local function"
	}

	// Shadow global type with local variable
	GlobalType := struct {
		Value int
	}{
		Value: 100,
	}

	return GlobalVariable + " " + GlobalFunction() + " " + string(rune(GlobalType.Value))
}

// NestedShadowing demonstrates shadowing in different scopes
// want "identifier .NestedShadowing. is only used in test files but is not part of test files"
func NestedShadowing() string {
	// Shadow global variable at top level
	GlobalVariable := "outer"

	// Create anonymous function with shadowing
	f := func() string {
		// Shadow variable from outer scope
		GlobalVariable := "inner"
		return GlobalVariable
	}

	// Another level of nesting
	{
		// Shadow in block
		GlobalVariable := "block"
		_ = GlobalVariable // Use variable to avoid warning
	}

	return GlobalVariable + " " + f()
}

// NotShadowed uses global identifiers without shadowing
// want "identifier .NotShadowed. is only used in test files but is not part of test files"
func NotShadowed() string {
	// Use global identifiers directly
	g := &GlobalType{Field: "direct access"}
	return GlobalVariable + " " + GlobalFunction() + " " + g.GlobalMethod()
}
