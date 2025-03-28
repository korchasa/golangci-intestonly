package nested_structures

// OuterType is used in non-test code
type OuterType struct {
	Field string

	// InnerType is only used in tests
	Inner struct { // want "type 'Inner' is only used in tests"
		Value int
	}
}

// OuterFunction is used in non-test code
func OuterFunction() string {
	return "outer"
}

// GetOuter returns an instance of OuterType
func GetOuter() OuterType {
	return OuterType{
		Field: OuterFunction(),
	}
}

// NestedStruct with both used and unused nested types
type NestedStruct struct {
	Name string

	// This nested type is used in regular code
	UsedNested struct {
		ID int
	}

	// This nested type is only used in tests
	TestOnlyNested struct { // want "type 'TestOnlyNested' is only used in tests"
		Value string
	}
}

// CreateNested creates a NestedStruct with the used nested type initialized
func CreateNested() NestedStruct {
	return NestedStruct{
		Name: "test",
		UsedNested: struct{ ID int }{
			ID: 1,
		},
	}
}

// Method that uses the regular nested type
func (n NestedStruct) GetID() int {
	return n.UsedNested.ID
}

// OuterWithTestMethod has a method that's only used in tests
type OuterWithTestMethod struct {
	Value string
}

// Used in regular code
func (o OuterWithTestMethod) RegularMethod() string {
	return o.Value
}

// Only used in tests
func (o OuterWithTestMethod) TestOnlyMethod() string { // want "method 'TestOnlyMethod' is only used in tests"
	return "test: " + o.Value
}

// Factory function
func NewOuter() OuterWithTestMethod {
	o := OuterWithTestMethod{Value: "value"}
	_ = o.RegularMethod()
	return o
}
