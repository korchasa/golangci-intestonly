package intestonly

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestProcessImports(t *testing.T) {
	// Create test data
	src := `
package example

import (
	"fmt"
	"os"
	alias "path/filepath"
	. "strings"
	_ "net/http"
)
`
	// Parse the source into an AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Create analysis result
	result := NewAnalysisResult()

	// Process imports
	processImports(file, result)

	// Check imports are correctly processed
	expected := map[string]string{
		"fmt":     "fmt",
		"os":      "os",
		"alias":   "path/filepath",
		"strings": "strings",
		"http":    "net/http",
	}

	// Check that all expected imports were recorded
	for pkgName, importPath := range expected {
		path, ok := result.ImportedPkgs[pkgName]
		if !ok {
			t.Errorf("Import for package %s was not recorded", pkgName)
			continue
		}
		if path != importPath {
			t.Errorf("Import path for %s = %s, want %s", pkgName, path, importPath)
		}
	}

	// Check that no unexpected imports were recorded
	for pkgName := range result.ImportedPkgs {
		if _, ok := expected[pkgName]; !ok {
			t.Errorf("Unexpected import recorded: %s -> %s", pkgName, result.ImportedPkgs[pkgName])
		}
	}
}

func TestProcessFuncDecl(t *testing.T) {
	tests := []struct {
		name           string
		src            string
		isTest         bool
		expectDecl     bool
		expectedName   string
		expectedMethod bool
		expectedType   DeclType
		expectedRecv   string
	}{
		{
			name: "regular function",
			src: `
package example
func myFunction() {}
`,
			isTest:         false,
			expectDecl:     true,
			expectedName:   "myFunction",
			expectedMethod: false,
			expectedType:   DeclFunction,
			expectedRecv:   "",
		},
		{
			name: "method with pointer receiver",
			src: `
package example
type MyType struct{}
func (m *MyType) MyMethod() {}
`,
			isTest:         false,
			expectDecl:     true,
			expectedName:   "MyMethod",
			expectedMethod: true,
			expectedType:   DeclMethod,
			expectedRecv:   "MyType",
		},
		{
			name: "method with value receiver",
			src: `
package example
type AnotherType int
func (a AnotherType) Method() {}
`,
			isTest:         false,
			expectDecl:     true,
			expectedName:   "Method",
			expectedMethod: true,
			expectedType:   DeclMethod,
			expectedRecv:   "AnotherType",
		},
		{
			name: "test helper function should be skipped",
			src: `
package example
func assertSomething() {}
`,
			isTest:     false,
			expectDecl: false, // Should be skipped due to test helper name
		},
		{
			name: "function in test file",
			src: `
package example
func testFunction() {}
`,
			isTest:         true,
			expectDecl:     true,
			expectedName:   "testFunction",
			expectedMethod: false,
			expectedType:   DeclFunction,
			expectedRecv:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source into an AST
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "example.go", tt.src, 0)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Find the function declaration
			var funcDecl *ast.FuncDecl
			for _, decl := range file.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok {
					funcDecl = fd
					break
				}
			}
			if funcDecl == nil {
				t.Fatalf("Failed to find function declaration in AST")
			}

			// Create analysis result and config
			result := NewAnalysisResult()
			config := &Config{
				Debug:              false,
				ExcludeTestHelpers: true,
				TestHelperPatterns: []string{"assert", "mock", "fake", "stub", "test"},
			}

			// Process function declaration
			processFuncDecl(funcDecl, "example.go", "github.com/example", result, config, tt.isTest)

			// Check results
			if tt.expectDecl {
				// Check declaration was recorded
				decl, ok := result.Declarations[tt.expectedName]
				if !ok {
					t.Errorf("Declaration for %s was not recorded", tt.expectedName)
					return
				}

				// Check declaration details
				if decl.Name != tt.expectedName {
					t.Errorf("Declaration name = %s, want %s", decl.Name, tt.expectedName)
				}
				if decl.IsMethod != tt.expectedMethod {
					t.Errorf("Declaration IsMethod = %v, want %v", decl.IsMethod, tt.expectedMethod)
				}
				if decl.DeclType != tt.expectedType {
					t.Errorf("Declaration DeclType = %v, want %v", decl.DeclType, tt.expectedType)
				}
				if decl.ReceiverType != tt.expectedRecv {
					t.Errorf("Declaration ReceiverType = %s, want %s", decl.ReceiverType, tt.expectedRecv)
				}
				if decl.PkgPath != "github.com/example" {
					t.Errorf("Declaration PkgPath = %s, want %s", decl.PkgPath, "github.com/example")
				}
				if decl.FilePath != "example.go" {
					t.Errorf("Declaration FilePath = %s, want %s", decl.FilePath, "example.go")
				}

				// Check test usage
				if tt.isTest {
					usages, ok := result.TestUsages[tt.expectedName]
					if !ok {
						t.Errorf("Expected test usage for %s, but none was recorded", tt.expectedName)
					} else if len(usages) != 1 {
						t.Errorf("Expected 1 test usage for %s, got %d", tt.expectedName, len(usages))
					} else {
						usage := usages[0]
						if !usage.IsTest {
							t.Errorf("Usage IsTest = %v, want true", usage.IsTest)
						}
						if usage.FilePath != "example.go" {
							t.Errorf("Usage FilePath = %s, want %s", usage.FilePath, "example.go")
						}
					}
				} else {
					usages, ok := result.TestUsages[tt.expectedName]
					if ok && len(usages) > 0 {
						t.Errorf("Unexpected test usage recorded for %s", tt.expectedName)
					}
				}
			} else {
				// Check declaration was not recorded
				for name := range result.Declarations {
					if name == "assertSomething" {
						t.Errorf("Test helper function %s should have been skipped", name)
					}
				}
			}
		})
	}
}

