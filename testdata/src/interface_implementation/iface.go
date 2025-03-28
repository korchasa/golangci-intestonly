package interface_implementation

// TestOnlyInterface is an interface only used in tests
type TestOnlyInterface interface { // want "type 'TestOnlyInterface' is only used in tests"
	DoTest() string
}

// ProductionInterface is used in production code
type ProductionInterface interface {
	DoProd() string
}

// Both interfaces implemented together
type TestDoubleInterface interface {
	TestOnlyInterface
	ProductionInterface
}

// TestOnlyImplementer implements a test-only interface
type TestOnlyImplementer struct { // want "type 'TestOnlyImplementer' is only used in tests"
	Value string
}

// DoTest implements TestOnlyInterface
func (t TestOnlyImplementer) DoTest() string { // want "method 'DoTest' is only used in tests"
	return "test: " + t.Value
}

// ProductionImplementer implements the production interface
type ProductionImplementer struct {
	Data string
}

// DoProd implements ProductionInterface
func (p ProductionImplementer) DoProd() string {
	return "prod: " + p.Data
}

// DualImplementer implements both interfaces but is only used in tests
type DualImplementer struct { // want "type 'DualImplementer' is only used in tests"
	Name string
}

// DoTest implements TestOnlyInterface
func (d DualImplementer) DoTest() string { // want "method 'DoTest' is only used in tests"
	return "test dual: " + d.Name
}

// DoProd implements ProductionInterface
func (d DualImplementer) DoProd() string { // want "method 'DoProd' is only used in tests"
	return "prod dual: " + d.Name
}

// Helper function to use the production interface
func UseProductionInterface(p ProductionInterface) string {
	return p.DoProd()
}
