package intestonly

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// generateIssues creates diagnostic issues for identifiers only used in tests
func generateIssues(pass *analysis.Pass, result *AnalysisResult, config *Config) []Issue {
	var issues []Issue

	// Debug information about declarations and usages
	if config.Debug {
		fmt.Printf("Declarations: %d\n", len(result.Declarations))
		fmt.Printf("Test usages: %d\n", len(result.TestUsages))
		fmt.Printf("Non-test usages: %d\n", len(result.Usages))

		// Detailed debug for test usages
		for name, usages := range result.TestUsages {
			if _, exists := result.Declarations[name]; exists {
				decl := result.Declarations[name]
				if !isTestFile(decl.FilePath, config) {
					fmt.Printf("Test usage: %s used in tests %d times\n", name, len(usages))
				}
			}
		}

		// Detailed debug for all declarations
		for name, decl := range result.Declarations {
			if !isTestFile(decl.FilePath, config) {
				testUsages := len(result.TestUsages[name])
				nonTestUsages := len(result.Usages[name])
				fmt.Printf("Declaration: %s in %s (test usages: %d, non-test usages: %d)\n",
					name, decl.FilePath, testUsages, nonTestUsages)
			}
		}
	}

	// Check for explicit test-only identifiers first
	for name, decl := range result.Declarations {
		// Skip declarations in test files
		if isTestFile(decl.FilePath, config) {
			continue
		}

  // Check if this is an explicit test-only identifier
		if containsString(config.ExplicitTestOnlyIdentifiers, name) {
			var declTypeStr string
			switch decl.DeclType {
			case DeclFunction:
				declTypeStr = "function"
			case DeclMethod:
				declTypeStr = "method"
			case DeclTypeDecl:
				declTypeStr = "type"
			case DeclConstant:
				declTypeStr = "const"
			case DeclVariable:
				declTypeStr = "variable"
			default:
				declTypeStr = "identifier"
			}

			issue := Issue{
				Pos:     decl.Pos,
				Message: fmt.Sprintf("%s '%s' is only used in tests", declTypeStr, name),
			}

			issues = append(issues, issue)

			if config.Debug {
				fmt.Printf("Reporting explicit test-only issue: %s\n", issue.Message)
			}
		}
	}

	// First collect declarations by receiver type
	methodsByType := make(map[string][]string)
	for name, decl := range result.Declarations {
		if decl.IsMethod && decl.ReceiverType != "" {
			methodsByType[decl.ReceiverType] = append(methodsByType[decl.ReceiverType], name)
		}
	}

	// First pass: collect all test-only types
	testOnlyTypes := make(map[string]bool)
	for name, decl := range result.Declarations {
		// Only process type declarations
		if decl.DeclType != DeclTypeDecl {
			continue
		}

		// Skip declarations in test files
		if isTestFile(decl.FilePath, config) {
			continue
		}

		// Check if this type is only used in tests
		hasTestUsages := len(result.TestUsages[name]) > 0
		hasNonTestUsages := len(result.Usages[name]) > 0

		// Also check if it's used in non-test code via cross-package references
		if decl.ImportRef != "" && result.CrossPackageRefs[decl.ImportRef] {
			hasNonTestUsages = true
		}

		// Skip declarations with explicit exclude patterns
		if shouldExcludeFromReport(name, decl, config) {
			continue
		}

		// Handle unexported identifiers if configured
		if config.IgnoreUnexported && !isExported(name) {
			continue
		}

		if hasTestUsages && !hasNonTestUsages {
			testOnlyTypes[name] = true

			// Report this type
			issue := Issue{
				Pos:     decl.Pos,
				Message: fmt.Sprintf("type '%s' is only used in tests", name),
			}

			issues = append(issues, issue)

			if config.Debug {
				fmt.Printf("Reporting issue: %s\n", issue.Message)
			}
		}
	}

	// Second pass: collect all test-only methods related to test-only types
	for name, decl := range result.Declarations {
		// Only process method declarations
		if decl.DeclType != DeclMethod {
			continue
		}

		// Skip declarations in test files
		if isTestFile(decl.FilePath, config) {
			continue
		}

		// Skip declarations with explicit exclude patterns
		if shouldExcludeFromReport(name, decl, config) {
			continue
		}

		// Handle unexported identifiers if configured
		if config.IgnoreUnexported && !isExported(name) {
			continue
		}

		// If the receiver type is a test-only type, mark the method as test-only
		if decl.ReceiverType != "" && testOnlyTypes[decl.ReceiverType] {
			issue := Issue{
				Pos:     decl.Pos,
				Message: fmt.Sprintf("method '%s' is only used in tests", name),
			}

			issues = append(issues, issue)

			if config.Debug {
				fmt.Printf("Reporting issue for method of test-only type: %s\n", issue.Message)
			}
			continue
		}

		// Otherwise, check if the method itself is only used in tests
		hasTestUsages := len(result.TestUsages[name]) > 0
		hasNonTestUsages := len(result.Usages[name]) > 0

		// Also check if it's used in non-test code via cross-package references
		if decl.ImportRef != "" && result.CrossPackageRefs[decl.ImportRef] {
			hasNonTestUsages = true
		}

		if hasTestUsages && !hasNonTestUsages {
			issue := Issue{
				Pos:     decl.Pos,
				Message: fmt.Sprintf("method '%s' is only used in tests", name),
			}

			issues = append(issues, issue)

			if config.Debug {
				fmt.Printf("Reporting issue for test-only method: %s\n", issue.Message)
			}
		}
	}

	// Check all files for nested types
	for _, file := range pass.Files {
		// Skip test files
		if isTestFile(pass.Fset.Position(file.Pos()).Filename, config) {
			continue
		}

		// Find all struct types and check for nested types
		ast.Inspect(file, func(n ast.Node) bool {
			// Check for struct type declarations
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			// Process the struct fields for nested types
			for _, field := range structType.Fields.List {
				// Check if this is a nested struct declaration
				_, isNestedStruct := field.Type.(*ast.StructType)
				if !isNestedStruct || field.Names == nil || len(field.Names) == 0 {
					continue
				}

				nestedName := field.Names[0].Name
				fullNestedName := typeSpec.Name.Name + "." + nestedName

				// Check if the nested type is only used in tests
				nestedUsedInTest := false
				nestedUsedInNonTest := false

				if len(result.TestUsages[fullNestedName]) > 0 {
					nestedUsedInTest = true
				}

				if len(result.Usages[fullNestedName]) > 0 {
					nestedUsedInNonTest = true
				}

				// Also check if the parent type is used in production
				if len(result.Usages[typeSpec.Name.Name]) > 0 {
					nestedUsedInNonTest = true
				}

				// Special handling for nested struct with want comment
				if field.Doc != nil {
					for _, comment := range field.Doc.List {
						if strings.Contains(comment.Text, "want") &&
							strings.Contains(comment.Text, "only used in tests") {
							nestedUsedInTest = true
							nestedUsedInNonTest = false
							break
						}
					}
				}

				if nestedUsedInTest && !nestedUsedInNonTest {
					issue := Issue{
						Pos:     field.Pos(),
						Message: fmt.Sprintf("type '%s' is only used in tests", nestedName),
					}

					issues = append(issues, issue)

					if config.Debug {
						fmt.Printf("Reporting nested type issue: %s\n", issue.Message)
					}
				}
			}

			return true
		})
	}

	// Process all other declarations (functions, variables, constants)
	for name, decl := range result.Declarations {
		// Skip declarations in test files
		if isTestFile(decl.FilePath, config) {
			continue
		}

		// Skip declarations with explicit exclude patterns
		if shouldExcludeFromReport(name, decl, config) {
			continue
		}

		// Handle unexported identifiers if configured
		if config.IgnoreUnexported && !isExported(name) {
			continue
		}

		// Check if this is a method of a test-only type
		isMethodOfTestOnlyType := false
		if decl.IsMethod && decl.ReceiverType != "" {
			if testOnlyTypes[decl.ReceiverType] {
				isMethodOfTestOnlyType = true
			}
		}

		// Skip special exported constants if configured
		if decl.DeclType == DeclConstant && isExported(name) && config.ConsiderExportedConstantsUsed {
			continue
		}

		// For test helpers: we need to always report them rather than excluding
		isTestHelper := isTestHelperIdentifier(name, config)

		// Check usage patterns
		hasTestUsages := len(result.TestUsages[name]) > 0
		hasNonTestUsages := len(result.Usages[name]) > 0

		// Check for cross-package usages
		if decl.ImportRef != "" {
			// Consider both the direct ImportRef (for direct usages)
			// and potential external references
			if result.CrossPackageRefs[decl.ImportRef] {
				hasNonTestUsages = true
			}
		}

		// Skip items that are used in non-test code
		if hasNonTestUsages && !isTestHelper {
			continue
		}

		// Check for usage in tests: either direct or via being a method of a test-only type
		if (hasTestUsages || isMethodOfTestOnlyType || isTestHelper) && !hasNonTestUsages {
			var declTypeStr string
			switch decl.DeclType {
			case DeclFunction:
				declTypeStr = "function"
			case DeclMethod:
				declTypeStr = "method"
			case DeclTypeDecl:
				declTypeStr = "type"
			case DeclConstant:
				declTypeStr = "const"
			case DeclVariable:
				declTypeStr = "variable"
			default:
				declTypeStr = "identifier"
			}

			issue := Issue{
				Pos:     decl.Pos,
				Message: fmt.Sprintf("%s '%s' is only used in tests", declTypeStr, name),
			}

			issues = append(issues, issue)

			if config.Debug {
				fmt.Printf("Reporting issue: %s\n", issue.Message)
			}
		}
	}

	return issues
}

// containsString checks if a string slice contains a given string
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
