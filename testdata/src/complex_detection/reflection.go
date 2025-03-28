// Package complex_detection tests complex detection scenarios.
package complex_detection

import (
	"encoding/json"
)

// ComplexReflectionStruct представляет структуру для тестирования сложных случаев reflection
// want "identifier .ComplexReflectionStruct. is only used in test files but is not part of test files"
type ComplexReflectionStruct struct {
	Field1 string
	Field2 int
	inner  innerStruct
}

// innerStruct - вложенная структура
// want "identifier .innerStruct. is only used in test files but is not part of test files"
type innerStruct struct {
	Value string
}

// DynamicMethod - метод, который может вызываться через reflection
// want "identifier .DynamicMethod. is only used in test files but is not part of test files"
func (c *ComplexReflectionStruct) DynamicMethod(arg string) string {
	return arg + c.Field1
}

// GetInnerValue получает значение из вложенной структуры
// want "identifier .GetInnerValue. is only used in test files but is not part of test files"
func (c *ComplexReflectionStruct) GetInnerValue() string {
	return c.inner.Value
}

// GenericReflectionHandler - обработчик, который использует reflection для работы с разными типами
// want "identifier .GenericReflectionHandler. is only used in test files but is not part of test files"
func GenericReflectionHandler(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data)
}

// ReflectionWrapper скрывает использование reflection за абстракцией
// want "identifier .ReflectionWrapper. is only used in test files but is not part of test files"
type ReflectionWrapper struct {
	target interface{}
}

// CallMethod вызывает метод через reflection
// want "identifier .CallMethod. is only used in test files but is not part of test files"
func (r *ReflectionWrapper) CallMethod(methodName string, args ...interface{}) interface{} {
	// В реальном коде здесь была бы реализация с использованием reflect.ValueOf().MethodByName()
	return nil
}
