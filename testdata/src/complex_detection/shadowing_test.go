package complex_detection

import (
	"testing"
)

func TestGlobalAccess(t *testing.T) {
	// Проверяем прямой доступ к глобальным идентификаторам
	t.Logf("Global variable: %s", GlobalVariable)
	t.Logf("Global function: %s", GlobalFunction())

	g := &GlobalType{Field: "test field"}
	t.Logf("Global type method: %s", g.GlobalMethod())
}

func TestShadowingTypes(t *testing.T) {
	// Проверяем затенение типов в контейнерах
	container := &ShadowingContainer{
		GlobalVariable: "shadowed variable",
		GlobalType:     42,
	}

	// Доступ к затененным полям
	t.Logf("Container fields: %s, %d", container.GlobalVariable, container.GlobalType)

	// Параллельное использование глобальных идентификаторов и затененных полей
	t.Logf("Global and shadowed: %s vs %s", GlobalVariable, container.GlobalVariable)
}

func TestShadowingFunctions(t *testing.T) {
	// Проверяем затенение в функциях
	result := ShadowingFunction("parameter")
	t.Logf("Shadowing function result: %s", result)
}

func TestNestedShadowing(t *testing.T) {
	// Проверяем вложенное затенение
	result := NestedShadowing()
	t.Logf("Nested shadowing result: %s", result)
}

func TestDirectAccessWithShadowing(t *testing.T) {
	// Локальное затенение в тесте
	GlobalVariable := "test local"
	GlobalFunction := func() string { return "test function" }

	// Параллельно используем и локальные, и глобальные версии
	t.Logf("Local shadowed: %s, %s", GlobalVariable, GlobalFunction())

	// Для доступа к глобальным версиям нужно использовать квалифицированные имена
	// В тестах это сложнее, так как мы в том же пакете, поэтому используем
	// специальную функцию, которая обращается к глобальным версиям
	global := NotShadowed()
	t.Logf("Global via function: %s", global)
}

func TestComplexShadowing(t *testing.T) {
	// Создаем локальную переменную с тем же именем, что и глобальный тип
	GlobalType := "string value"

	// Используем локальное и глобальное значение
	t.Logf("Local string: %s", GlobalType)

	// Для доступа к глобальному типу используем явное приведение
	g := &(struct {
		Field string
	}{"field value"})

	t.Logf("Struct field: %s", g.Field)

	// Сложное затенение с передачей затененных переменных в другие функции
	localVar := GlobalVariable // Не затенение, а копирование глобальной переменной
	func() {
		// Затенение внутри анонимной функции
		GlobalVariable := "anonymous"
		t.Logf("In anonymous function: %s, outer: %s", GlobalVariable, localVar)
	}()
}
