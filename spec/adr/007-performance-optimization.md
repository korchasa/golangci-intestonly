# 7. Performance Optimization for Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter has several performance issues that make it impractical for use on large codebases:

1. **Inefficient Resource Management**: The linter repeatedly reads files from disk, does not cache results, and uses unoptimized data structures, causing excessive memory and CPU consumption.

2. **Lack of Concurrency**: The analysis is performed sequentially, not taking advantage of modern multi-core processors.

3. **Excessive AST Traversal**: The code traverses the AST multiple times, doing redundant work rather than collecting all needed information in a single pass.

4. **Suboptimal Error Handling**: Error handling is minimal, with many errors simply being ignored, leading to potential silent failures and incomplete analysis.

5. **No Incremental Analysis**: Every run analyzes the entire codebase, even if only a small portion has changed.

These issues can cause the linter to be prohibitively slow on large codebases, leading users to disable it or avoid using it altogether.

## Decision
Implement comprehensive performance optimizations in the intestonly linter to make it practical for use with large codebases.

## Implementation Details

### Optimized Data Structures

1. Replace simple maps with more efficient, purpose-built structures:

```go
// IdentifierUseMap provides fast lookup of identifier uses
type IdentifierUseMap struct {
    // Internal storage maps identifier names to positions
    // Using a slice of positions rather than a map for memory efficiency
    uses map[string][]token.Pos

    // Mutex for concurrent access
    mu sync.RWMutex
}

// NewIdentifierUseMap creates a new, initialized identifier use map
func NewIdentifierUseMap() *IdentifierUseMap {
    return &IdentifierUseMap{
        uses: make(map[string][]token.Pos),
    }
}

// Add adds a position where an identifier is used
func (m *IdentifierUseMap) Add(name string, pos token.Pos) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.uses[name] = append(m.uses[name], pos)
}

// Has checks if an identifier has any uses
func (m *IdentifierUseMap) Has(name string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    _, exists := m.uses[name]
    return exists && len(m.uses[name]) > 0
}

// Get returns all positions where an identifier is used
func (m *IdentifierUseMap) Get(name string) []token.Pos {
    m.mu.RLock()
    defer m.mu.RUnlock()

    return m.uses[name]
}
```

### Parallel Processing

1. Implement concurrent analysis of files:

```go
// analyzeFilesInParallel processes files concurrently
func analyzeFilesInParallel(pass *analysis.Pass, config *Config) *AnalysisResult {
    result := NewAnalysisResult()

    // Limit concurrency based on available processors
    maxWorkers := runtime.NumCPU()
    if config.MaxWorkers > 0 && config.MaxWorkers < maxWorkers {
        maxWorkers = config.MaxWorkers
    }

    // Create a worker pool
    var wg sync.WaitGroup
    fileCh := make(chan *ast.File, maxWorkers)

    // Start worker goroutines
    for i := 0; i < maxWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            for file := range fileCh {
                fileName := pass.Fset.File(file.Pos()).Name()
                isTest := isTestFile(fileName)

                // Skip ignored files
                if shouldIgnoreFile(fileName, config) {
                    continue
                }

                // Process the file
                if !isTest {
                    collectDeclarationsFromFile(file, fileName, pass, result)
                }

                analyzeUsagesInFile(file, fileName, isTest, pass, result)
            }
        }()
    }

    // Feed files to workers
    for _, file := range pass.Files {
        fileCh <- file
    }
    close(fileCh)

    // Wait for all workers to finish
    wg.Wait()

    // Process cross-package references and other global analyses sequentially
    // as they depend on the collected declarations
    analyzeCrossPackageReferences(pass, result, config)

    // If enabled, process content-based detection
    if config.EnableContentBasedDetection {
        analyzeContentBasedUsages(pass, result, config)
    }

    return result
}
```

### File Content Caching

1. Implement an efficient file content cache:

