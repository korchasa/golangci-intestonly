package cross_package_ref

import "testing"

func TestExportedFuncOnlyTest(t *testing.T) {
	result := ExportedFuncOnlyTest()
	if result != "I'm exported but only used in test files" {
		t.Errorf("Unexpected result: %s", result)
	}
}
