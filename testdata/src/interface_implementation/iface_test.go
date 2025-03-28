package interface_implementation

import "testing"

func TestInterfaceImplementation(t *testing.T) {
	// Test the test-only interface and implementer
	var testIface TestOnlyInterface
	testImpl := TestOnlyImplementer{Value: "test value"}
	testIface = testImpl

	testResult := testIface.DoTest()
	if testResult != "test: test value" {
		t.Errorf("Expected 'test: test value', got '%s'", testResult)
	}

	// Test the production interface and implementer
	var prodIface ProductionInterface
	prodImpl := ProductionImplementer{Data: "prod data"}
	prodIface = prodImpl

	prodResult := prodIface.DoProd()
	if prodResult != "prod: prod data" {
		t.Errorf("Expected 'prod: prod data', got '%s'", prodResult)
	}

	// Test the dual implementer that implements both interfaces
	var dualTestIface TestOnlyInterface
	var dualProdIface ProductionInterface
	dualImpl := DualImplementer{Name: "dual"}

	dualTestIface = dualImpl
	dualProdIface = dualImpl

	dualTestResult := dualTestIface.DoTest()
	if dualTestResult != "test dual: dual" {
		t.Errorf("Expected 'test dual: dual', got '%s'", dualTestResult)
	}

	dualProdResult := dualProdIface.DoProd()
	if dualProdResult != "prod dual: dual" {
		t.Errorf("Expected 'prod dual: dual', got '%s'", dualProdResult)
	}

	// Test using the helper function
	useResult := UseProductionInterface(prodImpl)
	if useResult != "prod: prod data" {
		t.Errorf("Expected 'prod: prod data', got '%s'", useResult)
	}

	// Test using the helper function with the dual implementer
	dualUseResult := UseProductionInterface(dualImpl)
	if dualUseResult != "prod dual: dual" {
		t.Errorf("Expected 'prod dual: dual', got '%s'", dualUseResult)
	}
}
