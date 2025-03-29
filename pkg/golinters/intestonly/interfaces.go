// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
// This file implements enhanced interface implementation detection.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// analyzeInterfaceImplementations detects interface implementations throughout the codebase
func analyzeInterfaceImplementations(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	if config.Debug {
		fmt.Println("Analyzing interface implementations...")
	}

	// First pass: collect all interfaces and their methods
	collectInterfaces(pass, result, config)

	// Second pass: find types that implement these interfaces
	findImplementors(pass, result, config)

	// Third pass: propagate usage information
	propagateInterfaceUsages(result, config)
}

// collectInterfaces finds all interface declarations in the package
func collectInterfaces(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		// Skip test files when collecting interface declarations
		fileName := pass.Fset.File(file.Pos()).Name()
		if isTestFile(fileName, config) {
			continue
		}

		// Find interface declarations
		ast.Inspect(file, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil {
				return true
			}

			// Only interested in interface types
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}

			interfaceName := typeSpec.Name.Name
			result.Interfaces[interfaceName] = []string{}

			// No methods in this interface
			if interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
				return true
			}

			// Extract method names from the interface
			for _, method := range interfaceType.Methods.List {
				// Skip embedded interfaces for now
				if len(method.Names) > 0 {
					methodName := method.Names[0].Name
					result.Interfaces[interfaceName] = append(result.Interfaces[interfaceName], methodName)

					if config.Debug {
						fmt.Printf("Interface %s has method %s\n", interfaceName, methodName)
					}
				}
			}

			return true
		})
	}
}

