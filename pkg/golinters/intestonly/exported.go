// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
// This file implements handling of exported identifiers.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// processExportedIdentifiers handles special cases for exported identifiers
func processExportedIdentifiers(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Processing exported identifiers...")
	}

	// Track exported declarations
	for name, info := range result.Declarations {
		if isExported(name) {
			result.ExportedDecls[name] = true

			// If configured to consider exported constants used
			if config.ConsiderExportedConstantsUsed && info.DeclType == DeclConstant {
				usage := UsageInfo{
					Pos:      token.NoPos,
					FilePath: "",
					IsTest:   false,
				}
				result.Usages[name] = append(result.Usages[name], usage)
				if config.Debug {
					fmt.Printf("Marking exported constant %s as used\n", name)
				}
			}
		}
	}

	// Process each file for exported identifier usage
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName, config)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) {
			continue
		}

		// Analyze exported identifier usage
		analyzeExportedIdentifierUsage(file, fileName, isTest, result, config)
	}
}

// analyzeExportedIdentifierUsage analyzes how exported identifiers are used in a file
func analyzeExportedIdentifierUsage(file *ast.File, fileName string, isTest bool, result *AnalysisResult, config *Config) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			// Skip if this is a declaration
			if _, isDeclPos := result.DeclPositions[node.Pos()]; isDeclPos {
				return true
			}

			// Check if this identifier is exported and tracked
			if isExported(node.Name) {
				if _, isDeclared := result.Declarations[node.Name]; isDeclared {
					usage := UsageInfo{
						Pos:      node.Pos(),
						FilePath: fileName,
						IsTest:   isTest,
					}
					if isTest {
						result.TestUsages[node.Name] = append(result.TestUsages[node.Name], usage)
					} else {
						result.Usages[node.Name] = append(result.Usages[node.Name], usage)
					}
				}
			}
		}
		return true
	})
}
