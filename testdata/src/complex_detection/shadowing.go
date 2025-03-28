package complex_detection

// GlobalVariable - глобальная переменная, которая может быть затенена
// want "identifier .GlobalVariable. is only used in test files but is not part of test files"
var GlobalVariable = "global value"

// GlobalFunction - глобальная функция, которая может быть затенена
// want "identifier .GlobalFunction. is only used in test files but is not part of test files"
func GlobalFunction() string {
	return "global function"
}

// GlobalType - глобальный тип, который может быть затенен
// want "identifier .GlobalType. is only used in test files but is not part of test files"
type GlobalType struct {
	Field string
}

// GlobalMethod - метод глобального типа
// want "identifier .GlobalMethod. is only used in test files but is not part of test files"
func (g *GlobalType) GlobalMethod() string {
	return g.Field
}

// ShadowingContainer содержит поля с теми же именами, что и глобальные идентификаторы
// want "identifier .ShadowingContainer. is only used in test files but is not part of test files"
type ShadowingContainer struct {
	// Затеняет глобальную переменную
	GlobalVariable string

	// Затеняет глобальный тип
	GlobalType int
}

// ShadowingFunction затеняет глобальные идентификаторы локальными переменными
// want "identifier .ShadowingFunction. is only used in test files but is not part of test files"
func ShadowingFunction(GlobalVariable string) string {
	// Локальная переменная затеняет параметр функции
	GlobalVariable = "local in function"

	// Затеняем глобальную функцию локальной переменной
	GlobalFunction := func() string {
		return "local function"
	}

	// Затеняем глобальный тип локальной переменной
	GlobalType := struct {
		Value int
	}{
		Value: 100,
	}

	return GlobalVariable + " " + GlobalFunction() + " " + string(rune(GlobalType.Value))
}

// NestedShadowing демонстрирует затенение в разных областях видимости
// want "identifier .NestedShadowing. is only used in test files but is not part of test files"
func NestedShadowing() string {
	// Затеняем глобальную переменную на верхнем уровне
	GlobalVariable := "outer"

	// Создаем анонимную функцию с затенением
	f := func() string {
		// Затеняем переменную из внешней области видимости
		GlobalVariable := "inner"
		return GlobalVariable
	}

	// Еще один уровень вложенности
	{
		// Затеняем в блоке
		GlobalVariable := "block"
		_ = GlobalVariable // Используем переменную, чтобы не было предупреждения
	}

	return GlobalVariable + " " + f()
}

// NotShadowed использует глобальные идентификаторы без затенения
// want "identifier .NotShadowed. is only used in test files but is not part of test files"
func NotShadowed() string {
	// Используем глобальные идентификаторы напрямую
	g := &GlobalType{Field: "direct access"}
	return GlobalVariable + " " + GlobalFunction() + " " + g.GlobalMethod()
}
