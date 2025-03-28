package complex_detection

import (
	"fmt"
	"sync"
)

// Handler - handler function type
// want "identifier .Handler. is only used in test files but is not part of test files"
type Handler func(data interface{}) (interface{}, error)

// Registry - global handler registry
// want "identifier .Registry. is only used in test files but is not part of test files"
var Registry = struct {
	sync.RWMutex
	handlers map[string]Handler
}{
	handlers: make(map[string]Handler),
}

// RegisterHandler registers a handler in the registry
// want "identifier .RegisterHandler. is only used in test files but is not part of test files"
func RegisterHandler(name string, handler Handler) {
	Registry.Lock()
	defer Registry.Unlock()
	Registry.handlers[name] = handler
}

// GetHandler returns a handler by name
// want "identifier .GetHandler. is only used in test files but is not part of test files"
func GetHandler(name string) (Handler, error) {
	Registry.RLock()
	defer Registry.RUnlock()
	handler, ok := Registry.handlers[name]
	if !ok {
		return nil, fmt.Errorf("handler %s not found", name)
	}
	return handler, nil
}

// StringHandler processes string data
// want "identifier .StringHandler. is only used in test files but is not part of test files"
func StringHandler(data interface{}) (interface{}, error) {
	if str, ok := data.(string); ok {
		return "Processed: " + str, nil
	}
	return nil, fmt.Errorf("expected string, got %T", data)
}

// IntHandler processes integer data
// want "identifier .IntHandler. is only used in test files but is not part of test files"
func IntHandler(data interface{}) (interface{}, error) {
	if num, ok := data.(int); ok {
		return num * 2, nil
	}
	return nil, fmt.Errorf("expected int, got %T", data)
}

// ExecuteHandler executes a handler by name
// want "identifier .ExecuteHandler. is only used in test files but is not part of test files"
func ExecuteHandler(name string, data interface{}) (interface{}, error) {
	handler, err := GetHandler(name)
	if err != nil {
		return nil, err
	}
	return handler(data)
}

// Plugin represents a plugin with a set of handlers
// want "identifier .Plugin. is only used in test files but is not part of test files"
type Plugin struct {
	Name      string
	Handlers  map[string]Handler
	IsEnabled bool
}

// RegisterPlugin registers all handlers from a plugin
// want "identifier .RegisterPlugin. is only used in test files but is not part of test files"
func RegisterPlugin(plugin Plugin) {
	if !plugin.IsEnabled {
		return
	}

	for name, handler := range plugin.Handlers {
		RegisterHandler(plugin.Name+"."+name, handler)
	}
}

func init() {
	// Register basic handlers during package initialization
	RegisterHandler("string", StringHandler)
	RegisterHandler("int", IntHandler)
}
