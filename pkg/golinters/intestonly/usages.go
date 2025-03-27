package intestonly

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// analyzeUsages examines all files to track where declarations are used
func analyzeUsages(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) && !isTest {
			continue
		}

		// Process the file for usages
		processFileUsages(file, isTest, pass, result, config)
	}
}

// processFileUsages analyzes a file for identifier usages
func processFileUsages(file *ast.File, isTest bool, pass *analysis.Pass, result *AnalysisResult, config *Config) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.Ident:
			processIdentUsage(n, isTest, result, config)
		case *ast.SelectorExpr:
			processSelectorUsage(n, isTest, result, config)
		}
		return true
	})
}

// processIdentUsage processes a direct identifier usage
func processIdentUsage(n *ast.Ident, isTest bool, result *AnalysisResult, config *Config) {
	if n == nil || n.Name == "" {
		return
	}

	// Skip if this is a declaration position
	if _, isDeclPos := result.DeclPositions[n.Pos()]; isDeclPos {
		return
	}

	// Record usage
	if _, isDeclared := result.Declarations[n.Name]; isDeclared {
		if isTest {
			result.TestUsages[n.Name] = append(result.TestUsages[n.Name], n.Pos())
			if config.Debug {
				fmt.Printf("Test usage of %s\n", n.Name)
			}
		} else {
			result.NonTestUsages[n.Name] = append(result.NonTestUsages[n.Name], n.Pos())
			if config.Debug {
				fmt.Printf("Non-test usage of %s\n", n.Name)
			}
		}
	}
}

// processSelectorUsage processes a selector expression (x.y)
func processSelectorUsage(n *ast.SelectorExpr, isTest bool, result *AnalysisResult, config *Config) {
	if n == nil || n.X == nil || n.Sel == nil || n.Sel.Name == "" {
		return
	}

	if x, ok := n.X.(*ast.Ident); ok && x != nil {
		// Check if x is an imported package
		if importPath, ok := result.ImportedPkgs[x.Name]; ok {
			// Construct full reference: import_path.selector_name
			fullRef := importPath + "." + n.Sel.Name

			// Check if this matches any declaration
			for _, info := range result.Declarations {
				if info.ImportRef == fullRef {
					if isTest {
						result.TestUsages[info.Name] = append(result.TestUsages[info.Name], n.Sel.Pos())
						if config.Debug {
							fmt.Printf("Test usage of imported %s via %s\n", info.Name, fullRef)
						}
					} else {
						result.NonTestUsages[info.Name] = append(result.NonTestUsages[info.Name], n.Sel.Pos())
						if config.Debug {
							fmt.Printf("Non-test usage of imported %s via %s\n", info.Name, fullRef)
						}
					}
				}
			}
		}

		// Check if the selector (method name) is a known declaration
		if _, isDeclared := result.Declarations[n.Sel.Name]; isDeclared {
			if isTest {
				result.TestUsages[n.Sel.Name] = append(result.TestUsages[n.Sel.Name], n.Sel.Pos())
				if config.Debug {
					fmt.Printf("Test usage of method %s\n", n.Sel.Name)
				}
			} else {
				result.NonTestUsages[n.Sel.Name] = append(result.NonTestUsages[n.Sel.Name], n.Sel.Pos())
				if config.Debug {
					fmt.Printf("Non-test usage of method %s\n", n.Sel.Name)
				}
			}
		}

		// Also check if the base type is a known declaration
		if _, isDeclared := result.Declarations[x.Name]; isDeclared {
			if isTest {
				result.TestUsages[x.Name] = append(result.TestUsages[x.Name], x.Pos())
			} else {
				result.NonTestUsages[x.Name] = append(result.NonTestUsages[x.Name], x.Pos())
			}
		}
	}
}
