package implicit_usage

import "testing"

func TestImplicitlyUsedFunction(t *testing.T) {
	result := ImplicitlyUsedFunction()
	if result != "I'm implicitly used in non-test code" {
		t.Errorf("Unexpected result: %s", result)
	}
}
