package p

// Test case for interfaces
type TestInterface interface {
	InterfaceMethod() string
}

type TestInterfaceImpl struct{}

func (t *TestInterfaceImpl) InterfaceMethod() string {
	return "interface"
}

// Test case for type aliases
type TestAlias = string

func aliasFunction() TestAlias {
	return "alias"
}

// Test case for different access modifiers
type testPrivateType struct {
	field string
}

func (t *testPrivateType) privateMethod() string {
	return "private"
}

func (t *testPrivateType) PublicMethod() string {
	return "public"
}
