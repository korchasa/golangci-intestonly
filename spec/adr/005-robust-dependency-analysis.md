# 5. Robust Dependency Analysis in Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter has significant limitations in dependency analysis:

1. **Absence of Call Graph Analysis**: The linter doesn't construct a call graph, leading to missed relationships between functions. If function A is only used in tests, but calls function B that's used in production code, the linter will incorrectly flag function B.

2. **Poor Interface Implementation Detection**: The linter fails to recognize when methods implement interfaces used in production code, causing false positives when flagging methods as "test-only."

3. **Inadequate Cross-Package Reference Analysis**: The current analysis of cross-package references is primitive and fails to handle alias imports, complex dependencies, and implicit interface implementations.

4. **Missing Exported Identifier Context**: Code marked as exported (capitalized identifiers) is often meant for consumption by other packages, but the linter does not consider this special status in its analysis.

These limitations severely impact the accuracy of the linter's results, causing false positives that reduce user trust in the tool.

## Decision
Implement a robust dependency analysis system in the intestonly linter that addresses all the identified shortcomings.

## Implementation Details

### Call Graph Construction

1. Integrate with Go's callgraph package to build a proper call graph:

```go
import (
    "golang.org/x/tools/go/callgraph"
    "golang.org/x/tools/go/callgraph/cha"
    "golang.org/x/tools/go/callgraph/rta"
    "golang.org/x/tools/go/ssa"
)

// buildCallGraph constructs a call graph for the package being analyzed
func buildCallGraph(pass *analysis.Pass) *callgraph.Graph {
    prog := ssa.NewProgram(pass.Fset, ssa.BuilderMode(ssa.SanityCheckFunctions))

    // Build SSA packages for all imported packages
    for _, pkg := range pass.Pkg.Imports() {
        prog.CreatePackage(pkg, nil, nil, true)
    }

    // Build SSA for the package being analyzed
    ssaPkg := prog.CreatePackage(pass.Pkg, pass.Files, pass.TypesInfo, false)

    // Build the call graph using the Class Hierarchy Analysis algorithm
    entryPoints := make(map[*ssa.Function]bool)
    for _, m := range ssaPkg.Members {
        if fn, ok := m.(*ssa.Function); ok {
            entryPoints[fn] = true
        }
    }

    return cha.CallGraph(prog)
}
```

2. Traverse the call graph to identify reachability from production code:

```go
// isReachableFromProduction determines if a function is reachable from production code
func isReachableFromProduction(cg *callgraph.Graph, fn *ssa.Function, pass *analysis.Pass, visitedNodes map[*callgraph.Node]bool) bool {
    if visitedNodes == nil {
        visitedNodes = make(map[*callgraph.Node]bool)
    }

    node := cg.Nodes[fn]
    if node == nil {
        return false
    }

    if visitedNodes[node] {
        return false // Avoid cycles
    }
    visitedNodes[node] = true

    // Check direct callers of this function
    for _, edge := range node.In {
        caller := edge.Caller
        if caller.Func == nil {
            continue
        }

        // If the caller is in a non-test file, this function is reachable from production
        pos := pass.Fset.Position(caller.Func.Pos())
        if !isTestFile(pos.Filename) {
            return true
        }

        // Recursively check if the caller is reachable from production
        if isReachableFromProduction(cg, caller.Func, pass, visitedNodes) {
            return true
        }
    }

    return false
}
```

### Interface Implementation Detection

1. Use types.Implements to detect when a type implements an interface used in production:

```go
// isMethodImplementingProductionInterface checks if a method implements an interface used in production
func isMethodImplementingProductionInterface(pass *analysis.Pass, funcDecl *ast.FuncDecl) bool {
    if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
        return false // Not a method
    }

    // Get the receiver type
    var receiverType types.Type
    recvExpr := funcDecl.Recv.List[0].Type
    recvInfo, ok := pass.TypesInfo.Types[recvExpr]
    if !ok {
        return false
    }
    receiverType = recvInfo.Type

    // Get the method set for this type
    methodName := funcDecl.Name.Name
    methodSet := types.NewMethodSet(receiverType)

    // For each interface in the program
    for _, pkg := range pass.Pkg.Imports() {
        for _, name := range pkg.Scope().Names() {
            obj := pkg.Scope().Lookup(name)
            if obj == nil {
                continue
            }

            // Check if this is an interface type
            iface, ok := obj.Type().Underlying().(*types.Interface)
            if !ok {
                continue
            }

            // Check if receiver type implements this interface
            if types.Implements(receiverType, iface) {
                // Check if this interface is used in production code
                if isInterfaceUsedInProduction(pass, obj) {
                    return true
                }
            }
        }
    }

    return false
}

// isInterfaceUsedInProduction checks if an interface is used in non-test files
func isInterfaceUsedInProduction(pass *analysis.Pass, ifaceObj types.Object) bool {
    // Implementation to check for uses of this interface in non-test files
    // This would need to traverse the AST looking for references to the interface
    // ...
}
```

