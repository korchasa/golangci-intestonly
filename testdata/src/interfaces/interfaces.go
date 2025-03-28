package interfaces

// TestOnlyInterface is an interface only used in tests
type TestOnlyInterface interface { // want "type 'TestOnlyInterface' is only used in tests"
	TestMethod() string
}

// RegularInterface is used in both test and non-test code
type RegularInterface interface {
	RegularMethod() string
}

// TestOnlyImplementation implements TestOnlyInterface but is only used in tests
type TestOnlyImplementation struct { // want "type 'TestOnlyImplementation' is only used in tests"
	Data string
}

// TestMethod implements TestOnlyInterface
func (t TestOnlyImplementation) TestMethod() string { // want "method 'TestMethod' is only used in tests"
	return t.Data
}

// RegularImplementation implements RegularInterface
type RegularImplementation struct {
	Value string
}

// RegularMethod implements RegularInterface
func (r RegularImplementation) RegularMethod() string {
	return r.Value
}

// TestOnlyMethod is only used in tests
func (r RegularImplementation) TestOnlyMethod() string { // want "method 'TestOnlyMethod' is only used in tests"
	return "test: " + r.Value
}

// Used in regular code
func UseRegularInterface(i RegularInterface) string {
	return "Using: " + i.RegularMethod()
}

// Factory function for RegularImplementation
func NewRegularImplementation(value string) RegularImplementation {
	return RegularImplementation{Value: value}
}