func TestProcessTypeSpec(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		isTest     bool
		expectDecl bool
	}{
		{
			name: "regular type",
			src: `
package example
type MyType struct {}
`,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "interface type",
			src: `
package example
type MyInterface interface {
	Do() error
}
`,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "type alias",
			src: `
package example
type MyInt int
`,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "test helper type should be skipped",
			src: `
package example
type MockType struct {}
`,
			isTest:     false,
			expectDecl: false, // Should be skipped due to test helper name
		},
		{
			name: "type in test file",
			src: `
package example
type TestType struct {}
`,
			isTest:     true,
			expectDecl: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source into an AST
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "example.go", tt.src, 0)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Find the type declaration
			var typeSpec *ast.TypeSpec
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					typeSpec = ts
					return false
				}
				return true
			})
			if typeSpec == nil {
				t.Fatalf("Failed to find type declaration in AST")
			}

			// Create analysis result and config
			result := NewAnalysisResult()
			config := &Config{
				Debug:              false,
				ExcludeTestHelpers: true,
				TestHelperPatterns: []string{"assert", "mock", "fake", "stub", "test"},
			}

			// Process type declaration
			processTypeSpec(typeSpec, "example.go", "github.com/example", result, config, tt.isTest)

			// Extract expected name from the source
			expectedName := ""
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					expectedName = ts.Name.Name
					return false
				}
				return true
			})

			// Check results
			if tt.expectDecl {
				// Check declaration was recorded
				decl, ok := result.Declarations[expectedName]
				if !ok {
					t.Errorf("Declaration for %s was not recorded", expectedName)
					return
				}

				// Check declaration details
				if decl.Name != expectedName {
					t.Errorf("Declaration name = %s, want %s", decl.Name, expectedName)
				}
				if decl.IsMethod {
					t.Errorf("Declaration IsMethod = %v, want false", decl.IsMethod)
				}
				if decl.DeclType != DeclTypeDecl {
					t.Errorf("Declaration DeclType = %v, want %v", decl.DeclType, DeclTypeDecl)
				}
				if decl.PkgPath != "github.com/example" {
					t.Errorf("Declaration PkgPath = %s, want %s", decl.PkgPath, "github.com/example")
				}
				if decl.FilePath != "example.go" {
					t.Errorf("Declaration FilePath = %s, want %s", decl.FilePath, "example.go")
				}

				// Check test usage
				if tt.isTest {
					usages, ok := result.TestUsages[expectedName]
					if !ok {
						t.Errorf("Expected test usage for %s, but none was recorded", expectedName)
					} else if len(usages) != 1 {
						t.Errorf("Expected 1 test usage for %s, got %d", expectedName, len(usages))
					} else {
						usage := usages[0]
						if !usage.IsTest {
							t.Errorf("Usage IsTest = %v, want true", usage.IsTest)
						}
						if usage.FilePath != "example.go" {
							t.Errorf("Usage FilePath = %s, want %s", usage.FilePath, "example.go")
						}
					}
				} else {
					usages, ok := result.TestUsages[expectedName]
					if ok && len(usages) > 0 {
						t.Errorf("Unexpected test usage recorded for %s", expectedName)
					}
				}
			} else {
				// Check declaration was not recorded
				for name := range result.Declarations {
					if name == "MockType" {
						t.Errorf("Test helper type %s should have been skipped", name)
					}
				}
			}
		})
	}
}

