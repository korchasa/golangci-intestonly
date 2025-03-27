package intestonly

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// analyzeCrossPackageReferences examines imports to detect cross-package usages
func analyzeCrossPackageReferences(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// For each imported package
	for _, importedPkg := range pass.Pkg.Imports() {
		if importedPkg.Scope() == nil {
			continue
		}

		// Process the imported package
		processImportedPackage(importedPkg, pass, result, config)
	}
}

// processImportedPackage analyzes references from an imported package
func processImportedPackage(importedPkg *types.Package, pass *analysis.Pass, result *AnalysisResult, config *Config) {
	importPath := importedPkg.Path()

	// Check each name in the imported package's scope
	for _, name := range importedPkg.Scope().Names() {
		obj := importedPkg.Scope().Lookup(name)
		if obj == nil {
			continue
		}

		importRef := importPath + "." + obj.Name()

		// Check if this object references one of our declarations
		for declName, info := range result.Declarations {
			if info.ImportRef == importRef {
				// This package is imported, so its objects may be used
				// Mark as non-test usage
				result.NonTestUsages[declName] = append(result.NonTestUsages[declName], token.NoPos)
				if config.Debug {
					fmt.Printf("Cross-package reference to %s from %s\n", declName, importPath)
				}
			}
		}
	}
}
