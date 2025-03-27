# 1. Add Synchronization for Parallel Processing in Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter processes files sequentially within a single function. The unused linter example demonstrates the use of synchronization mechanisms to safely accumulate results in a concurrent environment, which can improve performance on large codebases.

## Decision
Implement proper synchronization mechanisms in the intestonly linter to allow for parallel processing of files and safe aggregation of results.

## Implementation Details

### Parallel Processing Model

1. Introduce a mutex to protect shared data structures during result aggregation:

```go
var mu sync.Mutex
var resIssues []goanalysis.Issue

analyzer := &analysis.Analyzer{
    Name: "intestonly",
    Doc:  "Checks for code that is only used in tests but is not part of test files",
    Run: func(pass *analysis.Pass) (any, error) {
        issues := runIntestonly(pass, settings)
        if len(issues) == 0 {
            return nil, nil
        }

        mu.Lock()
        resIssues = append(resIssues, issues...)
        mu.Unlock()

        return nil, nil
    },
    // ...other properties
}
```

2. Split the analysis process into stages that can be parallelized:

```go
func runIntestonly(pass *analysis.Pass, settings *config.IntestOnlySettings) []goanalysis.Issue {
    // Stage 1: Collect declarations (thread-safe with local data)
    decls := collectDeclarations(pass)

    // Stage 2: Analyze usages (can be parallelized)
    testUsages, nonTestUsages := analyzeUsages(pass, decls)

    // Stage 3: Generate issues (local operation with results)
    return generateIssues(pass, decls, testUsages, nonTestUsages)
}
```

3. Use goroutines for file processing where appropriate:

```go
func analyzeUsages(pass *analysis.Pass, decls map[string]intestOnlyInfo) (map[string]bool, map[string]bool) {
    testUsages := make(map[string]bool)
    nonTestUsages := make(map[string]bool)

    var wg sync.WaitGroup
    var mu sync.Mutex

    for _, file := range pass.Files {
        wg.Add(1)
        go func(file *ast.File) {
            defer wg.Done()

            fileName := pass.Fset.File(file.Pos()).Name()
            isTest := isTestFile(fileName)

            localTestUsages := make(map[string]bool)
            localNonTestUsages := make(map[string]bool)

            // Analyze file for usages...

            // Merge results safely
            mu.Lock()
            for k, v := range localTestUsages {
                testUsages[k] = v
            }
            for k, v := range localNonTestUsages {
                nonTestUsages[k] = v
            }
            mu.Unlock()
        }(file)
    }

    wg.Wait()
    return testUsages, nonTestUsages
}
```

### Thread-Safe Result Structures

Create a dedicated result type to encapsulate the analysis results:

```go
type AnalysisResult struct {
    mu            sync.Mutex
    Declarations  map[string]intestOnlyInfo
    TestUsages    map[string]bool
    NonTestUsages map[string]bool
}

func (r *AnalysisResult) AddDeclaration(name string, info intestOnlyInfo) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.Declarations[name] = info
}

func (r *AnalysisResult) AddTestUsage(name string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.TestUsages[name] = true
}

func (r *AnalysisResult) AddNonTestUsage(name string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.NonTestUsages[name] = true
}
```

### Linter Registration

Update the linter registration to properly handle the accumulated results:

```go
func New(settings *config.IntestOnlySettings) *goanalysis.Linter {
    var mu sync.Mutex
    var resIssues []goanalysis.Issue

    analyzer := &analysis.Analyzer{
        // ...properties
    }

    return goanalysis.NewLinter(
        "intestonly",
        "Checks for code that is only used in tests but is not part of test files",
        []*analysis.Analyzer{analyzer},
        nil,
    ).WithIssuesReporter(func(_ *linter.Context) []goanalysis.Issue {
        return resIssues
    }).WithLoadMode(goanalysis.LoadModeTypesInfo)
}
```

## Consequences

### Positive
- Improved performance on large codebases through parallel processing
- More efficient use of multi-core processors
- Reduced analysis time in CI/CD pipelines
- Aligns with modern Go concurrency patterns

### Negative
- Increased complexity in the code structure
- Potential for concurrency bugs if synchronization is not implemented correctly
- May require additional memory for local result aggregation

### Mitigations
- Comprehensive testing of concurrent behavior
- Use of established concurrency patterns from the Go standard library
- Careful profiling to ensure performance improvements are realized

## References
- Unused linter implementation in golangci-lint
- Go Concurrency Patterns: https://blog.golang.org/pipelines
- Effective Go - Concurrency: https://golang.org/doc/effective_go.html#concurrency