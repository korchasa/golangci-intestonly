package intestonly

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestAnalyzeExportedIdentifierUsage(t *testing.T) {
	tests := []struct {
		name              string
		fileContent       string
		declarations      map[string]DeclInfo
		declPositions     map[token.Pos]string
		isTest            bool
		expectedTestUsage map[string]bool
		expectedUsage     map[string]bool
	}{
		{
			name: "Exported identifier used in non-test file",
			fileContent: `
package example

func main() {
	// Use ExportedFunc
	result := ExportedFunc()
}
`,
			declarations: map[string]DeclInfo{
				"ExportedFunc": {
					Name:     "ExportedFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			declPositions: map[token.Pos]string{},
			isTest:        false,
			expectedTestUsage: map[string]bool{},
			expectedUsage: map[string]bool{
				"ExportedFunc": true,
			},
		},
		{
			name: "Exported identifier used in test file",
			fileContent: `
package example

func TestSomething() {
	// Use ExportedFunc in test
	result := ExportedFunc()
}
`,
			declarations: map[string]DeclInfo{
				"ExportedFunc": {
					Name:     "ExportedFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			declPositions: map[token.Pos]string{},
			isTest:        true,
			expectedTestUsage: map[string]bool{
				"ExportedFunc": true,
			},
			expectedUsage: map[string]bool{},
		},
		{
			name: "Unexported identifier not tracked",
			fileContent: `
package example

func main() {
	// Use unexportedFunc
	result := unexportedFunc()
}
`,
			declarations: map[string]DeclInfo{
				"unexportedFunc": {
					Name:     "unexportedFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			declPositions: map[token.Pos]string{},
			isTest:        false,
			expectedTestUsage: map[string]bool{},
			expectedUsage: map[string]bool{},
		},
		{
			name: "Exported identifier not in declarations",
			fileContent: `
package example

func main() {
	// Use UndeclaredFunc
	result := UndeclaredFunc()
}
`,
			declarations: map[string]DeclInfo{
				"ExportedFunc": {
					Name:     "ExportedFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			declPositions: map[token.Pos]string{},
			isTest:        false,
			expectedTestUsage: map[string]bool{},
			expectedUsage: map[string]bool{},
		},
		{
			name: "Skip declaration positions",
			fileContent: `
package example

func ExportedFunc() {
	// This is a declaration, not a usage
}

func main() {
	// Use ExportedFunc
	result := ExportedFunc()
}
`,
			declarations: map[string]DeclInfo{
				"ExportedFunc": {
					Name:     "ExportedFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			// We'll populate declPositions in the test
			declPositions: map[token.Pos]string{},
			isTest:        false,
			expectedTestUsage: map[string]bool{},
			expectedUsage: map[string]bool{
				"ExportedFunc": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the file content
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.fileContent, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse file content: %v", err)
			}

			// Create test result and config
			result := NewAnalysisResult()
			config := &Config{Debug: false}

			// Add declarations to the result
			for name, info := range tt.declarations {
				result.Declarations[name] = info
			}

			// For the "Skip declaration positions" test, find the declaration position
			if tt.name == "Skip declaration positions" {
				ast.Inspect(file, func(n ast.Node) bool {
					if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "ExportedFunc" {
						result.DeclPositions[funcDecl.Name.Pos()] = "ExportedFunc"
					}
					return true
				})
			}

			// Run the function under test
			analyzeExportedIdentifierUsage(file, "test.go", tt.isTest, result, config)

			// Check test usages
			for name, expected := range tt.expectedTestUsage {
				if expected {
					if len(result.TestUsages[name]) == 0 {
						t.Errorf("Expected %s to have test usages, but it doesn't", name)
					}
				} else if _, exists := tt.declarations[name]; exists {
					if len(result.TestUsages[name]) > 0 {
						t.Errorf("Expected %s to not have test usages, but it does", name)
					}
				}
			}

			// Check non-test usages
			for name, expected := range tt.expectedUsage {
				if expected {
					if len(result.Usages[name]) == 0 {
						t.Errorf("Expected %s to have non-test usages, but it doesn't", name)
					}
				} else if _, exists := tt.declarations[name]; exists {
					if len(result.Usages[name]) > 0 {
						t.Errorf("Expected %s to not have non-test usages, but it does", name)
					}
				}
			}
		})
	}
}