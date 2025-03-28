// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
// This file implements handling of exported identifiers.
package intestonly

import (
	"fmt"
	"go/token"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

// processExportedIdentifiers handles special cases for exported identifiers
func processExportedIdentifiers(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Processing exported identifiers...")
	}

	// Process declarations to identify exported identifiers
	identifyExportedDeclarations(result, config)

	// Apply heuristics for exported identifiers
	applyExportedIdentifierHeuristics(result, config)
}

// identifyExportedDeclarations marks which declarations are exported (capitalized)
func identifyExportedDeclarations(result *AnalysisResult, config *Config) {
	for declName := range result.Declarations {
		// Check if the first character is uppercase (exported)
		if len(declName) > 0 && unicode.IsUpper(rune(declName[0])) {
			result.ExportedDecls[declName] = true

			if config.Debug {
				fmt.Printf("Identified exported identifier: %s\n", declName)
			}
		}
	}
}

// applyExportedIdentifierHeuristics applies special handling for exported declarations
func applyExportedIdentifierHeuristics(result *AnalysisResult, config *Config) {
	// For each exported declaration
	for declName := range result.ExportedDecls {
		declInfo, exists := result.Declarations[declName]
		if !exists {
			continue
		}

		// Case 1: Exported functions in main packages are likely used as entrypoints
		if isMainPackage(declInfo.PkgPath) && !declInfo.IsMethod {
			// Mark as used in production to prevent false positives
			if _, exists := result.NonTestUsages[declName]; !exists {
				result.NonTestUsages[declName] = []token.Pos{token.NoPos}
				if config.Debug {
					fmt.Printf("Marking exported function %s in main package as used\n", declName)
				}
			}
			continue
		}

		// Case 2: Public API functions with documented examples in test files
		if hasDocumentedExamples(declName, result) {
			// Mark as used in production to prevent false positives
			if _, exists := result.NonTestUsages[declName]; !exists {
				result.NonTestUsages[declName] = []token.Pos{token.NoPos}
				if config.Debug {
					fmt.Printf("Marking exported function %s with examples as used\n", declName)
				}
			}
			continue
		}

		// Case 3: Types with exported fields or methods
		if typeName := extractTypeName(declName); typeName != "" && typeName == declName {
			if hasExportedMethods(typeName, result) {
				// Mark as used in production to prevent false positives
				if _, exists := result.NonTestUsages[declName]; !exists {
					result.NonTestUsages[declName] = []token.Pos{token.NoPos}
					if config.Debug {
						fmt.Printf("Marking type %s with exported methods as used\n", declName)
					}
				}
				continue
			}
		}

		// Case 4: Exported methods of exported types
		if typeName := extractTypeName(declName); typeName != "" && typeName != declName {
			if result.ExportedDecls[typeName] {
				// Mark as used in production to prevent false positives
				if _, exists := result.NonTestUsages[declName]; !exists {
					result.NonTestUsages[declName] = []token.Pos{token.NoPos}
					if config.Debug {
						fmt.Printf("Marking exported method %s of exported type as used\n", declName)
					}
				}
				continue
			}
		}

		// Case 5: Exported constants and variables used as public API
		if !declInfo.IsMethod && (isConstantName(declName) || isVariableName(declName)) {
			// Consider exported constants and variables potentially used by API consumers
			if config.ConsiderExportedConstantsUsed {
				if _, exists := result.NonTestUsages[declName]; !exists {
					result.NonTestUsages[declName] = []token.Pos{token.NoPos}
					if config.Debug {
						fmt.Printf("Marking exported constant/variable %s as used\n", declName)
					}
				}
			}
		}
	}
}

// isMainPackage checks if the package is a main package
func isMainPackage(pkgPath string) bool {
	return pkgPath == "main"
}

// hasDocumentedExamples checks if a declaration has documented examples in test files
func hasDocumentedExamples(declName string, result *AnalysisResult) bool {
	exampleName := "Example" + declName

	// Check if there's a test usage of an example function for this declaration
	for testName := range result.TestUsages {
		if testName == exampleName {
			return true
		}
	}

	return false
}

// extractTypeName extracts the type name from a qualified method name (Type.Method)
func extractTypeName(name string) string {
	// Find the dot separator in method names
	for i, char := range name {
		if char == '.' {
			return name[:i]
		}
	}
	return name // If no dot, return the entire name
}

// hasExportedMethods checks if a type has any exported methods
func hasExportedMethods(typeName string, result *AnalysisResult) bool {
	methods, ok := result.MethodsOfType[typeName]
	if !ok {
		return false
	}

	for _, method := range methods {
		if len(method) > 0 && unicode.IsUpper(rune(method[0])) {
			return true
		}
	}

	return false
}

// isConstantName uses heuristics to determine if a name likely represents a constant
func isConstantName(name string) bool {
	// Typical Go constant naming patterns include ALL_CAPS or CamelCase
	allCaps := true
	for _, r := range name {
		if unicode.IsLower(r) && r != '_' {
			allCaps = false
			break
		}
	}

	return allCaps || (len(name) > 0 && unicode.IsUpper(rune(name[0])))
}

// isVariableName uses heuristics to determine if a name likely represents a variable
func isVariableName(name string) bool {
	return len(name) > 0 && unicode.IsUpper(rune(name[0])) && !isConstantName(name)
}