```go
// FileCache provides cached access to file contents
type FileCache struct {
    // Map from file path to content
    contents map[string][]byte

    // Mutex for concurrent access
    mu sync.RWMutex
}

// NewFileCache creates a new file cache
func NewFileCache() *FileCache {
    return &FileCache{
        contents: make(map[string][]byte),
    }
}

// Get returns file content, reading from disk if not cached
func (c *FileCache) Get(filePath string) ([]byte, error) {
    c.mu.RLock()
    content, ok := c.contents[filePath]
    c.mu.RUnlock()

    if ok {
        return content, nil
    }

    // Read from disk
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    // Cache for future use
    c.mu.Lock()
    c.contents[filePath] = content
    c.mu.Unlock()

    return content, nil
}

// HasIdentifier checks if a file contains an identifier without loading full content
func (c *FileCache) HasIdentifier(filePath string, identifier string) (bool, error) {
    content, err := c.Get(filePath)
    if err != nil {
        return false, err
    }

    // Use more sophisticated check than simple string matching
    // This is just a placeholder - real implementation would use regex with word boundaries
    pattern := "\\b" + regexp.QuoteMeta(identifier) + "\\b"
    matched, err := regexp.Match(pattern, content)
    return matched, err
}
```

### Single-Pass AST Traversal

1. Consolidate AST traversal into a single, comprehensive pass:

```go
// processFileComprehensive performs a single-pass analysis of a file
func processFileComprehensive(file *ast.File, fileName string, isTest bool, pass *analysis.Pass, result *AnalysisResult, config *Config) {
    // Track local declarations to handle shadowing
    localDecls := make(map[string]token.Pos)

    // Process imports first
    for _, imp := range file.Imports {
        if imp.Path != nil {
            importPath := strings.Trim(imp.Path.Value, "\"")
            var pkgName string
            if imp.Name != nil {
                pkgName = imp.Name.Name
            } else {
                // Extract package name from import path
                parts := strings.Split(importPath, "/")
                pkgName = parts[len(parts)-1]
            }

            result.AddImport(pkgName, importPath)
        }
    }

    // Track current package
    pkgPath := pass.Pkg.Path()

    // First pass to collect declarations
    if !isTest {
        ast.Inspect(file, func(n ast.Node) bool {
            switch node := n.(type) {
            case *ast.FuncDecl:
                if node.Name != nil && node.Name.Name != "" {
                    name := node.Name.Name

                    // Skip test helpers
                    if isTestHelperIdentifier(name, config) && !isExplicitTestOnly(name, config) {
                        return true
                    }

                    // Record as local declaration
                    localDecls[name] = node.Name.Pos()

                    importRef := pkgPath + "." + name
                    result.AddImportRef(importRef, name)

                    // Add to declarations
                    if node.Recv != nil && len(node.Recv.List) > 0 {
                        // Method
                        result.AddDeclaration(name, &DeclInfo{
                            Pos:       node.Name.Pos(),
                            Name:      name,
                            FilePath:  fileName,
                            IsMethod:  true,
                            PkgPath:   pkgPath,
                            ImportRef: importRef,
                        })
                    } else {
                        // Function
                        result.AddDeclaration(name, &DeclInfo{
                            Pos:       node.Name.Pos(),
                            Name:      name,
                            FilePath:  fileName,
                            IsMethod:  false,
                            PkgPath:   pkgPath,
                            ImportRef: importRef,
                        })
                    }
                }
            // Handle other declaration types (types, constants, variables)
            // ...
            }
            return true
        })
    }

    // Second pass to check usages
    ast.Inspect(file, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.Ident:
            // Skip if this is a declaration
            if _, isDeclPos := localDecls[node.Name]; isDeclPos && node.Pos() == localDecls[node.Name] {
                return true
            }

            // Check if this is a reference to one of our declarations
            obj := pass.TypesInfo.Uses[node]
            if obj != nil {
                declName := obj.Name()
                if result.HasDeclaration(declName) {
                    // Record usage
                    if isTest {
                        result.AddTestUsage(declName, node.Pos())
                    } else {
                        result.AddNonTestUsage(declName, node.Pos())
                    }
                }
            }

        // Handle selector expressions, method calls, etc.
        // ...
        }
        return true
    })
}
```

