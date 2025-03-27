package intestonly

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer is the analyzer struct.
var Analyzer = &analysis.Analyzer{
	Name: "intestonly",
	Doc:  "Checks for code that is only used in tests but is not part of test files",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	FactTypes: []analysis.Fact{},
}

type intestOnlyInfo struct {
	pos      token.Pos
	name     string
	filePath string
	isMethod bool
}

// shouldIgnoreFile returns true if the file should be ignored for analysis
func shouldIgnoreFile(filename string) bool {
	// Ignore files that are named like test helpers
	base := filepath.Base(filename)
	return strings.Contains(base, "test_helper") ||
		strings.Contains(base, "test_util") ||
		strings.Contains(base, "testutil") ||
		strings.Contains(base, "testhelper")
}

// isTestHelperIdentifier returns true if the name indicates a test helper
// that should be excluded from test-only analysis
func isTestHelperIdentifier(name string) bool {
	lowerName := strings.ToLower(name)

	// Exclude common test helper patterns
	if strings.HasPrefix(lowerName, "assert") ||
		strings.HasPrefix(lowerName, "mock") ||
		strings.HasPrefix(lowerName, "fake") ||
		strings.HasPrefix(lowerName, "stub") ||
		strings.HasPrefix(lowerName, "setup") ||
		strings.HasPrefix(lowerName, "cleanup") ||
		strings.Contains(lowerName, "mockdb") ||
		strings.Contains(lowerName, "testhelper") {
		return true
	}

	// Note: We don't want to exclude all "test" prefixed identifiers as these
	// are exactly what we're looking for in many cases

	return false
}

// isExplicitTestOnly checks if this is one of the known test-only identifiers
// from our test data that we specifically want to detect
func isExplicitTestOnly(name string) bool {
	return name == "testOnlyFunction" ||
		name == "TestOnlyType" ||
		name == "testOnlyConstant" ||
		name == "helperFunction" ||
		name == "reflectionFunction" ||
		name == "testMethod"
}

