package p

// Test case for nested structures
type OuterStruct struct {
	InnerStruct
}

type InnerStruct struct {
	Field string
}

// Test case for nested methods
func (o *OuterStruct) outerMethod() string {
	return "outer"
}

func (i *InnerStruct) innerMethod() string {
	return "inner"
}

// Test case for embedded types
type EmbeddedType struct {
	string
}

func (e *EmbeddedType) embeddedMethod() string {
	return "embedded"
}