### Incremental Analysis

1. Implement incremental analysis for repeated runs:

```go
// ResultCache stores analysis results for incremental processing
type ResultCache struct {
    // Map from file path to last modification time
    fileModTimes map[string]time.Time

    // Last full analysis result
    lastResult *AnalysisResult

    // Mutex for concurrent access
    mu sync.RWMutex
}

// NewResultCache creates a new result cache
func NewResultCache() *ResultCache {
    return &ResultCache{
        fileModTimes: make(map[string]time.Time),
    }
}

// ShouldAnalyzeFile determines if a file needs to be analyzed
func (c *ResultCache) ShouldAnalyzeFile(filePath string) bool {
    c.mu.RLock()
    lastModTime, exists := c.fileModTimes[filePath]
    c.mu.RUnlock()

    if !exists {
        return true
    }

    // Check if file has been modified
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        // If we can't stat the file, analyze it to be safe
        return true
    }

    return fileInfo.ModTime().After(lastModTime)
}

// UpdateFileStatus updates the modification time of a file
func (c *ResultCache) UpdateFileStatus(filePath string) {
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return
    }

    c.mu.Lock()
    c.fileModTimes[filePath] = fileInfo.ModTime()
    c.mu.Unlock()
}

// SetLastResult stores the last full analysis result
func (c *ResultCache) SetLastResult(result *AnalysisResult) {
    c.mu.Lock()
    c.lastResult = result
    c.mu.Unlock()
}

// GetLastResult retrieves the last full analysis result
func (c *ResultCache) GetLastResult() *AnalysisResult {
    c.mu.RLock()
    defer c.mu.RUnlock()

    return c.lastResult
}
```

### Error Handling and Reporting

1. Implement robust error handling:

```go
// ErrorCollector collects and aggregates errors during analysis
type ErrorCollector struct {
    // Errors encountered during analysis
    errors []error

    // Warning messages
    warnings []string

    // Mutex for concurrent access
    mu sync.Mutex
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
    return &ErrorCollector{
        errors:   make([]error, 0),
        warnings: make([]string, 0),
    }
}

// AddError adds an error to the collector
func (c *ErrorCollector) AddError(err error) {
    if err == nil {
        return
    }

    c.mu.Lock()
    c.errors = append(c.errors, err)
    c.mu.Unlock()
}

// AddWarning adds a warning message
func (c *ErrorCollector) AddWarning(warning string) {
    if warning == "" {
        return
    }

    c.mu.Lock()
    c.warnings = append(c.warnings, warning)
    c.mu.Unlock()
}

// HasErrors checks if any errors were collected
func (c *ErrorCollector) HasErrors() bool {
    c.mu.Lock()
    defer c.mu.Unlock()

    return len(c.errors) > 0
}

// GetErrors returns all collected errors
func (c *ErrorCollector) GetErrors() []error {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.errors
}

// GetWarnings returns all collected warnings
func (c *ErrorCollector) GetWarnings() []string {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.warnings
}
```

### Memory Optimization

1. Implement memory-efficient analysis:

