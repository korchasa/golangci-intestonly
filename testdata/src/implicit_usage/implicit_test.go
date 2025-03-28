package implicit_usage

import "testing"

func TestImplicitUsage(t *testing.T) {
	// Test IndirectlyUsed by assigning to a function variable
	var fn func() string = IndirectlyUsed
	result := fn()
	if result != "indirectly used" {
		t.Errorf("Expected 'indirectly used', got '%s'", result)
	}

	// Test normal functions
	aResult := AssignedToVar()
	if aResult != "assigned to var" {
		t.Errorf("Expected 'assigned to var', got '%s'", aResult)
	}

	mResult := UsedViaFuncMap()
	if mResult != "via func map" {
		t.Errorf("Expected 'via func map', got '%s'", mResult)
	}

	// Test the GetFunctions function
	funcs := GetFunctions()
	if f, ok := funcs["map_func"]; ok {
		fResult := f()
		if fResult != "via func map" {
			t.Errorf("Expected 'via func map', got '%s'", fResult)
		}
	} else {
		t.Error("Expected 'map_func' key to exist in funcs map")
	}

	// Test UseFunctions
	UseFunctions()
}