// findImplementors identifies types that implement the collected interfaces
func findImplementors(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Use the types package to identify implementations
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName, config)

		// Find all type declarations
		ast.Inspect(file, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil {
				return true
			}

			typeName := typeSpec.Name.Name

			// Skip interfaces themselves
			if _, isIntf := typeSpec.Type.(*ast.InterfaceType); isIntf {
				return true
			}

			// Get methods for this type
			collectMethodsForType(pass, file, typeName, result, config)

			// Check if this type implements any known interfaces
			for interfaceName, interfaceMethods := range result.Interfaces {
				// Check if this type implements the interface
				if implementsInterface(typeName, interfaceMethods, result) {
					// Add to implementations map
					result.Implementations[interfaceName] = append(
						result.Implementations[interfaceName], typeName)

					if config.Debug {
						fmt.Printf("Type %s implements interface %s\n", typeName, interfaceName)
					}

					// If interface is used in production code, mark implementation as used
					if _, usedInProd := result.Usages[interfaceName]; usedInProd {
						// Mark type as used in production
						if _, exists := result.Usages[typeName]; !exists {
							usage := UsageInfo{
								Pos:      token.NoPos,
								FilePath: "",
								IsTest:   false,
							}
							result.Usages[typeName] = append(result.Usages[typeName], usage)
							if config.Debug {
								fmt.Printf("Marking %s as used in production because it implements %s\n",
									typeName, interfaceName)
							}
						}

						// Also mark its methods as used in production
						for _, method := range interfaceMethods {
							qualifiedMethod := typeName + "." + method

							// Check if this method exists for this type
							if methodExists(qualifiedMethod, result) {
								if _, exists := result.Usages[qualifiedMethod]; !exists {
									usage := UsageInfo{
										Pos:      token.NoPos,
										FilePath: "",
										IsTest:   false,
									}
									result.Usages[qualifiedMethod] = append(result.Usages[qualifiedMethod], usage)
									if config.Debug {
										fmt.Printf("Marking method %s as used in production\n", qualifiedMethod)
									}
								}
							}
						}
					}

					// If type is used in tests, also mark the interface as used in tests
					if isTest {
						if _, exists := result.TestUsages[typeName]; exists {
							if _, marked := result.TestUsages[interfaceName]; !marked {
								usage := UsageInfo{
									Pos:      token.NoPos,
									FilePath: "",
									IsTest:   true,
								}
								result.TestUsages[interfaceName] = append(result.TestUsages[interfaceName], usage)
								if config.Debug {
									fmt.Printf("Marking interface %s as used in tests due to implementation %s\n",
										interfaceName, typeName)
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

// collectMethodsForType finds all methods for a given type
func collectMethodsForType(pass *analysis.Pass, file *ast.File, typeName string, result *AnalysisResult, config *Config) {
	// Initialize method slice if not already present
	if _, exists := result.MethodsOfType[typeName]; !exists {
		result.MethodsOfType[typeName] = []string{}
	}

	// Find method declarations for this type
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		// Check if this method belongs to the type we're looking for
		recvType := funcDecl.Recv.List[0].Type
		var recvTypeName string

		switch recv := recvType.(type) {
		case *ast.Ident:
			// Value receiver
			recvTypeName = recv.Name
		case *ast.StarExpr:
			// Pointer receiver
			if ident, ok := recv.X.(*ast.Ident); ok {
				recvTypeName = ident.Name
			}
		}

		if recvTypeName == typeName {
			methodName := funcDecl.Name.Name

			// Add to methods map if not already present
			found := false
			for _, existingMethod := range result.MethodsOfType[typeName] {
				if existingMethod == methodName {
					found = true
					break
				}
			}

			if !found {
				result.MethodsOfType[typeName] = append(result.MethodsOfType[typeName], methodName)
				if config.Debug {
					fmt.Printf("Type %s has method %s\n", typeName, methodName)
				}
			}
		}
	}
}

// implementsInterface checks if a type implements all methods of an interface
func implementsInterface(typeName string, interfaceMethods []string, result *AnalysisResult) bool {
	typeMethods, exists := result.MethodsOfType[typeName]
	if !exists {
		return false
	}

	// Check if the type has all the methods required by the interface
	for _, interfaceMethod := range interfaceMethods {
		found := false
		for _, typeMethod := range typeMethods {
			if typeMethod == interfaceMethod {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// methodExists checks if a qualified method name exists in the result
func methodExists(qualifiedMethod string, result *AnalysisResult) bool {
	if _, exists := result.Declarations[qualifiedMethod]; exists {
		return true
	}

	for typeName, methods := range result.MethodsOfType {
		for _, method := range methods {
			if typeName+"."+method == qualifiedMethod {
				return true
			}
		}
	}

	return false
}

// propagateInterfaceUsages propagates usage information between interfaces and implementations
func propagateInterfaceUsages(result *AnalysisResult, config *Config) {
	// Process each interface
	for interfaceName, implementations := range result.Implementations {
		// Check if the interface is used via the call graph
		if config.Debug {
			fmt.Printf("Analyzing interface %s with %d implementations\n",
				interfaceName, len(implementations))
		}

		// If the interface is used in tests only
		if _, usedInTests := result.TestUsages[interfaceName]; usedInTests {
			if _, usedInProd := result.Usages[interfaceName]; !usedInProd {
				// For each implementation, check if it's used in production
				for _, implType := range implementations {
					if _, usedInProd := result.Usages[implType]; usedInProd {
						// Mark interface as used in production since its implementation is
						usage := UsageInfo{
							Pos:      token.NoPos,
							FilePath: "",
							IsTest:   false,
						}
						result.Usages[interfaceName] = append(result.Usages[interfaceName], usage)
						if config.Debug {
							fmt.Printf("Marking interface %s as used in production due to implementation %s\n",
								interfaceName, implType)
						}
						break
					}

					// Check if any methods of this implementation are used in production code
					// via the call graph
					if methods, ok := result.Interfaces[interfaceName]; ok {
						for _, method := range methods {
							qualifiedMethod := implType + "." + method
							if methodExists(qualifiedMethod, result) {
								// Check if this method is used in production (either directly or via call graph)
								if _, directlyUsed := result.Usages[qualifiedMethod]; directlyUsed {
									usage := UsageInfo{
										Pos:      token.NoPos,
										FilePath: "",
										IsTest:   false,
									}
									result.Usages[interfaceName] = append(result.Usages[interfaceName], usage)

									if config.Debug {
										fmt.Printf("Marking interface %s as used in production due to method %s\n",
											interfaceName, qualifiedMethod)
									}
									break
								}

								// Check if the method is called by production code via the call graph
								if callers, hasCalls := result.CalledBy[qualifiedMethod]; hasCalls {
									for _, caller := range callers {
										// If the caller is used in production code, mark interface as used
										if _, callerInProd := result.Usages[caller]; callerInProd {
											usage := UsageInfo{
												Pos:      token.NoPos,
												FilePath: "",
												IsTest:   false,
											}
											result.Usages[interfaceName] = append(result.Usages[interfaceName], usage)

											if config.Debug {
												fmt.Printf("Marking interface %s as used in production via call graph (caller: %s)\n",
													interfaceName, caller)
											}
											break
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// If the interface is used in production
		if _, usedInProd := result.Usages[interfaceName]; usedInProd {
			// All implementations must be considered used in production
			for _, implType := range implementations {
				// Mark implementation as used in production
				if _, exists := result.Usages[implType]; !exists {
					usage := UsageInfo{
						Pos:      token.NoPos,
						FilePath: "",
						IsTest:   false,
					}
					result.Usages[implType] = append(result.Usages[implType], usage)
					if config.Debug {
						fmt.Printf("Marking implementation %s as used in production due to interface %s\n",
							implType, interfaceName)
					}
				}

				// Also mark all methods that are part of the interface
				if methods, ok := result.Interfaces[interfaceName]; ok {
					for _, method := range methods {
						qualifiedMethod := implType + "." + method
						if methodExists(qualifiedMethod, result) {
							if _, exists := result.Usages[qualifiedMethod]; !exists {
								usage := UsageInfo{
									Pos:      token.NoPos,
									FilePath: "",
									IsTest:   false,
								}
								result.Usages[qualifiedMethod] = append(result.Usages[qualifiedMethod], usage)
								if config.Debug {
									fmt.Printf("Marking method %s as used in production\n", qualifiedMethod)
								}
							}
						}
					}
				}
			}
		} else {
			// The interface is not directly used in production
			// Check if any methods are being called on a variable of interface type
			// by looking at the call graph
			callsViaInterface := false

			if methods, ok := result.Interfaces[interfaceName]; ok {
				for _, method := range methods {
					// For each method in the interface, check if it's called by production code
					for implType := range result.MethodsOfType {
						qualifiedMethod := implType + "." + method

						// Check if this method exists and is called by production code
						if methodExists(qualifiedMethod, result) {
							if callers, hasCalls := result.CalledBy[qualifiedMethod]; hasCalls {
								for _, caller := range callers {
									// If the caller is used in production, mark as used via call graph
									if _, callerInProd := result.Usages[caller]; callerInProd {
										callsViaInterface = true

										if config.Debug {
											fmt.Printf("Interface %s is used in production via call to %s from %s\n",
												interfaceName, qualifiedMethod, caller)
										}
										break
									}
								}
							}
						}

						if callsViaInterface {
							break
						}
					}

					if callsViaInterface {
						break
					}
				}
			}

			// If we found calls via the interface, mark it and its implementations as used
			if callsViaInterface {
				// Mark interface as used in production
				usage := UsageInfo{
					Pos:      token.NoPos,
					FilePath: "",
					IsTest:   false,
				}
				result.Usages[interfaceName] = append(result.Usages[interfaceName], usage)

				// Mark all implementations as used in production
				for _, implType := range implementations {
					if _, exists := result.Usages[implType]; !exists {
						implUsage := UsageInfo{
							Pos:      token.NoPos,
							FilePath: "",
							IsTest:   false,
						}
						result.Usages[implType] = append(result.Usages[implType], implUsage)

						if config.Debug {
							fmt.Printf("Marking implementation %s as used in production via call graph\n", implType)
						}
					}
				}
			}
		}
	}
}
