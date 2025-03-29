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

// Global variable to store results across runs
var globalResult *AnalysisResult

// ResetGlobalResult resets the global result to nil
// This is useful for tests to ensure each test starts with a fresh state
func ResetGlobalResult() {
	globalResult = nil
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

	// Initialize global result if it doesn't exist
	if globalResult == nil {
		globalResult = NewAnalysisResult()
	}

	// Create a new result container for this run
	result := NewAnalysisResult()

	// Always update the current package path
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

	// Step 9: Merge results from this run with global results
	mergeResults(globalResult, result)

	// Step 10: Decide which items are used only in tests
	issues := generateIssues(pass, globalResult, config)

	// Report all issues
	for _, issue := range issues {
		pass.Report(issue.ToAnalysisIssue(pass))
	}

	return globalResult, nil
}

// mergeResults merges the results from the current run into the global results
func mergeResults(global, current *AnalysisResult) {
	// Merge declarations
	for name, decl := range current.Declarations {
		global.Declarations[name] = decl
	}

	// Merge test usages
	for name, usages := range current.TestUsages {
		global.TestUsages[name] = append(global.TestUsages[name], usages...)
	}

	// Merge non-test usages
	for name, usages := range current.Usages {
		global.Usages[name] = append(global.Usages[name], usages...)
	}

	// Merge declaration positions
	for pos, name := range current.DeclPositions {
		global.DeclPositions[pos] = name
	}

	// Merge import references
	for ref, name := range current.ImportRefs {
		global.ImportRefs[ref] = name
	}

	// Merge imported packages
	for name, path := range current.ImportedPkgs {
		global.ImportedPkgs[name] = path
	}

	// Merge call graph
	for caller, callees := range current.CallGraph {
		global.CallGraph[caller] = append(global.CallGraph[caller], callees...)
	}

	// Merge called by
	for callee, callers := range current.CalledBy {
		global.CalledBy[callee] = append(global.CalledBy[callee], callers...)
	}

	// Merge interfaces
	for name, methods := range current.Interfaces {
		global.Interfaces[name] = append(global.Interfaces[name], methods...)
	}

	// Merge implementations
	for name, impls := range current.Implementations {
		global.Implementations[name] = append(global.Implementations[name], impls...)
	}

	// Merge methods of type
	for name, methods := range current.MethodsOfType {
		global.MethodsOfType[name] = append(global.MethodsOfType[name], methods...)
	}

	// Merge exported declarations
	for name, exported := range current.ExportedDecls {
		global.ExportedDecls[name] = exported
	}

	// Merge cross-package references
	for ref, used := range current.CrossPackageTestRefs {
		global.CrossPackageTestRefs[ref] = used
	}
	for ref, used := range current.CrossPackageRefs {
		global.CrossPackageRefs[ref] = used
	}
	for path, refs := range current.CrossPackageRefsList {
		global.CrossPackageRefsList[path] = append(global.CrossPackageRefsList[path], refs...)
	}

	// Merge package imports
	for path, imports := range current.PackageImports {
		if _, exists := global.PackageImports[path]; !exists {
			global.PackageImports[path] = make(map[string]bool)
		}
		for imp, used := range imports {
			global.PackageImports[path][imp] = used
		}
	}

	// Merge test package imports
	for path, imports := range current.TestPackageImports {
		if _, exists := global.TestPackageImports[path]; !exists {
			global.TestPackageImports[path] = make(map[string]bool)
		}
		for imp, used := range imports {
			global.TestPackageImports[path][imp] = used
		}
	}
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

	// Initialize global result if it doesn't exist
	if globalResult == nil {
		globalResult = NewAnalysisResult()
	}

	// Create result container for this run
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

	// Step 9: Merge results from this run with global results
	mergeResults(globalResult, result)

	// Step 10: Generate issues
	return generateIssues(pass, globalResult, config)
}
