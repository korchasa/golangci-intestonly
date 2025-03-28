package complex_detection

import (
	"testing"
)

// CustomHandler - custom handler for tests
func CustomHandler(data interface{}) (interface{}, error) {
	if str, ok := data.(string); ok {
		return "Custom: " + str, nil
	}
	return nil, nil
}

func TestBasicRegistry(t *testing.T) {
	// Check the operation with basic handlers registered in init()
	result, err := ExecuteHandler("string", "test data")
	if err != nil {
		t.Errorf("Error executing string handler: %v", err)
	}
	t.Logf("String handler result: %v", result)

	result, err = ExecuteHandler("int", 42)
	if err != nil {
		t.Errorf("Error executing int handler: %v", err)
	}
	t.Logf("Int handler result: %v", result)
}

func TestCustomHandlerRegistration(t *testing.T) {
	// Register a new handler
	RegisterHandler("custom", CustomHandler)

	// Check that the handler is successfully registered
	handler, err := GetHandler("custom")
	if err != nil {
		t.Errorf("Error getting custom handler: %v", err)
	}

	// Check the operation of the registered handler
	result, err := handler("custom test")
	if err != nil {
		t.Errorf("Error running custom handler: %v", err)
	}
	t.Logf("Custom handler result: %v", result)

	// Check through the common execution function
	result, err = ExecuteHandler("custom", "via execute")
	if err != nil {
		t.Errorf("Error executing custom handler: %v", err)
	}
	t.Logf("Execute custom handler result: %v", result)
}

func TestPluginRegistration(t *testing.T) {
	// Create and register a plugin
	plugin := Plugin{
		Name: "testPlugin",
		Handlers: map[string]Handler{
			"double": func(data interface{}) (interface{}, error) {
				if str, ok := data.(string); ok {
					return str + str, nil
				}
				return nil, nil
			},
			"count": func(data interface{}) (interface{}, error) {
				if str, ok := data.(string); ok {
					return len(str), nil
				}
				return nil, nil
			},
		},
		IsEnabled: true,
	}

	// Register the plugin
	RegisterPlugin(plugin)

	// Check handlers from the plugin
	result, err := ExecuteHandler("testPlugin.double", "abc")
	if err != nil {
		t.Errorf("Error executing plugin double handler: %v", err)
	}
	t.Logf("Plugin double handler result: %v", result)

	result, err = ExecuteHandler("testPlugin.count", "abc")
	if err != nil {
		t.Errorf("Error executing plugin count handler: %v", err)
	}
	t.Logf("Plugin count handler result: %v", result)

	// Check that the handler can be retrieved via GetHandler
	handler, err := GetHandler("testPlugin.double")
	if err != nil {
		t.Errorf("Error getting plugin double handler: %v", err)
	}
	result, _ = handler("xyz")
	t.Logf("Direct plugin handler result: %v", result)
}

func TestDisabledPlugin(t *testing.T) {
	// Create a disabled plugin
	disabledPlugin := Plugin{
		Name: "disabled",
		Handlers: map[string]Handler{
			"test": func(data interface{}) (interface{}, error) {
				return "disabled", nil
			},
		},
		IsEnabled: false,
	}

	// Register the disabled plugin
	RegisterPlugin(disabledPlugin)

	// Check that handlers from the disabled plugin are not registered
	_, err := GetHandler("disabled.test")
	if err == nil {
		t.Errorf("Disabled plugin handler should not be registered")
	}
}
