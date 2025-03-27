// Package cross_package_user uses code from cross_package_ref
package cross_package_user

import (
	"cross_package_ref"
)

// UseExportedFunc uses a function from another package
func UseExportedFunc() string {
	return cross_package_ref.ExportedFunc()
}