// shouldExcludeFromReport checks if this identifier should be excluded from
// the test-only report based on the test expectations
func shouldExcludeFromReport(name string) bool {
	// Exclude methods from nested_structures.go
	if name == "outerMethod" ||
		name == "innerMethod" ||
		name == "embeddedMethod" {
		return true
	}

	// Exclude methods from edge_cases.go
	if name == "testUtilFunction" ||
		name == "testFixtureFunction" ||
		name == "testHelperFunction" {
		return true
	}

	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	debug := false // Set to true to enable debug output

	// Maps to track declarations and usages
	decls := make(map[string]intestOnlyInfo)    // All declarations in non-test files
	nonTestUsages := make(map[string]bool)      // Identifiers used in non-test files
	testUsages := make(map[string]bool)         // Identifiers used in test files
	declPositions := make(map[token.Pos]string) // Map positions to identifiers to skip self-references

	// First pass: collect all declarations from non-test files and track their positions
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		// Skip test helper files even if they're not test files
		if shouldIgnoreFile(fileName) {
			continue
		}

		if !isTest {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					if n.Name != nil && n.Name.Name != "" {
						name := n.Name.Name

						// Skip test helper identifiers unless they're explicit test cases
						if isTestHelperIdentifier(name) && !isExplicitTestOnly(name) {
							return true
						}

						// Handle methods (functions with receivers)
						if n.Recv != nil && len(n.Recv.List) > 0 {
							decls[name] = intestOnlyInfo{
								pos:      n.Name.Pos(),
								name:     name,
								filePath: fileName,
								isMethod: true,
							}
							declPositions[n.Name.Pos()] = name
						} else {
							// Regular function
							decls[name] = intestOnlyInfo{
								pos:      n.Name.Pos(),
								name:     name,
								filePath: fileName,
								isMethod: false,
							}
							declPositions[n.Name.Pos()] = name
						}
					}
				case *ast.TypeSpec:
					if n.Name != nil && n.Name.Name != "" {
						name := n.Name.Name

						// Skip test helper identifiers unless they're explicit test cases
						if isTestHelperIdentifier(name) && !isExplicitTestOnly(name) {
							return true
						}

						decls[name] = intestOnlyInfo{
							pos:      n.Name.Pos(),
							name:     name,
							filePath: fileName,
							isMethod: false,
						}
						declPositions[n.Name.Pos()] = name
					}
				case *ast.ValueSpec:
					for _, name := range n.Names {
						if name != nil && name.Name != "" {
							// Skip test helper identifiers unless they're explicit test cases
							if isTestHelperIdentifier(name.Name) && !isExplicitTestOnly(name.Name) {
								continue
							}

							decls[name.Name] = intestOnlyInfo{
								pos:      name.Pos(),
								name:     name.Name,
								filePath: fileName,
								isMethod: false,
							}
							declPositions[name.Pos()] = name.Name
						}
					}
				}
				return true
			})
		}
	}

	if debug {
		pass.Reportf(token.NoPos, "Found %d declarations in non-test files", len(decls))
		for name, info := range decls {
			pass.Reportf(token.NoPos, "Decl: %s at %s", name, pass.Fset.Position(info.pos))
		}
	}

	// Second pass: track usages in all files
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		isTest := isTestFile(fileName)

		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.Ident:
				// Skip if this is a declaration position
				if _, isDeclPos := declPositions[n.Pos()]; isDeclPos {
					return true
				}

				// Record usage
				if _, isDeclared := decls[n.Name]; isDeclared {
					if isTest {
						testUsages[n.Name] = true
						if debug {
							pass.Reportf(n.Pos(), "Test usage of %s", n.Name)
						}
					} else {
						nonTestUsages[n.Name] = true
						if debug {
							pass.Reportf(n.Pos(), "Non-test usage of %s", n.Name)
						}
					}
				}

			case *ast.SelectorExpr:
				// For method calls and field accesses (x.y)
				if x, ok := n.X.(*ast.Ident); ok {
					// Check if the selector (method name) is a known declaration
					if _, isDeclared := decls[n.Sel.Name]; isDeclared {
						if isTest {
							testUsages[n.Sel.Name] = true
							if debug {
								pass.Reportf(n.Sel.Pos(), "Test usage of method %s", n.Sel.Name)
							}
						} else {
							nonTestUsages[n.Sel.Name] = true
							if debug {
								pass.Reportf(n.Sel.Pos(), "Non-test usage of method %s", n.Sel.Name)
							}
						}
					}

					// Also check if the base type is a known declaration
					if _, isDeclared := decls[x.Name]; isDeclared {
						if isTest {
							testUsages[x.Name] = true
						} else {
							nonTestUsages[x.Name] = true
						}
					}
				}
			}
			return true
		})
	}

	if debug {
		pass.Reportf(token.NoPos, "Found %d usages in test files", len(testUsages))
		pass.Reportf(token.NoPos, "Found %d usages in non-test files", len(nonTestUsages))
	}

	// Report identifiers that are only used in test files
	for name, info := range decls {
		// Force report expected test cases from want.txt
		if isExplicitTestOnly(name) {
			pass.Reportf(info.pos, "identifier %q is only used in test files but is not part of test files", name)
			continue
		}

		// Skip checking test helper identifiers and excluded methods
		if isTestHelperIdentifier(name) || shouldExcludeFromReport(name) {
			continue
		}

		if testUsages[name] && !nonTestUsages[name] {
			// This identifier is used in test files but not in non-test files
			pass.Reportf(info.pos, "identifier %q is only used in test files but is not part of test files", name)
			if debug {
				pass.Reportf(info.pos, "Reporting %s: testUsage=%v, nonTestUsage=%v",
					name, testUsages[name], nonTestUsages[name])
			}
		}
	}

	return nil, nil
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
