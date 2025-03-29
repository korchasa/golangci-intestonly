// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
package intestonly

import (
	"fmt"
	"reflect"

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
	ResultType:       reflect.TypeOf((*AnalysisResult)(nil)),
}

// run is the main entry point for the analyzer
func run(pass *analysis.Pass) (interface{}, error) {
	// Create config from optional settings
	config := getConfig(pass)

	// For debugging
	if config.Debug {
		fmt.Println("Running intestonly analyzer on package:", pass.Pkg.Path())
		fmt.Println("Total files:", len(pass.Files))
		for _, f := range pass.Files {
			fileName := pass.Fset.File(f.Pos()).Name()
			fmt.Printf("File: %s (isTest: %v)\n", fileName, isTestFile(fileName, config))
		}
	}

	// Create result container
	result := NewAnalysisResult()
	result.CurrentPkgPath = pass.Pkg.Path()

	// Step 1: Process declarations
	collectDeclarations(pass, result, config)

	// Step 2: If enabled, perform enhanced analysis of interfaces
	if config.EnableInterfaceImplementationDetection {
		analyzeInterfaceImplementations(pass, result, config)
	}

	// Step 3: Always build call graph (required for cross-package analysis)
	buildCallGraph(pass, result, config)

	// Step 4: If enabled, process export status
	if config.EnableExportedIdentifierHandling {
		processExportedIdentifiers(pass, result, config)
	}

	// Step 5: Analyze usages (now using the unified system)
	analyzeUsages(pass, result, config)

	// Step 6: Process cross-package references using call graph information
	analyzeCrossPackageReferences(pass, result, config)

	// Step 7: Analyze string references to detect functions mentioned in strings
	analyzeStringReferences(pass, result, config)

	// Step 8: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 9: Decide which items are used only in tests
	issues := generateIssues(pass, result, config)

	// Report all issues
	for _, issue := range issues {
		pass.Report(issue.ToAnalysisIssue(pass))
	}

	return result, nil
}

// AnalyzePackage analyzes the package and returns issues.
// This is an exported function for testing purposes.
func AnalyzePackage(pass *analysis.Pass, config *Config) []Issue {
	// For debugging
	if config.Debug {
		fmt.Println("Analyzing package:", pass.Pkg.Path())
		fmt.Println("Total files:", len(pass.Files))
		for _, f := range pass.Files {
			fileName := pass.Fset.File(f.Pos()).Name()
			fmt.Printf("File: %s (isTest: %v)\n", fileName, isTestFile(fileName, config))
		}
	}

	// Create result container
	result := NewAnalysisResult()
	result.CurrentPkgPath = pass.Pkg.Path()

	// Step 1: Process declarations
	collectDeclarations(pass, result, config)

	// Step 2: If enabled, perform enhanced analysis of interfaces
	if config.EnableInterfaceImplementationDetection {
		analyzeInterfaceImplementations(pass, result, config)
	}

	// Step 3: Always build call graph (required for cross-package analysis)
	buildCallGraph(pass, result, config)

	// Step 4: If enabled, process export status
	if config.EnableExportedIdentifierHandling {
		processExportedIdentifiers(pass, result, config)
	}

	// Step 5: Analyze usages (now using the unified system)
	analyzeUsages(pass, result, config)

	// Step 6: Process cross-package references using call graph information
	analyzeCrossPackageReferences(pass, result, config)

	// Step 7: Analyze string references to detect functions mentioned in strings
	analyzeStringReferences(pass, result, config)

	// Step 8: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 9: Generate issues
	return generateIssues(pass, result, config)
}
