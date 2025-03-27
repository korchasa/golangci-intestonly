# 2. Structured Result Types for Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter uses several independent maps to track declarations and usages. The unused linter example demonstrates a more structured approach with dedicated result types that encapsulate analysis results.

## Decision
Implement structured result types to better organize the data collected during analysis and improve the maintainability of the codebase.

## Implementation Details

### Core Result Types

1. Define a primary structure to hold declaration information:

```go
// DeclInfo holds information about a declaration in the codebase
type DeclInfo struct {
    Pos       token.Pos // Position of the declaration
    Name      string    // Name of the identifier
    FilePath  string    // Path to the file containing the declaration
    IsMethod  bool      // Whether this is a method
    PkgPath   string    // Path of the package containing the declaration
    ImportRef string    // Full import reference (pkgPath.name)
}
```

2. Create a comprehensive result structure to track analysis state:

```go
// AnalysisResult holds all collected data during the analysis
type AnalysisResult struct {
    // All declarations found in non-test files
    Declarations map[string]DeclInfo

    // Map tracking identifiers used in test files
    TestUsages map[string][]token.Pos

    // Map tracking identifiers used in non-test files
    NonTestUsages map[string][]token.Pos

    // Map of positions to identifier names to skip self-references
    DeclPositions map[token.Pos]string

    // Map associating import paths with their identifiers
    ImportRefs map[string]string

    // Map of imported package names to their paths
    ImportedPkgNames map[string]string
}

// NewAnalysisResult creates a new empty analysis result
func NewAnalysisResult() *AnalysisResult {
    return &AnalysisResult{
        Declarations:     make(map[string]DeclInfo),
        TestUsages:       make(map[string][]token.Pos),
        NonTestUsages:    make(map[string][]token.Pos),
        DeclPositions:    make(map[token.Pos]string),
        ImportRefs:       make(map[string]string),
        ImportedPkgNames: make(map[string]string),
    }
}
```

3. Define a structure for reported issues:

```go
// Issue represents a linter issue to be reported
type Issue struct {
    Pos     token.Pos // Position where the issue was detected
    Message string    // Message describing the issue
}

// ToAnalysisIssue converts an internal issue to the golangci-lint issue format
func (i *Issue) ToAnalysisIssue(pass *analysis.Pass) goanalysis.Issue {
    return goanalysis.NewIssue(&result.Issue{
        FromLinter: "intestonly",
        Text:       i.Message,
        Pos:        pass.Fset.Position(i.Pos),
    }, pass)
}
```

### Usage Example

The refactored analyzer would use these types as follows:

```go
func run(pass *analysis.Pass) (interface{}, error) {
    // Create a new analysis result
    result := NewAnalysisResult()

    // Step 1: Collect declarations from non-test files
    collectDeclarations(pass, result)

    // Step 2: Analyze usages in all files
    analyzeUsages(pass, result)

    // Step 3: Generate issues for test-only identifiers
    issues := generateIssues(pass, result)

    // Step 4: Report the issues
    for _, issue := range issues {
        pass.Reportf(issue.Pos, issue.Message)
    }

    return nil, nil
}

func collectDeclarations(pass *analysis.Pass, result *AnalysisResult) {
    // Implementation to collect declarations into result.Declarations
}

func analyzeUsages(pass *analysis.Pass, result *AnalysisResult) {
    // Implementation to analyze usages into result.TestUsages and result.NonTestUsages
}

func generateIssues(pass *analysis.Pass, result *AnalysisResult) []Issue {
    var issues []Issue

    for name, info := range result.Declarations {
        // Skip special cases...

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
```

### Configuration Type

Additionally, create a configuration type to centralize settings:

```go
// Config holds the configuration for the intestonly linter
type Config struct {
    // Whether to check methods
    CheckMethods bool

    // Whether to enable content-based usage detection
    EnableContentBasedDetection bool

    // Whether to exclude test helpers
    ExcludeTestHelpers bool

    // Custom patterns for test helper identification
    TestHelperPatterns []string

    // Debug mode
    Debug bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    return &Config{
        CheckMethods:               true,
        EnableContentBasedDetection: true,
        ExcludeTestHelpers:         true,
        TestHelperPatterns: []string{
            "assert",
            "mock",
            "fake",
            "stub",
            "setup",
            "cleanup",
        },
        Debug: false,
    }
}
```

## Consequences

### Positive
- Improved code organization and readability
- Better encapsulation of related data
- Clearer interfaces between components of the linter
- Easier to maintain and extend
- Enhanced type safety and reduced risk of errors
- Better documentation of data structures through types

### Negative
- Initial implementation overhead
- Additional code to maintain
- Potential performance overhead from struct creation vs. direct map usage

### Mitigations
- Ensure efficient memory usage with proper initialization
- Consider using pointer semantics where appropriate to avoid large copies
- Implement benchmarks to verify performance is not degraded

## References
- Unused linter implementation in golangci-lint
- Go documentation on struct types and type systems
- Effective Go guidelines on struct composition