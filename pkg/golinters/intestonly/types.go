package intestonly

import (
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// DeclType represents the type of declaration
type DeclType int

const (
	DeclUnknown DeclType = iota
	DeclFunction
	DeclMethod
	DeclTypeDecl
	DeclConstant
	DeclVariable
)

// DeclInfo stores information about a declaration
type DeclInfo struct {
	Pos          token.Pos // Position of the declaration
	Name         string    // Name of the declaration
	FilePath     string    // File path where the declaration is located
	IsMethod     bool      // Whether this is a method
	PkgPath      string    // Package path
	ImportRef    string    // Import reference (e.g., "fmt.Println")
	DeclType     DeclType  // Type of declaration
	ReceiverType string    // Type name of the receiver for methods
	Comment      string    // Documentation comment associated with the declaration
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

	// Whether to enable call graph analysis for more accurate dependency tracking
	EnableCallGraphAnalysis bool

	// Whether to enable enhanced interface implementation detection
	EnableInterfaceImplementationDetection bool

	// Whether to perform deeper cross-package reference analysis
	EnableRobustCrossPackageAnalysis bool

	// Whether to apply special handling for exported identifiers
	EnableExportedIdentifierHandling bool

	// Whether to consider exported constants used even if no direct usage is found
	ConsiderExportedConstantsUsed bool

	// Additional test files patterns to consider as test files
	AdditionalTests []string

	// List of files to ignore
	IgnoreFiles []string

	// List of directories to ignore
	IgnoreDirectories []string

	// List of explicit test cases to check
	ExplicitTestCases []string

	// List of patterns to identify files to ignore
	IgnoreDirPatterns []string

	// Current file being processed
	CurrentFile string
}

// AnalysisResult holds the results of the analysis
type AnalysisResult struct {
	Declarations   map[string]DeclInfo    // All declarations in non-test files
	TestUsages     map[string][]UsageInfo // Identifiers used in test files
	Usages         map[string][]UsageInfo // Identifiers used in non-test files
	DeclPositions  map[token.Pos]string   // Map positions to identifiers to skip self-references
	ImportRefs     map[string]string      // Map import path with identifier to full reference
	ImportedPkgs   map[string]string      // Map imported package name to its path
	CurrentPkgPath string                 // Current package path being analyzed

	// Call graph tracking
	CallGraph map[string][]string // Maps function to functions it calls
	CalledBy  map[string][]string // Maps function to functions that call it

	// Interface implementations
	Interfaces      map[string][]string // Maps interface name to its method names
	Implementations map[string][]string // Maps interface name to types that implement it
	MethodsOfType   map[string][]string // Maps type name to its methods

	// Export tracking
	ExportedDecls map[string]bool // Set of exported declarations

	// Cross-package reference tracking
	CrossPackageTestRefs map[string]bool // Map of references from test files in other packages
	CrossPackageRefs     map[string]bool // Map of references from non-test files in other packages
}

// UsageInfo stores information about where a declaration is used
type UsageInfo struct {
	Pos      token.Pos // Position of the usage
	FilePath string    // File path where the usage occurs
	IsTest   bool      // Whether the usage is in a test file
}

// NewAnalysisResult creates a new AnalysisResult instance
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Declarations:         make(map[string]DeclInfo),
		TestUsages:           make(map[string][]UsageInfo),
		Usages:               make(map[string][]UsageInfo),
		DeclPositions:        make(map[token.Pos]string),
		ImportRefs:           make(map[string]string),
		ImportedPkgs:         make(map[string]string),
		CallGraph:            make(map[string][]string),
		CalledBy:             make(map[string][]string),
		Interfaces:           make(map[string][]string),
		Implementations:      make(map[string][]string),
		MethodsOfType:        make(map[string][]string),
		ExportedDecls:        make(map[string]bool),
		CrossPackageTestRefs: make(map[string]bool),
		CrossPackageRefs:     make(map[string]bool),
	}
}
