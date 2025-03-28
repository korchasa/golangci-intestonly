// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
// This file implements call graph analysis for more accurate dependency tracking.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// buildCallGraph constructs a call graph for the package to track function dependencies
func buildCallGraph(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Building call graph...")
	}

	// Process each file in the package
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName, config)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) && !isTest {
			continue
		}

		// Use our AST visitor to find function calls
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			// Find function declarations to establish caller context
			case *ast.FuncDecl:
				if node.Name == nil {
					return true
				}

				// Get the current function name
				currentFuncName := node.Name.Name

				// For methods, include the receiver type in the name
				if node.Recv != nil && len(node.Recv.List) > 0 {
					// Get the receiver type name
					var recvTypeName string

					// Handle different receiver type expressions
					switch recvType := node.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						// Pointer receiver (e.g., func (t *Type) Method())
						if ident, ok := recvType.X.(*ast.Ident); ok {
							recvTypeName = ident.Name
						}
					case *ast.Ident:
						// Value receiver (e.g., func (t Type) Method())
						recvTypeName = recvType.Name
					}

					if recvTypeName != "" {
						currentFuncName = recvTypeName + "." + currentFuncName
					}
				}

				// Process function body to find calls
				if node.Body != nil {
					findFunctionCalls(pass, node.Body, currentFuncName, result, isTest, config)
				}

			// Process initializations at package level that might contain function calls
			case *ast.ValueSpec:
				if node.Values != nil {
					for _, expr := range node.Values {
						processFuncCallExpr(pass, expr, "package_init", result, isTest, config)
					}
				}
			}

			return true
		})
	}

	// After building the direct call graph, propagate dependencies
	propagateCallDependencies(result, config)
}

// findFunctionCalls analyzes a block of statements to find function calls
func findFunctionCalls(pass *analysis.Pass, block *ast.BlockStmt, callerName string, result *AnalysisResult, isTest bool, config *Config) {
	ast.Inspect(block, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			processFuncCallExpr(pass, node, callerName, result, isTest, config)
		}
		return true
	})
}

// processFuncCallExpr processes a function call expression and updates the call graph
func processFuncCallExpr(pass *analysis.Pass, expr ast.Expr, callerName string, result *AnalysisResult, isTest bool, config *Config) {
	switch callNode := expr.(type) {
	case *ast.CallExpr:
		// Get the called function name
		var calleeName string

		switch fun := callNode.Fun.(type) {
		case *ast.Ident:
			// Simple function call (e.g., foo())
			calleeName = fun.Name

		case *ast.SelectorExpr:
			// Method call or qualified function call (e.g., x.Method() or pkg.Func())
			if sel, ok := fun.X.(*ast.Ident); ok {
				// Handle package-qualified calls or method calls
				calleeName = sel.Name + "." + fun.Sel.Name
			}
		}

		if calleeName != "" {
			// Add to the call graph
			if _, exists := result.CallGraph[callerName]; !exists {
				result.CallGraph[callerName] = []string{}
			}

			// Add callee to the list of functions called by caller (if not already there)
			found := false
			for _, existing := range result.CallGraph[callerName] {
				if existing == calleeName {
					found = true
					break
				}
			}

			if !found {
				result.CallGraph[callerName] = append(result.CallGraph[callerName], calleeName)

				// Also update the reverse mapping (CalledBy)
				if _, exists := result.CalledBy[calleeName]; !exists {
					result.CalledBy[calleeName] = []string{}
				}

				// Add caller to the list of functions that call callee
				callerFound := false
				for _, existingCaller := range result.CalledBy[calleeName] {
					if existingCaller == callerName {
						callerFound = true
						break
					}
				}

				if !callerFound {
					result.CalledBy[calleeName] = append(result.CalledBy[calleeName], callerName)
				}

				if config.Debug {
					fmt.Printf("Call relationship: %s -> %s\n", callerName, calleeName)
				}
			}
		}

		// Recursively process function arguments that might be function calls
		for _, arg := range callNode.Args {
			processFuncCallExpr(pass, arg, callerName, result, isTest, config)
		}
	}
}

// propagateCallDependencies propagates transitive dependencies through the call graph
func propagateCallDependencies(result *AnalysisResult, config *Config) {
	// For each function in the non-test files
	for declName := range result.Declarations {
		// Skip if this is not a function or method
		if !isFunctionName(declName, result) {
			continue
		}

		// If this function is used in test files
		if _, usedInTests := result.TestUsages[declName]; usedInTests {
			// Propagate test usage to all functions called by this function
			propagateTestUsage(declName, result, make(map[string]bool), config)
		}
	}
}

// propagateTestUsage marks all functions in the call graph reachable from rootFunc as used in tests
func propagateTestUsage(rootFunc string, result *AnalysisResult, visited map[string]bool, config *Config) {
	// Avoid cycles in the call graph
	if visited[rootFunc] {
		return
	}
	visited[rootFunc] = true

	// Mark all functions called by rootFunc as used by tests
	if callees, exists := result.CallGraph[rootFunc]; exists {
		for _, callee := range callees {
			// If this callee is a declaration in our package
			if _, isDeclared := result.Declarations[callee]; isDeclared {
				// Add a fake position to indicate it's called indirectly
				if _, alreadyMarked := result.TestUsages[callee]; !alreadyMarked {
					if config.Debug {
						fmt.Printf("Propagating test usage from %s to %s\n", rootFunc, callee)
					}

					// Mark as used in tests by adding a synthetic position
					result.TestUsages[callee] = append(result.TestUsages[callee], token.NoPos)

					// Continue propagation
					propagateTestUsage(callee, result, visited, config)
				}
			}
		}
	}
}

// isFunctionName checks if the given name represents a function or method
func isFunctionName(name string, result *AnalysisResult) bool {
	// Simple heuristic: if it appears in CallGraph or CalledBy, it's a function
	if _, inCallGraph := result.CallGraph[name]; inCallGraph {
		return true
	}
	if _, inCalledBy := result.CalledBy[name]; inCalledBy {
		return true
	}

	// Or if it's a known declaration with certain characteristics
	if info, isDeclared := result.Declarations[name]; isDeclared {
		return info.IsMethod || func() bool {
			// Check if it appears as a method in MethodsOfType
			for _, methods := range result.MethodsOfType {
				for _, method := range methods {
					if method == name {
						return true
					}
				}
			}
			return false
		}()
	}

	return false
}
