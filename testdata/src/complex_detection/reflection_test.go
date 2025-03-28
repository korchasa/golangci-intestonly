package complex_detection

import (
	"reflect"
	"testing"
)

func TestComplexReflection(t *testing.T) {
	// Создаем структуру для тестирования
	obj := &ComplexReflectionStruct{
		Field1: "test",
		Field2: 42,
		inner: innerStruct{
			Value: "inner value",
		},
	}

	// Получаем значение через reflection
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()

	// Проверяем тип и поля
	t.Logf("Type name: %s", objType.Elem().Name())
	t.Logf("Field count: %d", objType.Elem().NumField())

	// Получаем и вызываем метод по имени
	method, found := objType.MethodByName("DynamicMethod")
	if !found {
		t.Errorf("Method DynamicMethod not found")
	}

	// Вызываем метод через reflection
	args := []reflect.Value{objValue}
	args = append(args, reflect.ValueOf("prefix-"))
	result := method.Func.Call(args)
	t.Logf("Method result: %s", result[0].String())

	// Получаем поле через reflection
	field := reflect.Indirect(objValue).FieldByName("Field1")
	t.Logf("Field1 value: %s", field.String())

	// Используем обертку для более скрытого использования reflection
	wrapper := &ReflectionWrapper{
		target: obj,
	}
	_ = wrapper.CallMethod("DynamicMethod", "arg")

	// Используем общий обработчик reflection
	jsonStr := GenericReflectionHandler(obj)
	t.Logf("JSON representation: %s", jsonStr)

	// Динамически проверяем, реализует ли структура какие-либо интерфейсы
	checkInterface(obj, t)
}

// checkInterface использует reflection для проверки интерфейсов
func checkInterface(obj interface{}, t *testing.T) {
	objType := reflect.TypeOf(obj)
	for i := 0; i < objType.NumMethod(); i++ {
		method := objType.Method(i)
		t.Logf("Found method: %s", method.Name)
	}

	// Проверяем внутреннее поле через reflection
	if crs, ok := obj.(*ComplexReflectionStruct); ok {
		t.Logf("Inner value: %s", crs.GetInnerValue())
	}
}
