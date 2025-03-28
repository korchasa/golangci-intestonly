// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
package intestonly

import (
	"fmt"

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

	// Step 3: If enabled, build call graph
	if config.EnableCallGraphAnalysis {
		buildCallGraph(pass, result, config)
	}

	// Step 4: If enabled, process export status
	if config.EnableExportedIdentifierHandling {
		processExportedIdentifiers(pass, result, config)
	}

	// Step 5: Analyze usages (now using the unified system)
	analyzeUsages(pass, result, config)

	// Step 6: Process cross-package references (enhanced if enabled)
	if config.EnableRobustCrossPackageAnalysis {
		analyzeRobustCrossPackageReferences(pass, result, config)
	} else {
		analyzeCrossPackageReferences(pass, result, config)
	}

	// Step 7: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 8: Generate and report issues
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

	// Step 3: If enabled, build call graph
	if config.EnableCallGraphAnalysis {
		buildCallGraph(pass, result, config)
	}

	// Step 4: If enabled, process export status
	if config.EnableExportedIdentifierHandling {
		processExportedIdentifiers(pass, result, config)
	}

	// Step 5: Analyze usages (now using the unified system)
	analyzeUsages(pass, result, config)

	// Step 6: Process cross-package references (enhanced if enabled)
	if config.EnableRobustCrossPackageAnalysis {
		analyzeRobustCrossPackageReferences(pass, result, config)
	} else {
		analyzeCrossPackageReferences(pass, result, config)
	}

	// Step 7: Additional content-based analysis if enabled
	if config.EnableContentBasedDetection {
		analyzeContentBasedUsages(pass, result, config)
	}

	// Step 8: Generate issues
	return generateIssues(pass, result, config)
}
