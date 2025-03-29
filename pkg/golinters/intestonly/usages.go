// Package intestonly provides advanced identifier usage detection.
// This file implements a unified system for analyzing identifier usage across test and non-test files.
package intestonly

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// getFileName returns the file name for a given token position.
func getFileName(fset *token.FileSet, pos token.Pos) string {
	return fset.File(pos).Name()
}

// analyzeUsages processes all files to find where declarations are used
func analyzeUsages(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	// Используем инспектор для обхода всех узлов типа *ast.Ident во всех файлах
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.Ident)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ident, ok := n.(*ast.Ident)
		if !ok || ident.Name == "_" {
			return
		}
		// Новая функция для получения имени файла
		fileName := getFileName(pass.Fset, ident.Pos())
		isTest := isTestFile(fileName, config)
		if decl, exists := result.Declarations[ident.Name]; exists {
			if decl.Pos == ident.Pos() {
				return
			}
			usage := UsageInfo{
				Pos:      ident.Pos(),
				FilePath: fileName,
				IsTest:   isTest,
			}
			// Записываем использование посредством вспомогательной функции
			recordUsage(result, ident.Name, usage, isTest)
		}
	})

	// Дополнительные анализы не изменяются
	if config.EnableTypeEmbeddingAnalysis {
		analyzeTypeEmbedding(pass, result, config)
	}
	if config.EnableReflectionAnalysis {
		for _, file := range pass.Files {
			fileName := getFileName(pass.Fset, file.Pos())
			isTest := isTestFile(fileName, config)
			if !shouldIgnoreFile(fileName, config) {
				analyzeReflectionUsages(file, pass.Fset, isTest, result, config)
			}
		}
	}
	if config.EnableRegistryPatternDetection {
		for _, file := range pass.Files {
			fileName := getFileName(pass.Fset, file.Pos())
			isTest := isTestFile(fileName, config)
			if !shouldIgnoreFile(fileName, config) {
				analyzeRegistryPatternUsages(file, pass.Fset, isTest, result, config)
			}
		}
	}
}

// processFileUsages collects usages from a file
func processFileUsages(pass *analysis.Pass, file *ast.File, fileName string, result *AnalysisResult, config *Config, isTest bool) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.Ident)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ident, ok := n.(*ast.Ident)
		if !ok || ident.Name == "_" {
			return
		}

		if decl, exists := result.Declarations[ident.Name]; exists {
			if decl.Pos == ident.Pos() {
				return
			}
			usage := UsageInfo{
				Pos:      ident.Pos(),
				FilePath: getFileName(pass.Fset, ident.Pos()),
				IsTest:   isTest,
			}
			recordUsage(result, ident.Name, usage, isTest)
		}
	})
}

// recordUsage records an identifier usage in the appropriate map based on its test context.
func recordUsage(result *AnalysisResult, identName string, usage UsageInfo, isTest bool) {
	if isTest {
		result.TestUsages[identName] = append(result.TestUsages[identName], usage)
	} else {
		result.Usages[identName] = append(result.Usages[identName], usage)
	}
}

// analyzeTypeEmbedding analyzes type embedding in all files
func analyzeTypeEmbedding(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := getFileName(pass.Fset, file.Pos())
		isTest := isTestFile(fileName, config)
		analyzeTypeEmbeddingForFile(file, pass.Fset, isTest, result, config)
	}
}

// analyzeTypeEmbeddingForFile analyzes type embedding in a single file.
// Обновлённая версия с использованием inspector.New для работы с *ast.File.
func analyzeTypeEmbeddingForFile(file *ast.File, fset *token.FileSet, isTest bool, result *AnalysisResult, config *Config) {
	// Создаём инспектор для одного файла.
	insp := inspector.New([]*ast.File{file})
	// Будем обрабатывать только узлы типа *ast.TypeSpec.
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}
		// Обрабатываем только структуру
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return
		}
		// Проходим по всем полям структуры для поиска встроенных (embedded) типов.
		for _, field := range st.Fields.List {
			if len(field.Names) == 0 { // Это встроенное поле
				switch fieldType := field.Type.(type) {
				case *ast.Ident:
					// Прямое встраивание типа.
					embeddedTypeName := fieldType.Name
					if _, isDeclared := result.Declarations[embeddedTypeName]; isDeclared {
						usage := UsageInfo{
							Pos:      fieldType.Pos(),
							FilePath: getFileName(fset, fieldType.Pos()),
							IsTest:   isTest,
						}
						if isTest {
							result.TestUsages[embeddedTypeName] = append(result.TestUsages[embeddedTypeName], usage)
						} else {
							result.Usages[embeddedTypeName] = append(result.Usages[embeddedTypeName], usage)
						}
					}

				case *ast.SelectorExpr:
					// Встраивание импорта: pkg.Type
					if x, ok := fieldType.X.(*ast.Ident); ok {
						if importPath, ok := result.ImportedPkgs[x.Name]; ok {
							fullName := importPath + "." + fieldType.Sel.Name
							for declName, info := range result.Declarations {
								if info.ImportRef == fullName {
									usage := UsageInfo{
										Pos:      fieldType.Sel.Pos(),
										FilePath: getFileName(fset, fieldType.Sel.Pos()),
										IsTest:   isTest,
									}
									if isTest {
										result.TestUsages[declName] = append(result.TestUsages[declName], usage)
									} else {
										result.Usages[declName] = append(result.Usages[declName], usage)
									}
								}
							}
						}
					}
				}
			}
		}
	})
}

// analyzeReflectionUsages analyzes reflection-based usages in a file
func analyzeReflectionUsages(file *ast.File, fset *token.FileSet, isTest bool, result *AnalysisResult, config *Config) {
	// Создаём инспектор для одного файла.
	insp := inspector.New([]*ast.File{file})
	// Фильтр узлов: только *ast.CallExpr
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok || pkgIdent.Name != "reflect" {
			return
		}
		// Проверяем, что вызываемая функция – TypeOf или ValueOf.
		if sel.Sel.Name != "TypeOf" && sel.Sel.Name != "ValueOf" {
			return
		}
		if len(callExpr.Args) == 0 {
			return
		}
		ident, ok := callExpr.Args[0].(*ast.Ident)
		if !ok {
			return
		}
		if _, isDeclared := result.Declarations[ident.Name]; isDeclared {
			usage := UsageInfo{
				Pos:      ident.Pos(),
				FilePath: config.CurrentFile,
				IsTest:   isTest,
			}
			recordUsage(result, ident.Name, usage, isTest)
		}
	})
}

// analyzeRegistryPatternUsages analyzes registry pattern usages in a file
func analyzeRegistryPatternUsages(file *ast.File, fset *token.FileSet, isTest bool, result *AnalysisResult, config *Config) {
	// Создаем инспектор для одного файла.
	insp := inspector.New([]*ast.File{file})
	// Фильтр узлов: только *ast.CallExpr
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if _, ok := sel.X.(*ast.Ident); !ok {
			return
		}
		// Look for registration function calls
		for _, arg := range callExpr.Args {
			if ident, ok := arg.(*ast.Ident); ok {
				if _, isDeclared := result.Declarations[ident.Name]; isDeclared {
					usage := UsageInfo{
						Pos:      ident.Pos(),
						FilePath: config.CurrentFile,
						IsTest:   isTest,
					}
					recordUsage(result, ident.Name, usage, isTest)
				}
			}
		}
	})
}
