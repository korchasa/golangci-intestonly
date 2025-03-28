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
2. Traverse the call graph to identify reachability from production code:

### Interface Implementation Detection

1. Use types.Implements to detect when a type implements an interface used in production:

### Enhanced Cross-Package Reference Analysis

1. Build a more comprehensive package dependency graph
2. Use the dependency graph to track cross-package references

### Exported Identifier Context

1. Add special handling for exported identifiers

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