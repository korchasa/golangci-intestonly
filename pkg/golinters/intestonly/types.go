package intestonly

import (
	"go/token"
)

// DeclInfo stores information about a declaration
type DeclInfo struct {
	Pos       token.Pos // Position of the declaration
	Name      string    // Name of the identifier
	FilePath  string    // Path to the file containing the declaration
	IsMethod  bool      // Whether the declaration is a method
	PkgPath   string    // Path of the package containing the declaration
	ImportRef string    // Full import reference (pkgPath.name)
}

// Issue represents a linter issue to be reported
type Issue struct {
	Pos     token.Pos // Position where the issue should be reported
	Message string    // Message describing the issue
}

// Config holds configuration options for the intestonly analyzer
type Config struct {
	Debug                       bool     // Enable debug output
	CheckMethods                bool     // Check method declarations
	ReportExplicitTestCases     bool     // Always report test cases from testdata
	ExcludeTestHelpers          bool     // Exclude identifiers that look like test helpers
	EnableContentBasedDetection bool     // Enable detection based on file contents
	ExcludePatterns             []string // Patterns to exclude from reporting
}

// AnalysisResult holds the results of the analysis
type AnalysisResult struct {
	Declarations   map[string]DeclInfo    // All declarations in non-test files
	TestUsages     map[string][]token.Pos // Identifiers used in test files
	NonTestUsages  map[string][]token.Pos // Identifiers used in non-test files
	DeclPositions  map[token.Pos]string   // Map positions to identifiers to skip self-references
	ImportRefs     map[string]string      // Map import path with identifier to full reference
	ImportedPkgs   map[string]string      // Map imported package name to its path
	CurrentPkgPath string                 // Current package path being analyzed
}

// NewAnalysisResult creates a new AnalysisResult instance
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Declarations:  make(map[string]DeclInfo),
		TestUsages:    make(map[string][]token.Pos),
		NonTestUsages: make(map[string][]token.Pos),
		DeclPositions: make(map[token.Pos]string),
		ImportRefs:    make(map[string]string),
		ImportedPkgs:  make(map[string]string),
	}
}