### Enhanced Cross-Package Reference Analysis

1. Build a more comprehensive package dependency graph:

```go
// PackageDependencyGraph represents dependencies between packages
type PackageDependencyGraph struct {
    // Map from package path to direct dependencies
    Dependencies map[string]map[string]bool

    // Map from package path to imported objects
    ImportedObjects map[string]map[string]types.Object
}

// buildPackageDependencyGraph constructs the package dependency graph
func buildPackageDependencyGraph(pass *analysis.Pass) *PackageDependencyGraph {
    graph := &PackageDependencyGraph{
        Dependencies: make(map[string]map[string]bool),
        ImportedObjects: make(map[string]map[string]types.Object),
    }

    pkg := pass.Pkg
    pkgPath := pkg.Path()

    // Initialize maps for this package
    graph.Dependencies[pkgPath] = make(map[string]bool)
    graph.ImportedObjects[pkgPath] = make(map[string]types.Object)

    // Add direct dependencies
    for _, imp := range pkg.Imports() {
        impPath := imp.Path()
        graph.Dependencies[pkgPath][impPath] = true

        // Record imported objects
        if imp.Scope() != nil {
            for _, name := range imp.Scope().Names() {
                obj := imp.Scope().Lookup(name)
                if obj != nil {
                    fullName := impPath + "." + name
                    graph.ImportedObjects[pkgPath][fullName] = obj
                }
            }
        }
    }

    return graph
}
```

2. Use the dependency graph to track cross-package references:

```go
// isUsedAcrossPackages determines if an identifier is used in other packages
func isUsedAcrossPackages(obj types.Object, graph *PackageDependencyGraph) bool {
    pkg := obj.Pkg()
    if pkg == nil {
        return false
    }

    pkgPath := pkg.Path()
    objName := obj.Name()
    fullName := pkgPath + "." + objName

    // Check if the object is exported (can be used by other packages)
    if !obj.Exported() {
        return false
    }

    // Look for packages that import this package
    for depPkg, deps := range graph.Dependencies {
        if depPkg == pkgPath {
            continue // Skip self
        }

        if deps[pkgPath] {
            // This package imports our package
            // Check if the specific object is used
            if _, ok := graph.ImportedObjects[depPkg][fullName]; ok {
                return true
            }
        }
    }

    return false
}
```

### Exported Identifier Context

1. Add special handling for exported identifiers:

```go
// shouldConsiderExportedStatus determines if an exported identifier should be exempt from "test-only" reporting
func shouldConsiderExportedStatus(name string, config *Config) bool {
    if !config.ConsiderExportedStatus {
        return false
    }

    // Check if the identifier is exported (starts with uppercase)
    return ast.IsExported(name)
}

// In the issue generation function:
for name, info := range result.Declarations {
    // Other checks...

    // If the identifier is exported and configured to respect export status,
    // consider its context before reporting
    if shouldConsiderExportedStatus(name, config) {
        // Check if it's part of the public API or intentionally exported
        if isPartOfPublicAPI(name, result) {
            continue // Skip reporting
        }
    }

    // Proceed with reporting if it's test-only
    // ...
}
```

## Consequences

### Positive
- Dramatically improved accuracy of test-only identifier detection
- Fewer false positives, especially for interface implementations
- Better understanding of complex dependency relationships
- Higher user trust in linter results
- Proper handling of exported identifiers and public API functions

### Negative
- Increased complexity in the analysis process
- Potential performance impact from building and traversing call graphs
- Need for more sophisticated type analysis
- Higher memory usage for storing dependency graphs

### Mitigations
- Implement efficient data structures for graph representation
- Use lazy loading where possible to reduce memory footprint
- Add caching mechanisms for previously analyzed packages
- Provide configuration options to control analysis depth
- Include performance benchmarks to optimize critical paths

## References
- Go callgraph package: golang.org/x/tools/go/callgraph
- Go SSA package: golang.org/x/tools/go/ssa
- Type analysis in Go: golang.org/x/tools/go/types
- Interface implementation detection: https://golang.org/pkg/go/types/#Implements
- Export rules in Go: https://golang.org/ref/spec#Exported_identifiers