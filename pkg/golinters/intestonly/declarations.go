package intestonly

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// collectDeclarations processes all non-test files to find declarations
func collectDeclarations(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()

		// Skip test files and test helpers
		if isTestFile(fileName) || shouldIgnoreFile(fileName, config) {
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
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			processFuncDecl(n, fileName, pkgPath, result, config)
		case *ast.TypeSpec:
			processTypeSpec(n, fileName, pkgPath, result, config)
		case *ast.ValueSpec:
			processValueSpec(n, fileName, pkgPath, result, config)
		}
		return true
	})
}

// processFuncDecl processes a function declaration
func processFuncDecl(n *ast.FuncDecl, fileName, pkgPath string, result *AnalysisResult, config *Config) {
	if n.Name == nil || n.Name.Name == "" {
		return
	}

	name := n.Name.Name

	// Skip test helper identifiers unless they're explicit test cases
	if isTestHelperIdentifier(name, config) && !isExplicitTestOnly(name, config) {
		return
	}

	importRef := pkgPath + "." + name
	result.ImportRefs[importRef] = name

	// Handle methods (functions with receivers)
	isMethod := false
	if n.Recv != nil && len(n.Recv.List) > 0 {
		isMethod = true
	}

	result.Declarations[name] = DeclInfo{
		Pos:       n.Name.Pos(),
		Name:      name,
		FilePath:  fileName,
		IsMethod:  isMethod,
		PkgPath:   pkgPath,
		ImportRef: importRef,
	}
	result.DeclPositions[n.Name.Pos()] = name
}

// processTypeSpec processes a type declaration
func processTypeSpec(n *ast.TypeSpec, fileName, pkgPath string, result *AnalysisResult, config *Config) {
	if n.Name == nil || n.Name.Name == "" {
		return
	}

	name := n.Name.Name

	// Skip test helper identifiers unless they're explicit test cases
	if isTestHelperIdentifier(name, config) && !isExplicitTestOnly(name, config) {
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
	}
	result.DeclPositions[n.Name.Pos()] = name
}

// processValueSpec processes a value declaration (constants, variables)
func processValueSpec(n *ast.ValueSpec, fileName, pkgPath string, result *AnalysisResult, config *Config) {
	for _, name := range n.Names {
		if name == nil || name.Name == "" {
			continue
		}

		// Skip test helper identifiers unless they're explicit test cases
		if isTestHelperIdentifier(name.Name, config) && !isExplicitTestOnly(name.Name, config) {
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
		}
		result.DeclPositions[name.Pos()] = name.Name
	}
}
