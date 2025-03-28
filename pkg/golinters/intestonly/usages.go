// Package intestonly provides advanced identifier usage detection.
// This file implements a unified system for analyzing identifier usage across test and non-test files.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzeUsages examines all files to track where declarations are used
// This unified system combines basic and advanced analysis methods
func analyzeUsages(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Process each file
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) && !isTest {
			continue
		}

		// Create a set of all identifiers in this file's scope to detect shadowing
		fileScope := make(map[string]bool)

		// First pass: collect all local declarations to handle shadowing
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Name != nil {
					fileScope[node.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range node.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							fileScope[name.Name] = true
						}
					} else if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name != nil {
							fileScope[typeSpec.Name.Name] = true
						}
					}
				}
			case *ast.AssignStmt:
				if node.Tok == token.DEFINE {
					for _, lhs := range node.Lhs {
						if ident, ok := lhs.(*ast.Ident); ok {
							fileScope[ident.Name] = true
						}
					}
				}
			}
			return true
		})

		// Second pass: analyze actual usages
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.Ident:
				// Skip declarations and local shadowed identifiers
				if fileScope[node.Name] {
					return true
				}

				// Check if this is a reference to a declaration we're tracking
				if _, exists := result.Declarations[node.Name]; !exists {
					return true
				}

				// Record usage based on whether this is a test file
				if isTest {
					result.TestUsages[node.Name] = append(result.TestUsages[node.Name], node.Pos())
					if config.Debug {
						fmt.Printf("Test usage of %s\n", node.Name)
					}
				} else {
					result.NonTestUsages[node.Name] = append(result.NonTestUsages[node.Name], node.Pos())
					if config.Debug {
						fmt.Printf("Non-test usage of %s\n", node.Name)
					}
				}

			case *ast.SelectorExpr:
				// Handle qualified references (pkg.Func or x.Method)
				if x, ok := node.X.(*ast.Ident); ok {
					sel := node.Sel

					// Check if this is a package-qualified reference
					if importPath, ok := result.ImportedPkgs[x.Name]; ok {
						// Package-qualified reference
						fullName := importPath + "." + sel.Name

						// Check if this matches one of our tracked declarations
						for declName, info := range result.Declarations {
							if info.ImportRef == fullName {
								if isTest {
									result.TestUsages[declName] = append(result.TestUsages[declName], sel.Pos())
									if config.Debug {
										fmt.Printf("Test usage of imported %s via %s\n", declName, fullName)
									}
								} else {
									result.NonTestUsages[declName] = append(result.NonTestUsages[declName], sel.Pos())
									if config.Debug {
										fmt.Printf("Non-test usage of imported %s via %s\n", declName, fullName)
									}
								}
							}
						}
					}

					// Also check if the selector (method name) is a known declaration
					if _, isDeclared := result.Declarations[sel.Name]; isDeclared {
						if isTest {
							result.TestUsages[sel.Name] = append(result.TestUsages[sel.Name], sel.Pos())
							if config.Debug {
								fmt.Printf("Test usage of method %s\n", sel.Name)
							}
						} else {
							result.NonTestUsages[sel.Name] = append(result.NonTestUsages[sel.Name], sel.Pos())
							if config.Debug {
								fmt.Printf("Non-test usage of method %s\n", sel.Name)
							}
						}
					}

					// Also check if the base type is a known declaration
					if _, isDeclared := result.Declarations[x.Name]; isDeclared {
						if isTest {
							result.TestUsages[x.Name] = append(result.TestUsages[x.Name], x.Pos())
							if config.Debug {
								fmt.Printf("Test usage of base type %s\n", x.Name)
							}
						} else {
							result.NonTestUsages[x.Name] = append(result.NonTestUsages[x.Name], x.Pos())
							if config.Debug {
								fmt.Printf("Non-test usage of base type %s\n", x.Name)
							}
						}
					}
				}
			}

			return true
		})

		// Perform additional type embedding analysis on this file
		if config.EnableTypeEmbeddingAnalysis {
			analyzeTypeEmbeddingForFile(file, pass.Fset, isTest, result, config)
		}
	}

	// Perform reflection-based usage analysis if enabled
	if config.EnableReflectionAnalysis {
		analyzeReflectionUsage(pass, result, config)
	}

	// Analyze registry and plugin patterns if enabled
	if config.EnableRegistryPatternDetection {
		analyzeRegistryPatterns(pass, result, config)
	}
}

