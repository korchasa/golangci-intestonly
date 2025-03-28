package complex_detection

import (
	"testing"
)

func TestSimpleEmbedding(t *testing.T) {
	// Check simple embedding - MiddleStruct contains BaseStruct
	middle := &MiddleStruct{
		BaseStruct: BaseStruct{
			BaseField: "base value",
		},
		MiddleField: 42,
	}

	// Call method of embedded structure
	baseResult := middle.BaseMethod()
	t.Logf("Base method result: %s", baseResult)

	// Call own method
	middleResult := middle.MiddleMethod()
	t.Logf("Middle method result: %d", middleResult)

	// Direct access to fields of embedded structure
	t.Logf("Base field via embedding: %s", middle.BaseField)
}

func TestMultiLevelEmbedding(t *testing.T) {
	// Check multi-level embedding - TopStruct > MiddleStruct > BaseStruct
	top := &TopStruct{
		MiddleStruct: MiddleStruct{
			BaseStruct: BaseStruct{
				BaseField: "base in top",
			},
			MiddleField: 100,
		},
		TopField: true,
	}

	// Call method from top level
	topResult := top.TopMethod()
	t.Logf("Top method result: %v", topResult)

	// Call method from middle level
	middleResult := top.MiddleMethod()
	t.Logf("Middle method via top: %d", middleResult)

	// Call method from base level
	baseResult := top.BaseMethod()
	t.Logf("Base method via top: %s", baseResult)

	// Access fields at different levels of embedding
	t.Logf("Multi-level field access: %s, %d, %v",
		top.BaseField, top.MiddleField, top.TopField)
}

func TestComplexMultipleEmbedding(t *testing.T) {
	// Check multiple embedding of different types
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

	// Call own method
	ownResult := complex.OwnMethod()
	t.Logf("Own method result: %f", ownResult)

	// Call methods from all embedded types
	baseResult := complex.BaseMethod()
	mixinOneResult := complex.MixinOneMethod()
	mixinTwoResult := complex.MixinTwoMethod()

	t.Logf("Results from embedded types: %s, %s, %d",
		baseResult, mixinOneResult, mixinTwoResult)

	// Access fields from all embedded types
	t.Logf("Fields from embedded types: %s, %s, %d, %f",
		complex.BaseField, complex.MixinOneField,
		complex.MixinTwoField, complex.OwnField)
}