```go
// MemoryOptimizedAnalyzer manages memory usage during analysis
type MemoryOptimizedAnalyzer struct {
    // Configuration
    config *Config

    // Analysis results
    result *AnalysisResult

    // File cache
    fileCache *FileCache

    // Result cache for incremental analysis
    resultCache *ResultCache

    // Error collector
    errorCollector *ErrorCollector
}

// NewMemoryOptimizedAnalyzer creates a new memory-optimized analyzer
func NewMemoryOptimizedAnalyzer(config *Config) *MemoryOptimizedAnalyzer {
    return &MemoryOptimizedAnalyzer{
        config:         config,
        result:         NewAnalysisResult(),
        fileCache:      NewFileCache(),
        resultCache:    NewResultCache(),
        errorCollector: NewErrorCollector(),
    }
}

// Analyze performs memory-efficient analysis
func (a *MemoryOptimizedAnalyzer) Analyze(pass *analysis.Pass) ([]Issue, error) {
    // Check if we can use incremental analysis
    if a.config.EnableIncrementalAnalysis && a.resultCache.GetLastResult() != nil {
        return a.analyzeIncrementally(pass)
    }

    // Perform full analysis
    return a.analyzeAll(pass)
}

// analyzeAll performs a full analysis of all files
func (a *MemoryOptimizedAnalyzer) analyzeAll(pass *analysis.Pass) ([]Issue, error) {
    // Process files in batches to control memory usage
    const batchSize = 20

    files := pass.Files
    totalFiles := len(files)

    for i := 0; i < totalFiles; i += batchSize {
        end := i + batchSize
        if end > totalFiles {
            end = totalFiles
        }

        // Process batch
        batch := files[i:end]
        a.processBatch(batch, pass)

        // Release memory after processing batch
        runtime.GC()
    }

    // Update result cache for incremental analysis
    if a.config.EnableIncrementalAnalysis {
        a.resultCache.SetLastResult(a.result)

        // Update file modification times
        for _, file := range files {
            fileName := pass.Fset.File(file.Pos()).Name()
            a.resultCache.UpdateFileStatus(fileName)
        }
    }

    // Generate issues
    issues := generateIssues(pass, a.result, a.config)

    // Check for errors
    if a.errorCollector.HasErrors() {
        // Combine errors into a single error
        var errMsgs []string
        for _, err := range a.errorCollector.GetErrors() {
            errMsgs = append(errMsgs, err.Error())
        }

        // Return issues along with a combined error
        return issues, fmt.Errorf("analysis completed with errors: %s", strings.Join(errMsgs, "; "))
    }

    return issues, nil
}

// processBatch processes a batch of files
func (a *MemoryOptimizedAnalyzer) processBatch(files []*ast.File, pass *analysis.Pass) {
    // Process files concurrently within batch
    var wg sync.WaitGroup

    for _, file := range files {
        wg.Add(1)
        go func(f *ast.File) {
            defer wg.Done()

            fileName := pass.Fset.File(f.Pos()).Name()
            isTest := isTestFile(fileName)

            // Skip ignored files
            if shouldIgnoreFile(fileName, a.config) {
                return
            }

            // Process file
            processFileComprehensive(f, fileName, isTest, pass, a.result, a.config)
        }(file)
    }

    wg.Wait()
}

// Remaining methods for incremental analysis...
```

## Consequences

### Positive
- Dramatically improved performance on large codebases
- Reduced memory consumption
- Better utilization of multi-core processors
- Shorter analysis times in CI/CD pipelines
- Better error reporting and handling
- Improved user experience through incremental analysis
- Higher adoption rate due to practical performance

### Negative
- Increased complexity in implementation
- Potential for subtle concurrency bugs
- More difficult debugging of internal issues
- Additional configuration options to manage

### Mitigations
- Comprehensive testing, especially for concurrency issues
- Performance benchmarks to validate optimizations
- Memory profiling to identify and fix leaks
- Clear documentation of performance characteristics
- Sensible defaults that work well for most codebases
- Detailed logging to help with troubleshooting

## References
- Go Concurrency Patterns: https://blog.golang.org/pipelines
- Performance optimization in Go: https://github.com/dgryski/go-perfbook
- Memory management in Go: https://golang.org/pkg/runtime/
- Effective concurrency in Go: https://golang.org/doc/effective_go.html#concurrency
- Go garbage collector guide: https://tip.golang.org/doc/gc-guide