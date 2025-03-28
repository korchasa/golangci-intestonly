package false_positives

// SharedFunction is used in both test and non-test code
func SharedFunction() string {
	return "shared"
}

// HelperType is used in both contexts
type HelperType struct {
	ID   int
	Name string
}

// Method is used in both contexts
func (h HelperType) Method() string {
	return h.Name
}

// CONSTANT is used in both contexts
const CONSTANT = "shared constant"

// sharedVariable is used in both contexts
var sharedVariable = "shared variable"

// UserFacing uses multiple shared elements
func UserFacing() HelperType {
	h := HelperType{
		ID:   1,
		Name: SharedFunction(),
	}

	if h.Name == sharedVariable {
		h.Name = CONSTANT
	}

	return h
}
