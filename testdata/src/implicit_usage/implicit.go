package implicit_usage

import "runtime"

// IndirectlyUsed is a function that is indirectly used through a function variable
func IndirectlyUsed() string {
	return "indirectly used"
}

// AssignedToVar is assigned to a variable and then used
func AssignedToVar() string {
	return "assigned to var"
}

// UsedViaFuncMap is used as an entry in a map of functions
func UsedViaFuncMap() string {
	return "via func map"
}

// GetFunctions returns a map of functions
func GetFunctions() map[string]func() string {
	return map[string]func() string{
		"map_func": UsedViaFuncMap,
	}
}

// UseFunctions demonstrates different ways functions can be used indirectly
func UseFunctions() {
	// Assign function to variable and then call
	fn := AssignedToVar
	result := fn()
	_ = result

	// Use map of functions
	funcs := GetFunctions()
	if f, ok := funcs["map_func"]; ok {
		_ = f()
	}

	// Using the runtime package to get function name indirectly
	pc, _, _, _ := runtime.Caller(0)
	_ = runtime.FuncForPC(pc).Name() // This is how reflection-based systems might detect function names
}
