package complex_detection

// BaseStruct - базовая структура для встраивания
// want "identifier .BaseStruct. is only used in test files but is not part of test files"
type BaseStruct struct {
	BaseField string
}

// BaseMethod - базовый метод, который будет доступен через встраивание
// want "identifier .BaseMethod. is only used in test files but is not part of test files"
func (b *BaseStruct) BaseMethod() string {
	return b.BaseField
}

// MiddleStruct - промежуточная структура, встраивающая базовую
// want "identifier .MiddleStruct. is only used in test files but is not part of test files"
type MiddleStruct struct {
	BaseStruct  // встроенная BaseStruct
	MiddleField int
}

// MiddleMethod - метод промежуточной структуры
// want "identifier .MiddleMethod. is only used in test files but is not part of test files"
func (m *MiddleStruct) MiddleMethod() int {
	return m.MiddleField
}

// TopStruct - верхнеуровневая структура, встраивающая промежуточную
// want "identifier .TopStruct. is only used in test files but is not part of test files"
type TopStruct struct {
	MiddleStruct // встроенная MiddleStruct
	TopField     bool
}

// TopMethod - метод верхнеуровневой структуры
// want "identifier .TopMethod. is only used in test files but is not part of test files"
func (t *TopStruct) TopMethod() bool {
	return t.TopField
}

// MixinOne - первый примесь (mixin)
// want "identifier .MixinOne. is only used in test files but is not part of test files"
type MixinOne struct {
	MixinOneField string
}

// MixinOneMethod - метод первого примеся
// want "identifier .MixinOneMethod. is only used in test files but is not part of test files"
func (m *MixinOne) MixinOneMethod() string {
	return m.MixinOneField
}

// MixinTwo - второй примесь
// want "identifier .MixinTwo. is only used in test files but is not part of test files"
type MixinTwo struct {
	MixinTwoField int
}

// MixinTwoMethod - метод второго примеся
// want "identifier .MixinTwoMethod. is only used in test files but is not part of test files"
func (m *MixinTwo) MixinTwoMethod() int {
	return m.MixinTwoField
}

// ComplexEmbedding - структура с множественным встраиванием
// want "identifier .ComplexEmbedding. is only used in test files but is not part of test files"
type ComplexEmbedding struct {
	BaseStruct // встраиваем базовую структуру напрямую
	MixinOne   // первый примесь
	MixinTwo   // второй примесь
	OwnField   float64
}

// OwnMethod - собственный метод структуры с множественным встраиванием
// want "identifier .OwnMethod. is only used in test files but is not part of test files"
func (c *ComplexEmbedding) OwnMethod() float64 {
	return c.OwnField
}
