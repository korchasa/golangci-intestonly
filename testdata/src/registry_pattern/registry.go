package registry_pattern

import "fmt"

// Registry of handlers/providers
var registry = make(map[string]interface{})

// RegisterHandler registers a handler function
func RegisterHandler(name string, handler interface{}) {
	registry[name] = handler
}

// GetHandler returns a handler by name
func GetHandler(name string) interface{} {
	return registry[name]
}

// UsedInProdHandler is a handler used in production
func UsedInProdHandler(data string) string {
	return "prod: " + data
}

// TestOnlyHandler is only used in tests
func TestOnlyHandler(data string) string { // want "function 'TestOnlyHandler' is only used in tests"
	return "test: " + data
}

// Register the production handler
func init() {
	RegisterHandler("prod", UsedInProdHandler)
}

// UseRegisteredHandler demonstrates using a handler from the registry
func UseRegisteredHandler() {
	if handler, ok := GetHandler("prod").(func(string) string); ok {
		result := handler("sample")
		fmt.Println(result)
	}
}
