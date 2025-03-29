// Package intestonly provides cross-package reference analysis.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// crossPackageRefs is a map to track cross-package references
// It's defined as a field in AnalysisResult to avoid global state

// analyzeCrossPackageReferences analyzes references to declarations from other packages
func analyzeCrossPackageReferences(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Use the fields in AnalysisResult to track imported packages and their declarations
	// These are already initialized in NewAnalysisResult

	// Process each file in the package
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName, config)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) {
			if config.Debug {
				fmt.Printf("Skipping cross-package analysis for file: %s\n", fileName)
			}
			continue
		}

		// Process imports and collect references
		imports := make(map[string]string) // pkgName -> importPath
		for _, imp := range file.Imports {
			if imp.Path != nil {
				importPath := strings.Trim(imp.Path.Value, "\"")
				var pkgName string
				if imp.Name != nil {
					pkgName = imp.Name.Name
				} else {
					// Extract package name from import path
					parts := strings.Split(importPath, "/")
					pkgName = parts[len(parts)-1]
				}
				imports[pkgName] = importPath

				// Track this imported package
				if isTest {
					if _, exists := result.TestPackageImports[importPath]; !exists {
						result.TestPackageImports[importPath] = make(map[string]bool)
					}
				} else {
					if _, exists := result.PackageImports[importPath]; !exists {
						result.PackageImports[importPath] = make(map[string]bool)
					}
				}
			}
		}

		// Analyze selector expressions for cross-package references
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
				importRef := pkgPath + "." + selectorExpr.Sel.Name

				// Record that this imported identifier is used
				if isTest {
					if result.TestPackageImports[pkgPath] != nil {
						result.TestPackageImports[pkgPath][selectorExpr.Sel.Name] = true
					}

					// Add to the tracking map for test references
					if _, exists := result.CrossPackageRefsList[pkgPath]; !exists {
						result.CrossPackageRefsList[pkgPath] = []string{}
					}

					// Add this identifier to the list of cross-package references if not already there
					found := false
					for _, ident := range result.CrossPackageRefsList[pkgPath] {
						if ident == selectorExpr.Sel.Name {
							found = true
							break
						}
					}
					if !found {
						result.CrossPackageRefsList[pkgPath] = append(result.CrossPackageRefsList[pkgPath], selectorExpr.Sel.Name)
					}
				} else {
					if result.PackageImports[pkgPath] != nil {
						result.PackageImports[pkgPath][selectorExpr.Sel.Name] = true
					}
				}

				// Create a usage info for this reference
				usage := UsageInfo{
					Pos:      selectorExpr.Sel.Pos(),
					FilePath: fileName,
					IsTest:   isTest,
				}

				// First check if this reference is to a known declaration
				found := false
				for declName, declInfo := range result.Declarations {
					if declInfo.ImportRef == importRef {
						// This is a reference to a known declaration
						if isTest {
							result.TestUsages[declName] = append(result.TestUsages[declName], usage)
						} else {
							result.Usages[declName] = append(result.Usages[declName], usage)
						}
						found = true

						if config.Debug {
							fmt.Printf("Cross-package %s usage: %s in %s (declName: %s)\n",
								map[bool]string{true: "test", false: "non-test"}[isTest], importRef, fileName, declName)
						}
					}
				}

				// Always record external references for matching later
				// This is important because packages might be analyzed out of order
				externalImportRef := importRef + "#external"
				if isTest {
					// Record test usage for this external reference
					externalUsage := UsageInfo{
						Pos:      selectorExpr.Sel.Pos(),
						FilePath: fileName,
						IsTest:   true,
					}
					result.TestUsages[externalImportRef] = append(result.TestUsages[externalImportRef], externalUsage)

					// Also add to a special map to track cross-package test references
					// This will help us correctly identify test-only usages in other packages
					result.CrossPackageTestRefs[importRef] = true
				} else {
					// Record non-test usage for this external reference
					externalUsage := UsageInfo{
						Pos:      selectorExpr.Sel.Pos(),
						FilePath: fileName,
						IsTest:   false,
					}
					result.Usages[externalImportRef] = append(result.Usages[externalImportRef], externalUsage)

					// Also record in a special map for tracking cross-package references in production code
					result.CrossPackageRefs[importRef] = true
				}

				if config.Debug && !found {
					fmt.Printf("External cross-package %s usage: %s in %s\n",
						map[bool]string{true: "test", false: "non-test"}[isTest], importRef, fileName)
				}
			}

			return true
		})
	}

	// Now match any declarations in the current package with external references
	for declName, declInfo := range result.Declarations {
		// Check if this declaration might be referred to externally
		if declInfo.ImportRef != "" && strings.HasPrefix(declInfo.ImportRef, result.CurrentPkgPath) {
			// This is an exported declaration in the current package
			externalImportRef := declInfo.ImportRef + "#external"

			// Check if there are any external test usages of this declaration
			if testUsages, hasTestUsages := result.TestUsages[externalImportRef]; hasTestUsages {
				// Add these test usages to the declaration
				result.TestUsages[declName] = append(result.TestUsages[declName], testUsages...)

				if config.Debug {
					fmt.Printf("Matched external test usages for %s (%s): %d usages\n",
						declName, declInfo.ImportRef, len(testUsages))
				}
			}

			// Check if there are any external non-test usages of this declaration
			if nonTestUsages, hasNonTestUsages := result.Usages[externalImportRef]; hasNonTestUsages {
				// Add these non-test usages to the declaration
				result.Usages[declName] = append(result.Usages[declName], nonTestUsages...)

				if config.Debug {
					fmt.Printf("Matched external non-test usages for %s (%s): %d usages\n",
						declName, declInfo.ImportRef, len(nonTestUsages))
				}
			}

			// Check cross-package references map in AnalysisResult
			if identifiers, exists := result.CrossPackageRefsList[result.CurrentPkgPath]; exists {
				for _, ident := range identifiers {
					if ident == declInfo.Name {
						// This package's declaration is referenced from a test in another package
						// Mark it as used in a test
						testUsage := UsageInfo{
							Pos:      token.NoPos,
							FilePath: "",
							IsTest:   true,
						}
						result.TestUsages[declName] = append(result.TestUsages[declName], testUsage)

						if config.Debug {
							fmt.Printf("Found cross-package test reference to %s\n", declName)
						}
					}
				}
			}
		}
	}

	// Use call graph to trace cross-package usages
	traceCrossPackageCallGraph(result, config)

	// Process external test references to ensure we mark all items as used in tests
	// that are referenced from test files in other packages
	processCrossPackageTestReferences(result, config)

	// Additional processing: propagate usage information from interface implementations
	if config.EnableInterfaceImplementationDetection {
		for interfaceName, implementations := range result.Implementations {
			// If interface is used in production, mark all implementations as used
			if usages, usedInProd := result.Usages[interfaceName]; usedInProd && len(usages) > 0 {
				for _, implType := range implementations {
					if _, exists := result.Usages[implType]; !exists {
						usage := UsageInfo{
							Pos:      token.NoPos,
							FilePath: "",
							IsTest:   false,
						}
						result.Usages[implType] = append(result.Usages[implType], usage)
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

// traceCrossPackageCallGraph uses the built call graph to find paths from test-only to production code
func traceCrossPackageCallGraph(result *AnalysisResult, config *Config) {
	// This function previously had logic that would mark a function as used in production
	// if it was called by a function used in production. This is incorrect for our use case.
	// A function that's only called by test code should be considered test-only,
	// even if it calls a function that's used in production.

	// For debugging only
	if config.Debug {
		// First, find all functions that are used in production code
		for name := range result.Declarations {
			if nonTestUsages := len(result.Usages[name]); nonTestUsages > 0 {
				fmt.Printf("Function used in production: %s\n", name)
			}
		}
	}
}

// processCrossPackageTestReferences processes external test references
func processCrossPackageTestReferences(result *AnalysisResult, config *Config) {
	// Go through all declarations
	for declName, declInfo := range result.Declarations {
		// If this declaration has an import reference, check if it's in the CrossPackageTestRefs map
		if declInfo.ImportRef != "" {
			if result.CrossPackageTestRefs[declInfo.ImportRef] {
				// This declaration is referenced in a test file in another package
				// Add a test usage
				testUsage := UsageInfo{
					Pos:      token.NoPos,
					FilePath: "",
					IsTest:   true,
				}

				result.TestUsages[declName] = append(result.TestUsages[declName], testUsage)

				if config.Debug {
					fmt.Printf("Added test usage for cross-package reference: %s (%s)\n",
						declName, declInfo.ImportRef)
				}
			}

			// Also check if it's used in non-test code
			if result.CrossPackageRefs[declInfo.ImportRef] {
				// This declaration is referenced in non-test code in another package
				// Add a non-test usage
				nonTestUsage := UsageInfo{
					Pos:      token.NoPos,
					FilePath: "",
					IsTest:   false,
				}

				result.Usages[declName] = append(result.Usages[declName], nonTestUsage)

				if config.Debug {
					fmt.Printf("Added non-test usage for cross-package reference: %s (%s)\n",
						declName, declInfo.ImportRef)
				}
			}
		}
	}
}
