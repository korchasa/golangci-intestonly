// Package cross_package_ref demonstrates cross-package references
package cross_package_ref

// ExportedFunc is a function that is exported and is used in another package
func ExportedFunc() string {
	return "I'm exported and should be detected as used in another package"
}

// ExportedFuncOnlyTest is a function that is exported but used only in tests
func ExportedFuncOnlyTest() string {
	return "I'm exported but only used in test files"
}

// internalFunc is not exported and only used internally
func internalFunc() string {
	return "I'm internal and only used in this package"
}

func UseInternalFunc() {
	_ = internalFunc()
}