func TestProcessValueSpec(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		declType   DeclType
		isTest     bool
		expectDecl bool
	}{
		{
			name: "constant",
			src: `
package example
const MyConst = 42
`,
			declType:   DeclConstant,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "multiple constants",
			src: `
package example
const (
	First = 1
	Second = 2
)
`,
			declType:   DeclConstant,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "variable",
			src: `
package example
var MyVar = "value"
`,
			declType:   DeclVariable,
			isTest:     false,
			expectDecl: true,
		},
		{
			name: "test helper variable should be skipped",
			src: `
package example
var testVariable = "test"
`,
			declType:   DeclVariable,
			isTest:     false,
			expectDecl: false, // Should be skipped due to test helper name
		},
		{
			name: "const in test file",
			src: `
package example
const TestConst = true
`,
			declType:   DeclConstant,
			isTest:     true,
			expectDecl: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source into an AST
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "example.go", tt.src, 0)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Find the value spec
			var valueSpec *ast.ValueSpec
			ast.Inspect(file, func(n ast.Node) bool {
				if vs, ok := n.(*ast.ValueSpec); ok {
					valueSpec = vs
					return false
				}
				return true
			})
			if valueSpec == nil {
				t.Fatalf("Failed to find value spec in AST")
			}

			// Create analysis result and config
			result := NewAnalysisResult()
			config := &Config{
				Debug:              false,
				ExcludeTestHelpers: true,
				TestHelperPatterns: []string{"assert", "mock", "fake", "stub", "test"},
			}

			// Process value spec
			processValueSpec(valueSpec, "example.go", "github.com/example", result, config, tt.declType, tt.isTest)

			// Find all expected names from the value spec
			var expectedNames []string
			for _, name := range valueSpec.Names {
				expectedNames = append(expectedNames, name.Name)
			}

			// Check results
			if tt.expectDecl {
				// For each expected name
				for _, expectedName := range expectedNames {
					// Skip test helper names
					if tt.expectDecl == false && isTestHelperIdentifier(expectedName, config) {
						continue
					}

					// Check declaration was recorded
					decl, ok := result.Declarations[expectedName]
					if !ok {
						t.Errorf("Declaration for %s was not recorded", expectedName)
						continue
					}

					// Check declaration details
					if decl.Name != expectedName {
						t.Errorf("Declaration name = %s, want %s", decl.Name, expectedName)
					}
					if decl.IsMethod {
						t.Errorf("Declaration IsMethod = %v, want false", decl.IsMethod)
					}
					if decl.DeclType != tt.declType {
						t.Errorf("Declaration DeclType = %v, want %v", decl.DeclType, tt.declType)
					}
					if decl.PkgPath != "github.com/example" {
						t.Errorf("Declaration PkgPath = %s, want %s", decl.PkgPath, "github.com/example")
					}
					if decl.FilePath != "example.go" {
						t.Errorf("Declaration FilePath = %s, want %s", decl.FilePath, "example.go")
					}

					// Check test usage
					if tt.isTest {
						usages, ok := result.TestUsages[expectedName]
						if !ok {
							t.Errorf("Expected test usage for %s, but none was recorded", expectedName)
						} else if len(usages) != 1 {
							t.Errorf("Expected 1 test usage for %s, got %d", expectedName, len(usages))
						} else {
							usage := usages[0]
							if !usage.IsTest {
								t.Errorf("Usage IsTest = %v, want true", usage.IsTest)
							}
							if usage.FilePath != "example.go" {
								t.Errorf("Usage FilePath = %s, want %s", usage.FilePath, "example.go")
							}
						}
					} else {
						usages, ok := result.TestUsages[expectedName]
						if ok && len(usages) > 0 {
							t.Errorf("Unexpected test usage recorded for %s", expectedName)
						}
					}
				}
			} else {
				// Check declaration was not recorded
				for _, expectedName := range expectedNames {
					if _, ok := result.Declarations[expectedName]; ok && isTestHelperIdentifier(expectedName, config) {
						t.Errorf("Test helper %s should have been skipped", expectedName)
					}
				}
			}
		})
	}
}

func TestIsTestHelperIdentifierInDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		patterns []string
		want     bool
	}{
		{
			name:     "assert function",
			ident:    "assertEqual",
			patterns: []string{"assert", "mock", "test"},
			want:     true,
		},
		{
			name:     "mock type",
			ident:    "MockService",
			patterns: []string{"assert", "mock", "test"},
			want:     true,
		},
		{
			name:     "test case",
			ident:    "testCase",
			patterns: []string{"assert", "mock", "test"},
			want:     true,
		},
		{
			name:     "case-insensitive test prefix",
			ident:    "TestStruct",
			patterns: []string{"assert", "mock", "test"},
			want:     true,
		},
		{
			name:     "normal identifier",
			ident:    "userService",
			patterns: []string{"assert", "mock", "test"},
			want:     false,
		},
		{
			name:     "normal identifier with pattern as substring",
			ident:    "assetManager",
			patterns: []string{"assert", "mock", "test"},
			want:     false, // Should not match "asset" as it's not an exact pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ExcludeTestHelpers: true,
				TestHelperPatterns: tt.patterns,
			}
			got := isTestHelperIdentifier(tt.ident, config)
			if got != tt.want {
				t.Errorf("isTestHelperIdentifier(%s) = %v, want %v", tt.ident, got, tt.want)
			}
		})
	}
}