// analyzeTypeEmbeddingForFile analyzes a single file for type embedding
func analyzeTypeEmbeddingForFile(file *ast.File, fset *token.FileSet, isTest bool, result *AnalysisResult, config *Config) {
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for type definitions with embedded fields
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Name == nil {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok || structType.Fields == nil {
			return true
		}

		// Find embedded fields
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 { // Embedded field
				// Handle embedded type usages
				switch fieldType := field.Type.(type) {
				case *ast.Ident:
					// Direct embedding of a named type
					embeddedTypeName := fieldType.Name
					if _, isDeclared := result.Declarations[embeddedTypeName]; isDeclared {
						if isTest {
							result.TestUsages[embeddedTypeName] = append(result.TestUsages[embeddedTypeName], fieldType.Pos())
							if config.Debug {
								fmt.Printf("Test usage of embedded type %s\n", embeddedTypeName)
							}
						} else {
							result.NonTestUsages[embeddedTypeName] = append(result.NonTestUsages[embeddedTypeName], fieldType.Pos())
							if config.Debug {
								fmt.Printf("Non-test usage of embedded type %s\n", embeddedTypeName)
							}
						}
					}
				case *ast.SelectorExpr:
					// Embedding from another package
					if x, ok := fieldType.X.(*ast.Ident); ok && x != nil {
						if importPath, ok := result.ImportedPkgs[x.Name]; ok {
							fullName := importPath + "." + fieldType.Sel.Name
							for declName, info := range result.Declarations {
								if info.ImportRef == fullName {
									if isTest {
										result.TestUsages[declName] = append(result.TestUsages[declName], fieldType.Sel.Pos())
										if config.Debug {
											fmt.Printf("Test usage of embedded imported type %s\n", declName)
										}
									} else {
										result.NonTestUsages[declName] = append(result.NonTestUsages[declName], fieldType.Sel.Pos())
										if config.Debug {
											fmt.Printf("Non-test usage of embedded imported type %s\n", declName)
										}
									}
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

// analyzeTypeEmbedding tracks usage through type embedding
// This function is kept for backwards compatibility
func analyzeTypeEmbedding(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Process each file
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		// Skip files that should be ignored
		if shouldIgnoreFile(fileName, config) && !isTest {
			continue
		}

		analyzeTypeEmbeddingForFile(file, pass.Fset, isTest, result, config)
	}
}

// analyzeReflectionUsage detects usage through reflection
func analyzeReflectionUsage(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		// Skip if we're not checking test files and this is a test file
		if !isTest && shouldIgnoreFile(fileName, config) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Look for reflection-based usage patterns
			// Such as reflect.TypeOf(), reflect.ValueOf() followed by .Method() or .FieldByName()

			switch node := n.(type) {
			case *ast.CallExpr:
				// Check for reflect.TypeOf() or reflect.ValueOf() calls
				if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
					if x, ok := sel.X.(*ast.Ident); ok && x.Name == "reflect" {
						if sel.Sel.Name == "TypeOf" || sel.Sel.Name == "ValueOf" || sel.Sel.Name == "New" {
							// Look for string literals in method or field access
							// that match our declarations
							for _, arg := range node.Args {
								// If the argument is a reference to a type or variable we're tracking,
								// mark it as used
								if ident, ok := arg.(*ast.Ident); ok {
									for declName := range result.Declarations {
										if declName == ident.Name {
											if isTest {
												result.TestUsages[declName] = append(result.TestUsages[declName], ident.Pos())
												if config.Debug {
													fmt.Printf("Test reflection usage of %s\n", declName)
												}
											} else {
												result.NonTestUsages[declName] = append(result.NonTestUsages[declName], ident.Pos())
												if config.Debug {
													fmt.Printf("Non-test reflection usage of %s\n", declName)
												}
											}
										}
									}
								}
							}

							// Mark all identifiers in the parent expression as potentially used
							// This covers cases like reflect.TypeOf(x).Method(0).Func.Call(...)
							parent := findParentNode(file, node)
							if parent != nil {
								ast.Inspect(parent, func(m ast.Node) bool {
									if ident, ok := m.(*ast.Ident); ok {
										if decl, exists := result.Declarations[ident.Name]; exists {
											if config.Debug {
												fmt.Printf("Found potential reflection parent usage of %s\n", ident.Name)
											}

											// Record this as a usage
											if isTest {
												result.TestUsages[decl.Name] = append(result.TestUsages[decl.Name], ident.Pos())
											} else {
												result.NonTestUsages[decl.Name] = append(result.NonTestUsages[decl.Name], ident.Pos())
											}
										}
									}
									return true
								})
							}
						}
					}
				}

			// Also check for common reflection method calls like MethodByName, FieldByName
			case *ast.SelectorExpr:
				if node.Sel.Name == "MethodByName" || node.Sel.Name == "FieldByName" {
					// Process parent call expression to find string arguments
					parent := findParentNode(file, node)
					if parentCall, ok := parent.(*ast.CallExpr); ok {
						for _, arg := range parentCall.Args {
							// Check for string literals that might reference our declarations
							if lit, ok := arg.(*ast.BasicLit); ok {
								literal := strings.Trim(lit.Value, "\"")
								// Check if this method or field name matches a declaration
								for declName := range result.Declarations {
									if declName == literal {
										if isTest {
											result.TestUsages[declName] = append(result.TestUsages[declName], lit.Pos())
											if config.Debug {
												fmt.Printf("Test reflection method/field access of %s\n", declName)
											}
										} else {
											result.NonTestUsages[declName] = append(result.NonTestUsages[declName], lit.Pos())
											if config.Debug {
												fmt.Printf("Non-test reflection method/field access of %s\n", declName)
											}
										}
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
}

// analyzeRegistryPatterns detects usage in registry and plugin patterns
func analyzeRegistryPatterns(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		ast.Inspect(file, func(n ast.Node) bool {
			// Look for common registration patterns
			// 1. init() functions with side effects
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				if funcDecl.Name != nil && funcDecl.Name.Name == "init" && funcDecl.Recv == nil {
					// Assume all declarations used in init() are used by the package
					ast.Inspect(funcDecl, func(m ast.Node) bool {
						if ident, ok := m.(*ast.Ident); ok {
							for declName := range result.Declarations {
								if declName == ident.Name {
									if isTest {
										result.TestUsages[declName] = append(result.TestUsages[declName], ident.Pos())
										if config.Debug {
											fmt.Printf("Test init() usage of %s\n", declName)
										}
									} else {
										result.NonTestUsages[declName] = append(result.NonTestUsages[declName], ident.Pos())
										if config.Debug {
											fmt.Printf("Non-test init() usage of %s\n", declName)
										}
									}
								}
							}
						}
						return true
					})
				}
			}

			// 2. Map assignments that may be registries
			if assign, ok := n.(*ast.AssignStmt); ok {
				for _, lhs := range assign.Lhs {
					// Check if this is a map assignment (e.g., registry[key] = value)
					if _, ok := lhs.(*ast.IndexExpr); ok {
						// This is a map or array assignment
						// Find identifiers in the right side that might be registering
						for _, rhs := range assign.Rhs {
							ast.Inspect(rhs, func(m ast.Node) bool {
								if ident, ok := m.(*ast.Ident); ok {
									if _, isDeclared := result.Declarations[ident.Name]; isDeclared {
										if isTest {
											result.TestUsages[ident.Name] = append(result.TestUsages[ident.Name], ident.Pos())
											if config.Debug {
												fmt.Printf("Test registry usage of %s\n", ident.Name)
											}
										} else {
											result.NonTestUsages[ident.Name] = append(result.NonTestUsages[ident.Name], ident.Pos())
											if config.Debug {
												fmt.Printf("Non-test registry usage of %s\n", ident.Name)
											}
										}
									}
								}
								return true
							})
						}
					}
				}
			}

			return true
		})
	}
}

// findParentNode finds a parent node that contains the target node
func findParentNode(file *ast.File, target ast.Node) ast.Node {
	var parent ast.Node

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch expr := n.(type) {
		case *ast.SelectorExpr:
			if expr.X == target {
				parent = expr
				return false
			}
		case *ast.CallExpr:
			if expr.Fun == target {
				parent = expr
				return false
			}
		}

		return true
	})

	return parent
}

// Helper function to add a test usage
func (result *AnalysisResult) AddTestUsage(name string) {
	result.TestUsages[name] = append(result.TestUsages[name], token.NoPos)
}

// Helper function to add a non-test usage
func (result *AnalysisResult) AddNonTestUsage(name string) {
	result.NonTestUsages[name] = append(result.NonTestUsages[name], token.NoPos)
}
