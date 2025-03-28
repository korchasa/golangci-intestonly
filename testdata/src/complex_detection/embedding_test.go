package complex_detection

import (
	"testing"
)

func TestSimpleEmbedding(t *testing.T) {
	// Проверяем простое встраивание - MiddleStruct содержит BaseStruct
	middle := &MiddleStruct{
		BaseStruct: BaseStruct{
			BaseField: "base value",
		},
		MiddleField: 42,
	}

	// Вызов метода встроенной структуры
	baseResult := middle.BaseMethod()
	t.Logf("Base method result: %s", baseResult)

	// Вызов собственного метода
	middleResult := middle.MiddleMethod()
	t.Logf("Middle method result: %d", middleResult)

	// Прямой доступ к полям встроенной структуры
	t.Logf("Base field via embedding: %s", middle.BaseField)
}

func TestMultiLevelEmbedding(t *testing.T) {
	// Проверяем многоуровневое встраивание - TopStruct > MiddleStruct > BaseStruct
	top := &TopStruct{
		MiddleStruct: MiddleStruct{
			BaseStruct: BaseStruct{
				BaseField: "base in top",
			},
			MiddleField: 100,
		},
		TopField: true,
	}

	// Вызов метода из верхнего уровня
	topResult := top.TopMethod()
	t.Logf("Top method result: %v", topResult)

	// Вызов метода из среднего уровня
	middleResult := top.MiddleMethod()
	t.Logf("Middle method via top: %d", middleResult)

	// Вызов метода из базового уровня
	baseResult := top.BaseMethod()
	t.Logf("Base method via top: %s", baseResult)

	// Доступ к полям на разных уровнях встраивания
	t.Logf("Multi-level field access: %s, %d, %v",
		top.BaseField, top.MiddleField, top.TopField)
}

func TestComplexMultipleEmbedding(t *testing.T) {
	// Проверяем множественное встраивание различных типов
	complex := &ComplexEmbedding{
		BaseStruct: BaseStruct{
			BaseField: "base in complex",
		},
		MixinOne: MixinOne{
			MixinOneField: "mixin one value",
		},
		MixinTwo: MixinTwo{
			MixinTwoField: 200,
		},
		OwnField: 3.14,
	}

	// Вызов собственного метода
	ownResult := complex.OwnMethod()
	t.Logf("Own method result: %f", ownResult)

	// Вызов методов из всех встроенных типов
	baseResult := complex.BaseMethod()
	mixinOneResult := complex.MixinOneMethod()
	mixinTwoResult := complex.MixinTwoMethod()

	t.Logf("Results from embedded types: %s, %s, %d",
		baseResult, mixinOneResult, mixinTwoResult)

	// Доступ к полям из всех встроенных типов
	t.Logf("Fields from embedded types: %s, %s, %d, %f",
		complex.BaseField, complex.MixinOneField,
		complex.MixinTwoField, complex.OwnField)
}
