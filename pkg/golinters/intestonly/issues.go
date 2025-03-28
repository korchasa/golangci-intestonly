package intestonly

import (
	"fmt"
	"strings"

	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// generateIssues creates diagnostic issues for identifiers only used in tests
func generateIssues(pass *analysis.Pass, result *AnalysisResult, config *Config) []Issue {
	var issues []Issue

	// Debug information about declarations and usages
	if config.Debug {
		fmt.Printf("Declarations: %d\n", len(result.Declarations))
		fmt.Printf("Test usages: %d\n", len(result.TestUsages))
		fmt.Printf("Non-test usages: %d\n", len(result.NonTestUsages))
	}

	// Handle each declaration
	for name, decl := range result.Declarations {
		// Skip declarations with explicit exclude patterns
		if shouldExcludeFromReport(name, decl, config) {
			continue
		}

		// Check if the identifier is only used in test files
		usedInTest := len(result.TestUsages[name]) > 0
		usedInNonTest := len(result.NonTestUsages[name]) > 0

		// Handle test helpers if configured
		if config.ExcludeTestHelpers && isTestHelperIdentifier(name, config) {
			continue
		}

		// Handle unexported identifiers if configured
		if config.IgnoreUnexported && !ast.IsExported(name) {
			continue
		}

		// Skip regular methods if configured
		if !config.CheckMethods && decl.IsMethod {
			continue
		}

		// Report identifiers that are used only in tests
		if usedInTest && !usedInNonTest {
			if decl.Pos != token.NoPos {
				message := fmt.Sprintf("identifier %q is only used in test files but is not part of test files", name)
				issues = append(issues, Issue{
					Pos:     decl.Pos,
					Message: message,
				})
			}
		}
	}

	return issues
}

// shouldExcludeFile returns true if the file should be excluded from the analysis
//
//nolint:unused // Will be used in future implementation
func shouldExcludeFile(fileName string, config *Config) bool {
	// Skip test files
	if isTestFile(fileName, config) || strings.HasSuffix(fileName, "_test.go") {
		return true
	}

	// Skip files that match ignore patterns
	return shouldIgnoreFile(fileName, config)
}
