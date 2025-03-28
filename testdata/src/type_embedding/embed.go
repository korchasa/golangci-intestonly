package type_embedding

// TestOnlyEmbedded is a type that is only embedded in test types
type TestOnlyEmbedded struct { // want "type 'TestOnlyEmbedded' is only used in tests"
	TestField string
}

// TestMethod is a method on the test-only embedded type
func (t TestOnlyEmbedded) TestMethod() string { // want "method 'TestMethod' is only used in tests"
	return t.TestField
}

// UsedEmbedded is used in both test and non-test code
type UsedEmbedded struct {
	Field string
}

// UsedMethod is used in both contexts
func (u UsedEmbedded) UsedMethod() string {
	return u.Field
}

// UsedType embeds a type that is used in production code
type UsedType struct {
	UsedEmbedded // Embedded type used in production
	ID           int
}

// GetField demonstrates usage of the embedded type's method
func (u UsedType) GetField() string {
	return u.UsedMethod() // Uses method from embedded type
}
