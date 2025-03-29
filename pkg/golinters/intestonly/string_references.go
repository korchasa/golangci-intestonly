// Package intestonly provides string reference analysis.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzeStringReferences scans for function and type references within string literals
func analyzeStringReferences(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Always enable string literal analysis for now since we don't have the config field yet
	// TODO: Add EnableStringLiteralAnalysis to Config struct if needed in the future

	// Process each file in the package
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName, config)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) {
			if config.Debug {
				fmt.Printf("Skipping string reference analysis for file: %s\n", fileName)
			}
			continue
		}

		// Walk the AST to find string literals
		ast.Inspect(file, func(n ast.Node) bool {
			// Check for string literals
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			strValue := strings.Trim(lit.Value, "\"'`")

			// Check for function references in this string literal
			findFunctionReferencesInString(strValue, result, fileName, isTest, config)

			return true
		})
	}
}

// findFunctionReferencesInString checks for potential function references within a string
func findFunctionReferencesInString(str string, result *AnalysisResult, fileName string, isTest bool, config *Config) {
	// Get all declarations to check for in the string
	for declName := range result.Declarations {
		// Skip very short names to avoid false positives
		if len(declName) < 3 {
			continue
		}

		// Check if the declaration name appears in the string with common function call patterns
		patterns := []string{
			declName + "()",
			declName + " ()",
			declName + "(",
			declName + " (",
		}

		for _, pattern := range patterns {
			if strings.Contains(str, pattern) {
				// This declaration is referenced in a string literal
				usage := UsageInfo{
					Pos:      token.NoPos, // We don't have exact position within the string
					FilePath: fileName,
					IsTest:   isTest,
				}

				if isTest {
					// Note: only recording detection in a test context is not enough
					// We're being conservative here and not marking test-only
					// when there's a string reference. Just recording for debugging.
					result.TestUsages[declName] = append(result.TestUsages[declName], usage)
				} else {
					// Record as a non-test usage since it's referenced in a string in production code
					result.Usages[declName] = append(result.Usages[declName], usage)
				}

				if config.Debug {
					fmt.Printf("String reference to %s found in %s (isTest: %v)\n",
						declName, fileName, isTest)
				}

				// Found one pattern match, no need to check the others
				break
			}
		}
	}

	// Also check for cross-package references in string by searching for import references
	for importRef := range result.ImportRefs {
		shortName := result.ImportRefs[importRef]
		if len(shortName) < 3 {
			continue
		}

		// Check for the pattern with common call syntax
		patterns := []string{
			importRef + "()",
			importRef + " ()",
			importRef + "(",
			importRef + " (",
		}

		for _, pattern := range patterns {
			if strings.Contains(str, pattern) {
				// Find the corresponding declaration
				for declName, declInfo := range result.Declarations {
					if declInfo.ImportRef == importRef {
						usage := UsageInfo{
							Pos:      token.NoPos,
							FilePath: fileName,
							IsTest:   isTest,
						}

						if isTest {
							result.TestUsages[declName] = append(result.TestUsages[declName], usage)
						} else {
							result.Usages[declName] = append(result.Usages[declName], usage)
						}

						if config.Debug {
							fmt.Printf("String reference to imported %s (%s) found in %s (isTest: %v)\n",
								declName, importRef, fileName, isTest)
						}
					}
				}

				// Found one pattern match, no need to check the others
				break
			}
		}
	}
}
