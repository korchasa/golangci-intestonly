// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
package intestonly

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer is the analyzer struct.
var Analyzer = &analysis.Analyzer{
	Name: "intestonly",
	Doc:  "Checks for code that is only used in tests but is not part of test files",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes:        []analysis.Fact{},
	RunDespiteErrors: true,
}

// run is the main entry point for the analyzer
func run(pass *analysis.Pass) (interface{}, error) {
	// Create config from optional settings
	config := getConfig(pass)

	// Create result container
	result := NewAnalysisResult()
	result.CurrentPkgPath = pass.Pkg.Path()

	// Step 1: Process declarations
	collectDeclarations(pass, result, config)

	// Step 2: Analyze usages
	analyzeUsages(pass, result, config)

	// Step 3: Process cross-package references
	analyzeCrossPackageReferences(pass, result, config)

	// Step 4: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 5: Generate and report issues
	issues := generateIssues(pass, result, config)
	for _, issue := range issues {
		// Use ToAnalysisIssue to convert the issue to a standard diagnostic
		diag := issue.ToAnalysisIssue(pass)
		pass.Report(diag)
	}

	return nil, nil
}
