package intestonly

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
)

// generateIssues identifies declarations that are only used in tests
func generateIssues(pass *analysis.Pass, result *AnalysisResult, config *Config) []Issue {
	var issues []Issue

	// Check each declaration
	for name, info := range result.Declarations {
		// Skip if it should be excluded
		if shouldExcludeFromReport(name, info, config) {
			continue
		}

		// Force report for explicit test cases if configured
		if config.ReportExplicitTestCases && isExplicitTestOnly(name, config) {
			issues = append(issues, Issue{
				Pos:     info.Pos,
				Message: fmt.Sprintf("identifier %q is only used in test files but is not part of test files", name),
			})
			continue
		}

		// Check if only used in tests
		hasTestUsages := len(result.TestUsages[name]) > 0
		hasNonTestUsages := len(result.NonTestUsages[name]) > 0

		if hasTestUsages && !hasNonTestUsages {
			issues = append(issues, Issue{
				Pos:     info.Pos,
				Message: fmt.Sprintf("identifier %q is only used in test files but is not part of test files", name),
			})
		}
	}

	return issues
}
