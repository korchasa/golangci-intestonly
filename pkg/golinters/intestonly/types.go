package intestonly

import (
	"go/token"

	"golang.org/x/tools/go/analysis"
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

// ToAnalysisIssue converts an internal issue to a diagnostic for reporting
// This method can be used for integration with golangci-lint if needed
func (i *Issue) ToAnalysisIssue(pass *analysis.Pass) analysis.Diagnostic {
	return analysis.Diagnostic{
		Pos:      i.Pos,
		Message:  i.Message,
		Category: "intestonly",
	}
}

// Config holds configuration options for the intestonly analyzer
type Config struct {
	// Whether to check methods (functions with receivers)
	CheckMethods bool

	// Whether to ignore unexported identifiers
	IgnoreUnexported bool

	// Whether to enable content-based usage detection
	// (checking for identifiers in file content)
	EnableContentBasedDetection bool

	// Whether to exclude test helpers from reporting
	ExcludeTestHelpers bool

	// Whether to output debug information
	Debug bool

	// Custom patterns for identifying test helpers
	TestHelperPatterns []string

	// Patterns for files to ignore in analysis
	IgnoreFilePatterns []string

	// Patterns for identifiers to always exclude from reporting
	ExcludePatterns []string

	// List of explicit test-only identifiers that should always be reported
	ExplicitTestOnlyIdentifiers []string

	// Whether to report explicit test-only identifiers regardless of usage
	ReportExplicitTestCases bool

	// Whether to enable type embedding analysis
	EnableTypeEmbeddingAnalysis bool

	// Whether to enable reflection usage detection
	EnableReflectionAnalysis bool

	// Whether to consider reflection-based access as a usage risk
	ConsiderReflectionRisky bool

	// Whether to enable detection of registry patterns
	EnableRegistryPatternDetection bool
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
