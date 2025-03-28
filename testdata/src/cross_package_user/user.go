package cross_package_user

import (
	"fmt"

	"cross_package_ref"
)

// UseExportedFunc uses the exported function from cross_package_ref
func UseExportedFunc() string {
	return fmt.Sprintf("Using: %s", cross_package_ref.ExportedUsedFunc())
}

// UseType uses the exported type from cross_package_ref
func UseType() int {
	t := cross_package_ref.UsedType{ID: 42}
	return t.ID
}

// UseConst uses the exported constant from cross_package_ref
func UseConst() string {
	return fmt.Sprintf("Constant: %s", cross_package_ref.UsedConst)
}

// UseVar uses the exported variable from cross_package_ref
func UseVar() string {
	return fmt.Sprintf("Variable: %s", cross_package_ref.UsedVar)
}
