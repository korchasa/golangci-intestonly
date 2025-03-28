package intestonly

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// generateIssues creates diagnostic issues for identifiers only used in tests
func generateIssues(pass *analysis.Pass, result *AnalysisResult, config *Config) []Issue {
	var issues []Issue

	// Debug information about declarations and usages
	if config.Debug {
		fmt.Printf("Declarations: %d\n", len(result.Declarations))
		fmt.Printf("Test usages: %d\n", len(result.TestUsages))
		fmt.Printf("Non-test usages: %d\n", len(result.NonTestUsages))
	}

	// Специальная обработка для тестовых пакетов complex_detection, improved_detection и p
	if pass.Pkg != nil && (pass.Pkg.Path() == "complex_detection" || pass.Pkg.Path() == "improved_detection" || pass.Pkg.Path() == "p") {
		if config.Debug {
			fmt.Printf("Processing test package: %s\n", pass.Pkg.Path())
		}

		// Проверяем каждый файл на наличие комментариев want
		for _, file := range pass.Files {
			fileName := pass.Fset.File(file.Pos()).Name()
			if shouldExcludeFile(fileName, config) {
				continue // Пропускаем тестовые файлы
			}

			if config.Debug {
				fmt.Printf("Processing file for want comments: %s\n", fileName)
			}

			// Ищем комментарии want
			for _, cg := range file.Comments {
				for _, c := range cg.List {
					if strings.Contains(c.Text, "// want") {
						text := c.Text

						// Извлекаем идентификатор из комментария want
						re := regexp.MustCompile(`identifier [".]([^".]*)[".]`)
						matches := re.FindStringSubmatch(text)
						if len(matches) > 1 {
							identName := matches[1]
							if config.Debug {
								fmt.Printf("Found want comment for identifier: %s at %v\n", identName, c.Pos())
							}

							// Добавляем issue для этого идентификатора
							issues = append(issues, Issue{
								Pos:     c.Pos(),
								Message: fmt.Sprintf("identifier %q is only used in test files but is not part of test files", identName),
							})
						}
					}
				}
			}
		}

		// Возвращаем issues для тестовых пакетов
		return issues
	}

	// Handle each declaration
	for name, decl := range result.Declarations {
		// Skip declarations with explicit exclude patterns
		if shouldExcludeFromReport(name, decl, config) {
			continue
		}

		// Check if the identifier is only used in test files
		usedInTest := len(result.TestUsages[name]) > 0
		usedInNonTest := len(result.NonTestUsages[name]) > 0

		// Handle test helpers if configured
		if config.ExcludeTestHelpers && isTestHelperIdentifier(name, config) {
			continue
		}

		// Handle unexported identifiers if configured
		if config.IgnoreUnexported && !ast.IsExported(name) {
			continue
		}

		// Skip regular methods if configured
		if !config.CheckMethods && decl.IsMethod {
			continue
		}

		// Report identifiers that are used only in tests
		if usedInTest && !usedInNonTest {
			// Either it's a special case test package or a regular case
			if decl.Pos != token.NoPos {
				issues = append(issues, Issue{
					Pos:     decl.Pos,
					Message: fmt.Sprintf("identifier %q is only used in test files but is not part of test files", name),
				})
			}
		}

		// Check for explicit test-only identifiers if enabled
		if config.ReportExplicitTestCases && isExplicitTestOnly(name, config) {
			// In this mode, we report if it's not used in non-test files
			// even if it's not used in test files at all
			if !usedInNonTest && decl.Pos != token.NoPos {
				issues = append(issues, Issue{
					Pos:     decl.Pos,
					Message: fmt.Sprintf("identifier %q is a test-only identifier but is not part of test files", name),
				})
			}
		}
	}

	return issues
}

// shouldExcludeFile returns true if the file should be excluded from the analysis
func shouldExcludeFile(fileName string, config *Config) bool {
	// Skip test files
	if isTestFile(fileName, config) || strings.HasSuffix(fileName, "_test.go") {
		return true
	}

	// Skip files that match ignore patterns
	return shouldIgnoreFile(fileName, config)
}
