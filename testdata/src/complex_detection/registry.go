package complex_detection

import (
	"fmt"
	"sync"
)

// Handler - тип функции-обработчика
// want "identifier .Handler. is only used in test files but is not part of test files"
type Handler func(data interface{}) (interface{}, error)

// Registry - глобальный реестр обработчиков
// want "identifier .Registry. is only used in test files but is not part of test files"
var Registry = struct {
	sync.RWMutex
	handlers map[string]Handler
}{
	handlers: make(map[string]Handler),
}

// RegisterHandler регистрирует обработчик в реестре
// want "identifier .RegisterHandler. is only used in test files but is not part of test files"
func RegisterHandler(name string, handler Handler) {
	Registry.Lock()
	defer Registry.Unlock()
	Registry.handlers[name] = handler
}

// GetHandler возвращает обработчик по имени
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

// StringHandler обрабатывает строковые данные
// want "identifier .StringHandler. is only used in test files but is not part of test files"
func StringHandler(data interface{}) (interface{}, error) {
	if str, ok := data.(string); ok {
		return "Processed: " + str, nil
	}
	return nil, fmt.Errorf("expected string, got %T", data)
}

// IntHandler обрабатывает целочисленные данные
// want "identifier .IntHandler. is only used in test files but is not part of test files"
func IntHandler(data interface{}) (interface{}, error) {
	if num, ok := data.(int); ok {
		return num * 2, nil
	}
	return nil, fmt.Errorf("expected int, got %T", data)
}

// ExecuteHandler выполняет обработчик по имени
// want "identifier .ExecuteHandler. is only used in test files but is not part of test files"
func ExecuteHandler(name string, data interface{}) (interface{}, error) {
	handler, err := GetHandler(name)
	if err != nil {
		return nil, err
	}
	return handler(data)
}

// Plugin представляет плагин с набором обработчиков
// want "identifier .Plugin. is only used in test files but is not part of test files"
type Plugin struct {
	Name      string
	Handlers  map[string]Handler
	IsEnabled bool
}

// RegisterPlugin регистрирует все обработчики из плагина
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
	// Регистрируем базовые обработчики при инициализации пакета
	RegisterHandler("string", StringHandler)
	RegisterHandler("int", IntHandler)
}
