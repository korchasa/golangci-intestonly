package true_positives

// TestOnlyExportedFunction is an exported function only used in tests
func TestOnlyExportedFunction() string { // want "function 'TestOnlyExportedFunction' is only used in tests"
	return "test export"
}

// testOnlyUnexportedFunction is an unexported function only used in tests
func testOnlyUnexportedFunction() string { // want "function 'testOnlyUnexportedFunction' is only used in tests"
	return "test unexport"
}

// ExportedTestType is a type only used in tests
type ExportedTestType struct { // want "type 'ExportedTestType' is only used in tests"
	Name string
	Age  int
}

// unexportedTestType is an unexported type only used in tests
type unexportedTestType struct { // want "type 'unexportedTestType' is only used in tests"
	field string
}

// Method on test-only type
func (e ExportedTestType) Method() string { // want "method 'Method' is only used in tests"
	return e.Name
}

// TestOnlyConst is only used in tests
const TestOnlyConst = "test constant" // want "const 'TestOnlyConst' is only used in tests"

// testOnlyVar is only used in tests
var testOnlyVar = 42 // want "variable 'testOnlyVar' is only used in tests"

// Regular function used in non-test code
func NormalFunction() string {
	return "normal"
}

// Function used in actual implementation
func ActualFunction() string {
	return NormalFunction()
}
