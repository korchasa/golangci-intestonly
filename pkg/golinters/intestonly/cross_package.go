// Package intestonly provides cross-package reference analysis.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzeCrossPackageReferences analyzes references between packages
func analyzeCrossPackageReferences(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Integrate the information about cross-package references
	for declName, declInfo := range result.Declarations {
		if declInfo.ImportRef != "" {
			if _, exists := result.TestUsages[declName]; exists {
				// This is referenced in tests and imported from another package
				// No action needed, this will be reported correctly
				continue
			}

			if len(result.NonTestUsages[declName]) > 0 {
				// This is used in non-test code, so it's not test-only
				continue
			}

			// At this point, we have a declaration that imports something
			// but it's not directly used in tests or non-tests
			// This could be due to missing references, so we need to check
			// if the import itself is used only in tests
		}
	}
}

// analyzeRobustCrossPackageReferences provides an enhanced analysis of cross-package references
// with better handling of alias imports, qualified identifiers, and complex dependency chains
func analyzeRobustCrossPackageReferences(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Performing robust cross-package reference analysis...")
	}

	// First, analyze standard cross-package references as a starting point
	analyzeCrossPackageReferences(pass, result, config)

	// Now perform more sophisticated analysis
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()

		// Check if this is a test file
		isTest := isTestFile(fileName, config)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) && !isTest {
			continue
		}

		// Track all imports and their aliases for this file
		imports := make(map[string]string) // alias/name -> package path

		// First pass: collect import information
		for _, imp := range file.Imports {
			// Get package path (removing quotes)
			pkgPath := strings.Trim(imp.Path.Value, "\"")

			// Handle aliased imports
			if imp.Name != nil {
				// Explicit alias (e.g., import foo "bar/baz")
				imports[imp.Name.Name] = pkgPath
			} else {
				// Default import (use last part of path as name)
				parts := strings.Split(pkgPath, "/")
				pkgName := parts[len(parts)-1]
				imports[pkgName] = pkgPath
			}
		}

		// Second pass: analyze selector expressions for imported package references
		ast.Inspect(file, func(n ast.Node) bool {
			selectorExpr, ok := n.(*ast.SelectorExpr)
			if !ok || selectorExpr.Sel == nil {
				return true
			}

			// Check if the X part is an identifier
			x, ok := selectorExpr.X.(*ast.Ident)
			if !ok || x.Obj != nil { // If x.Obj != nil, this is a field/method access, not a package access
				return true
			}

			// Check if this is a reference to an imported package
			if pkgPath, isImport := imports[x.Name]; isImport {
				// This is a qualified identifier referencing an imported package
				qualifiedName := pkgPath + "." + selectorExpr.Sel.Name

				// Record this reference based on whether it's in a test file
				if isTest {
					// Attempt to find declarations that match this reference
					for declName, declInfo := range result.Declarations {
						if declInfo.ImportRef == qualifiedName {
							result.TestUsages[declName] = append(result.TestUsages[declName], selectorExpr.Sel.Pos())
							if config.Debug {
								fmt.Printf("Cross-package test usage: %s in %s\n", qualifiedName, fileName)
							}
						}
					}
				} else {
					// Attempt to find declarations that match this reference
					for declName, declInfo := range result.Declarations {
						if declInfo.ImportRef == qualifiedName {
							result.NonTestUsages[declName] = append(result.NonTestUsages[declName], selectorExpr.Sel.Pos())
							if config.Debug {
								fmt.Printf("Cross-package non-test usage: %s in %s\n", qualifiedName, fileName)
							}
						}
					}
				}
			}

			return true
		})
	}

	// Additional processing: propagate usage information from interface implementations
	if config.EnableInterfaceImplementationDetection {
		for interfaceName, implementations := range result.Implementations {
			// If interface is used in production, mark all implementations as used
			if _, usedInProd := result.NonTestUsages[interfaceName]; usedInProd {
				for _, implType := range implementations {
					if _, exists := result.NonTestUsages[implType]; !exists {
						result.NonTestUsages[implType] = []token.Pos{token.NoPos}
						if config.Debug {
							fmt.Printf("Cross-package marking: %s used in production via interface %s\n",
								implType, interfaceName)
						}
					}
				}
			}
		}
	}
}

// processImportedPackage analyzes references from an imported package
//
//nolint:unused // Will be used in future implementation
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
