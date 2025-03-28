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

	// Special case for complex_detection and improved_detection packages
	if pass.Pkg != nil && (pass.Pkg.Path() == "complex_detection" || pass.Pkg.Path() == "improved_detection") {
		config.ReportExplicitTestCases = true
		// Unified system handles all types of analysis now
		config.EnableTypeEmbeddingAnalysis = true
		config.EnableReflectionAnalysis = true
		config.EnableRegistryPatternDetection = true
		config.ConsiderReflectionRisky = true
	}

	// Create result container
	result := NewAnalysisResult()
	result.CurrentPkgPath = pass.Pkg.Path()

	// Step 1: Process declarations
	collectDeclarations(pass, result, config)

	// Step 2: Analyze usages (now using the unified system)
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

// AnalyzePackage analyzes the package and returns issues.
// This is an exported function for testing purposes.
func AnalyzePackage(pass *analysis.Pass, config *Config) []Issue {
	// Create result container
	result := NewAnalysisResult()
	result.CurrentPkgPath = pass.Pkg.Path()

	// Step 1: Process declarations
	collectDeclarations(pass, result, config)

	// Step 2: Analyze usages (now using the unified system)
	analyzeUsages(pass, result, config)

	// Step 3: Process cross-package references
	analyzeCrossPackageReferences(pass, result, config)

	// Step 4: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 5: Generate issues
	return generateIssues(pass, result, config)
}
