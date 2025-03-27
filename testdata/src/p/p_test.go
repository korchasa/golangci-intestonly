package p

import "testing"

func TestHelperFunction(t *testing.T) {
	result := helperFunction()
	if result != "helper" {
		t.Errorf("Expected 'helper', got '%s'", result)
	}
}
