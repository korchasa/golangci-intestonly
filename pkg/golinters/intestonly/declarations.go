// Package intestonly provides a linter that checks for code that is only used in tests but is not part of test files.
package intestonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// collectDeclarations processes all files to find declarations
func collectDeclarations(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()

		// Skip files that should be ignored based on naming patterns
		if shouldIgnoreFile(fileName, config) {
			if config.Debug {
				fmt.Printf("Skipping file for declarations: %s\n", fileName)
			}
			continue
		}

		// Process imports
		processImports(file, result)

		// Process declarations
		processFileDeclarations(file, fileName, result.CurrentPkgPath, result, config)
	}
}

// processImports extracts import information from a file
func processImports(file *ast.File, result *AnalysisResult) {
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
			result.ImportedPkgs[pkgName] = importPath
		}
	}
}

// processFileDeclarations collects declarations from a file
func processFileDeclarations(file *ast.File, fileName, pkgPath string, result *AnalysisResult, config *Config) {
	isTest := isTestFile(fileName, config)

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			processFuncDecl(n, fileName, pkgPath, result, config, isTest)
		case *ast.GenDecl:
			switch n.Tok {
			case token.TYPE:
				for _, spec := range n.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						processTypeSpec(typeSpec, fileName, pkgPath, result, config, isTest)
					}
				}
			case token.CONST:
				for _, spec := range n.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						processValueSpec(valueSpec, fileName, pkgPath, result, config, DeclConstant, isTest)
					}
				}
			case token.VAR:
				for _, spec := range n.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						processValueSpec(valueSpec, fileName, pkgPath, result, config, DeclVariable, isTest)
					}
				}
			}
		}
		return true
	})
}

// processFuncDecl processes a function declaration
func processFuncDecl(n *ast.FuncDecl, fileName, pkgPath string, result *AnalysisResult, config *Config, isTest bool) {
	if n.Name == nil || n.Name.Name == "" {
		return
	}

	name := n.Name.Name

	// Skip test helper identifiers unless they're explicit test cases
	if isTestHelperIdentifier(name, config) {
		return
	}

	importRef := pkgPath + "." + name
	result.ImportRefs[importRef] = name

	// Handle methods (functions with receivers)
	isMethod := false
	var receiverType string
	if n.Recv != nil && len(n.Recv.List) > 0 {
		isMethod = true
		switch expr := n.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if ident, ok := expr.X.(*ast.Ident); ok {
				receiverType = ident.Name
			}
		case *ast.Ident:
			receiverType = expr.Name
		}
	}

	declType := DeclFunction
	if isMethod {
		declType = DeclMethod
	}

	result.Declarations[name] = DeclInfo{
		Pos:          n.Name.Pos(),
		Name:         name,
		FilePath:     fileName,
		IsMethod:     isMethod,
		PkgPath:      pkgPath,
		ImportRef:    importRef,
		DeclType:     declType,
		ReceiverType: receiverType,
	}
	result.DeclPositions[n.Name.Pos()] = name

	// If this is a declaration in a test file, mark it as used in tests
	if isTest {
		usage := UsageInfo{
			Pos:      n.Pos(),
			FilePath: fileName,
			IsTest:   true,
		}
		result.TestUsages[name] = append(result.TestUsages[name], usage)
	}
}

// processTypeSpec processes a type declaration
func processTypeSpec(n *ast.TypeSpec, fileName, pkgPath string, result *AnalysisResult, config *Config, isTest bool) {
	if n.Name == nil || n.Name.Name == "" {
		return
	}

	name := n.Name.Name

	// Skip test helper identifiers unless they're explicit test cases
	if isTestHelperIdentifier(name, config) {
		return
	}

	importRef := pkgPath + "." + name
	result.ImportRefs[importRef] = name

	result.Declarations[name] = DeclInfo{
		Pos:       n.Name.Pos(),
		Name:      name,
		FilePath:  fileName,
		IsMethod:  false,
		PkgPath:   pkgPath,
		ImportRef: importRef,
		DeclType:  DeclTypeDecl,
	}
	result.DeclPositions[n.Name.Pos()] = name

	// If this is a declaration in a test file, mark it as used in tests
	if isTest {
		usage := UsageInfo{
			Pos:      n.Pos(),
			FilePath: fileName,
			IsTest:   true,
		}
		result.TestUsages[name] = append(result.TestUsages[name], usage)
	}
}

// processValueSpec processes a value declaration (constants, variables)
func processValueSpec(n *ast.ValueSpec, fileName, pkgPath string, result *AnalysisResult, config *Config, declType DeclType, isTest bool) {
	for _, name := range n.Names {
		if name == nil || name.Name == "" {
			continue
		}

		// Skip test helper identifiers unless they're explicit test cases
		if isTestHelperIdentifier(name.Name, config) {
			continue
		}

		importRef := pkgPath + "." + name.Name
		result.ImportRefs[importRef] = name.Name

		result.Declarations[name.Name] = DeclInfo{
			Pos:       name.Pos(),
			Name:      name.Name,
			FilePath:  fileName,
			IsMethod:  false,
			PkgPath:   pkgPath,
			ImportRef: importRef,
			DeclType:  declType,
		}
		result.DeclPositions[name.Pos()] = name.Name

		// If this is a declaration in a test file, mark it as used in tests
		if isTest {
			usage := UsageInfo{
				Pos:      name.Pos(),
				FilePath: fileName,
				IsTest:   true,
			}
			result.TestUsages[name.Name] = append(result.TestUsages[name.Name], usage)
		}
	}
}
