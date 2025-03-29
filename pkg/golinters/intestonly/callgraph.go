// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
// This file implements call graph analysis for more accurate dependency tracking.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/callgraph/cha" // external callgraph analysis
	"golang.org/x/tools/go/ssa"           // for SSA building
)

// buildCallGraph constructs a call graph for the package to track function dependencies
func buildCallGraph(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Building call graph using SSA and external callgraph tool...")
	}
	// Determine builder mode for SSA
	builderMode := ssa.BuilderMode(0)
	if config.Debug {
		builderMode = ssa.GlobalDebug
	}
	// Создаём SSA-программу из файлов для данного анализа.
	prog := ssa.NewProgram(pass.Fset, builderMode)
	// Создаём SSA-пакет для текущего пакета.
	_ = prog.CreatePackage(pass.Pkg, pass.Files, pass.TypesInfo, true)
	// Создаем SSA-пакеты для всех импортов текущего пакета, чтобы удовлетворить зависимости (например, "testing").
	for _, impPkg := range pass.Pkg.Imports() {
		prog.CreatePackage(impPkg, nil, nil, true)
	}

	// Отлавливаем возможную панику при сборке SSA-программы.
	defer func() {
		if r := recover(); r != nil {
			if config.Debug {
				fmt.Printf("Recovered from SSA build panic: %v\n", r)
			}
			// При падении останется пустой call graph.
		}
	}()

	// Вместо pkgSSA.Build() вызываем сборку всей программы,
	// чтобы удовлетворить все импортируемые пакеты (например, "os", "testing").
	prog.Build()

	// Build call graph using the CHA algorithm from the external call graph tool.
	cg := cha.CallGraph(prog)

	// Initialize call graph maps if not already done
	if result.CallGraph == nil {
		result.CallGraph = make(map[string][]string)
	}
	if result.CalledBy == nil {
		result.CalledBy = make(map[string][]string)
	}

	// Populate our call graph maps based on the external call graph.
	for _, node := range cg.Nodes {
		// Only consider functions declared in the current package.
		if node.Func == nil || node.Func.Pkg == nil || node.Func.Pkg.Pkg != pass.Pkg {
			continue
		}

		callerName := node.Func.Name()
		for _, edge := range node.Out {
			if edge.Callee == nil || edge.Callee.Func == nil || edge.Callee.Func.Pkg == nil {
				continue
			}
			if edge.Callee.Func.Pkg.Pkg != pass.Pkg {
				continue
			}
			calleeName := edge.Callee.Func.Name()
			result.CallGraph[callerName] = append(result.CallGraph[callerName], calleeName)
			result.CalledBy[calleeName] = append(result.CalledBy[calleeName], callerName)
		}
	}

	// Propagate test usage information through the new call graph.
	propagateCallDependencies(result, config)
}

// analyzeFunctionCalls processes function calls in a file
func analyzeFunctionCalls(file *ast.File, result *AnalysisResult, config *Config) {
	var currentFunc string

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Track the current function we're in
			if node.Recv != nil {
				// This is a method
				if len(node.Recv.List) > 0 {
					if t, ok := node.Recv.List[0].Type.(*ast.StarExpr); ok {
						if ident, ok := t.X.(*ast.Ident); ok {
							currentFunc = ident.Name + "." + node.Name.Name
						}
					} else if ident, ok := node.Recv.List[0].Type.(*ast.Ident); ok {
						currentFunc = ident.Name + "." + node.Name.Name
					}
				}
			} else {
				currentFunc = node.Name.Name
			}

		case *ast.CallExpr:
			if currentFunc == "" {
				return true
			}

			// Extract called function name
			var calledFunc string
			switch fun := node.Fun.(type) {
			case *ast.Ident:
				calledFunc = fun.Name

				// Check if this is a declaration we're tracking
				if _, isDeclared := result.Declarations[calledFunc]; isDeclared {
					// Record in call graph
					result.CallGraph[currentFunc] = append(result.CallGraph[currentFunc], calledFunc)
					result.CalledBy[calledFunc] = append(result.CalledBy[calledFunc], currentFunc)
				}

			case *ast.SelectorExpr:
				if x, ok := fun.X.(*ast.Ident); ok {
					// Check if this is a method call on a known type
					if _, isDeclared := result.Declarations[x.Name]; isDeclared {
						calledFunc = x.Name + "." + fun.Sel.Name

						// Record in call graph
						result.CallGraph[currentFunc] = append(result.CallGraph[currentFunc], calledFunc)
						result.CalledBy[calledFunc] = append(result.CalledBy[calledFunc], currentFunc)
					}

					// Check if this might be a package-qualified function call
					// Try to find a matching function with the package qualifier
					for declName, declInfo := range result.Declarations {
						if declInfo.DeclType == DeclFunction && declName == fun.Sel.Name {
							if _, ok := result.ImportedPkgs[x.Name]; ok {
								if strings.HasSuffix(declInfo.ImportRef, "."+fun.Sel.Name) {
									calledFunc = declName

									// Record in call graph
									result.CallGraph[currentFunc] = append(result.CallGraph[currentFunc], calledFunc)
									result.CalledBy[calledFunc] = append(result.CalledBy[calledFunc], currentFunc)
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

// propagateCallDependencies propagates test usage information through the call graph
func propagateCallDependencies(result *AnalysisResult, config *Config) {
	// For each function in the non-test files
	for declName, declInfo := range result.Declarations {
		// Skip if this is not a function or method
		if declInfo.DeclType != DeclFunction && declInfo.DeclType != DeclMethod {
			continue
		}

		// If this function is used in test files
		if _, usedInTests := result.TestUsages[declName]; usedInTests {
			// Check if it's also used in non-test files
			if _, usedInProd := result.Usages[declName]; !usedInProd {
				// Propagate test usage to all functions called by this function
				if config.Debug {
					fmt.Printf("Propagating test usage from %s\n", declName)
				}
				propagateTestUsage(declName, result, make(map[string]bool), config)
			}
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
			declInfo, isDeclared := result.Declarations[callee]
			if isDeclared {
				// Only propagate to functions and methods
				if declInfo.DeclType != DeclFunction && declInfo.DeclType != DeclMethod {
					continue
				}

				// Add a fake position to indicate it's called indirectly
				if _, alreadyMarked := result.TestUsages[callee]; !alreadyMarked {
					if config.Debug {
						fmt.Printf("Propagating test usage from %s to %s\n", rootFunc, callee)
					}

					// Mark as used in tests by adding a synthetic position
					usage := UsageInfo{
						Pos:      token.NoPos,
						FilePath: "",
						IsTest:   true,
					}
					result.TestUsages[callee] = append(result.TestUsages[callee], usage)

					// Continue propagation
					propagateTestUsage(callee, result, visited, config)
				}
			}
		}
	}
}

// isFunctionName checks if the given name represents a function or method
func isFunctionName(name string, result *AnalysisResult) bool {
	if info, exists := result.Declarations[name]; exists {
		return info.DeclType == DeclFunction || info.DeclType == DeclMethod
	}
	return false
}
