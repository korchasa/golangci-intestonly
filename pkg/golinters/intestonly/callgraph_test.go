package intestonly

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestIsFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		declType DeclType
		want     bool
	}{
		{
			name:     "should return true for regular function",
			funcName: "myFunction",
			declType: DeclFunction,
			want:     true,
		},
		{
			name:     "should return true for method",
			funcName: "MyType.MyMethod",
			declType: DeclMethod,
			want:     true,
		},
		{
			name:     "should return false for type declaration",
			funcName: "MyType",
			declType: DeclTypeDecl,
			want:     false,
		},
		{
			name:     "should return false for constant",
			funcName: "MyConstant",
			declType: DeclConstant,
			want:     false,
		},
		{
			name:     "should return false for variable",
			funcName: "myVariable",
			declType: DeclVariable,
			want:     false,
		},
		{
			name:     "should return false for non-existent declaration",
			funcName: "nonExistentFunc",
			declType: DeclUnknown,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a result with the test declarations
			result := NewAnalysisResult()

			// Only add declaration info if it's not testing non-existent declaration
			if tt.funcName != "nonExistentFunc" {
				result.Declarations[tt.funcName] = DeclInfo{
					Name:     tt.funcName,
					DeclType: tt.declType,
				}
			}

			got := isFunctionName(tt.funcName, result)
			if got != tt.want {
				t.Errorf("isFunctionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropagateTestUsage(t *testing.T) {
	tests := []struct {
		name              string
		setupGraph        func(*AnalysisResult)
		rootFunc          string
		expectedTestUsage []string
	}{
		{
			name: "should propagate to direct callees",
			setupGraph: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["rootFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["calledFunc1"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["calledFunc2"] = DeclInfo{DeclType: DeclFunction}

				// Setup call graph
				result.CallGraph["rootFunc"] = []string{"calledFunc1", "calledFunc2"}
			},
			rootFunc:          "rootFunc",
			expectedTestUsage: []string{"calledFunc1", "calledFunc2"},
		},
		{
			name: "should handle multi-level propagation",
			setupGraph: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["rootFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["levelOneFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["levelTwoFunc"] = DeclInfo{DeclType: DeclFunction}

				// Setup multi-level call graph
				result.CallGraph["rootFunc"] = []string{"levelOneFunc"}
				result.CallGraph["levelOneFunc"] = []string{"levelTwoFunc"}
			},
			rootFunc:          "rootFunc",
			expectedTestUsage: []string{"levelOneFunc", "levelTwoFunc"},
		},
		{
			name: "should handle cycles in call graph",
			setupGraph: func(result *AnalysisResult) {
				// Setup declarations for functions in a cycle
				result.Declarations["funcA"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["funcB"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["funcC"] = DeclInfo{DeclType: DeclFunction}

				// Create a cycle: A -> B -> C -> A
				result.CallGraph["funcA"] = []string{"funcB"}
				result.CallGraph["funcB"] = []string{"funcC"}
				result.CallGraph["funcC"] = []string{"funcA"}
			},
			rootFunc:          "funcA",
			expectedTestUsage: []string{"funcB", "funcC"},
		},
		{
			name: "should skip non-declared callees",
			setupGraph: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["rootFunc"] = DeclInfo{DeclType: DeclFunction}
				// We deliberately don't declare "nonDeclaredFunc"

				// Setup call graph with a non-declared function
				result.CallGraph["rootFunc"] = []string{"nonDeclaredFunc", "anotherNonDeclaredFunc"}
			},
			rootFunc:          "rootFunc",
			expectedTestUsage: []string{}, // No propagation to non-declared functions
		},
		{
			name: "should not re-mark already marked functions",
			setupGraph: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["rootFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["alreadyMarkedFunc"] = DeclInfo{DeclType: DeclFunction}

				// Setup call graph
				result.CallGraph["rootFunc"] = []string{"alreadyMarkedFunc"}

				// Pre-mark the function as used in tests
				result.TestUsages["alreadyMarkedFunc"] = []UsageInfo{{
					Pos:      token.NoPos,
					FilePath: "test_file.go",
					IsTest:   true,
				}}
			},
			rootFunc:          "rootFunc",
			expectedTestUsage: []string{"alreadyMarkedFunc"}, // Should still be marked but not changed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test analysis result
			result := NewAnalysisResult()

			// Setup the test call graph
			tt.setupGraph(result)

			// Execute the function
			config := &Config{Debug: false}
			propagateTestUsage(tt.rootFunc, result, make(map[string]bool), config)

			// Verify the test usages were correctly propagated
			for _, expectedFunc := range tt.expectedTestUsage {
				usages, exists := result.TestUsages[expectedFunc]
				if !exists {
					t.Errorf("Expected function %s to be marked as used in tests", expectedFunc)
					continue
				}

				// If the function was pre-marked, we don't add an additional usage
				if expectedFunc != "alreadyMarkedFunc" {
					if len(usages) != 1 {
						t.Errorf("Expected exactly one test usage for %s", expectedFunc)
						continue
					}
					if usages[0].Pos != token.NoPos {
						t.Errorf("Expected NoPos for synthetic usage, got %v", usages[0].Pos)
					}
					if usages[0].FilePath != "" {
						t.Errorf("Expected empty file path for synthetic usage, got %s", usages[0].FilePath)
					}
					if !usages[0].IsTest {
						t.Errorf("Expected IsTest to be true for synthetic usage")
					}
				}
			}

			// Make sure no unexpected functions were marked
			for funcName := range result.TestUsages {
				if funcName != tt.rootFunc && funcName != "alreadyMarkedFunc" {
					found := false
					for _, expected := range tt.expectedTestUsage {
						if funcName == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected function %s was marked as used in tests", funcName)
					}
				}
			}
		})
	}
}

func TestPropagateCallDependencies(t *testing.T) {
	tests := []struct {
		name                string
		setupResult         func(*AnalysisResult)
		expectedPropagation map[string]bool // functions that should be marked as used in tests
	}{
		{
			name: "should propagate to functions called by test-only functions",
			setupResult: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["testOnlyFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["helperFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["prodFunc"] = DeclInfo{DeclType: DeclFunction}

				// Mark testOnlyFunc as used in tests
				result.TestUsages["testOnlyFunc"] = []UsageInfo{{
					Pos:      token.Pos(100),
					FilePath: "test_file.go",
					IsTest:   true,
				}}

				// Mark prodFunc as used in production
				result.Usages["prodFunc"] = []UsageInfo{{
					Pos:      token.Pos(200),
					FilePath: "main.go",
					IsTest:   false,
				}}

				// Setup call graph
				result.CallGraph["testOnlyFunc"] = []string{"helperFunc"}
			},
			expectedPropagation: map[string]bool{
				"helperFunc": true,
			},
		},
		{
			name: "should not propagate to functions used in both test and production",
			setupResult: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["testOnlyFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["dualUsageFunc"] = DeclInfo{DeclType: DeclFunction}

				// Mark testOnlyFunc as used in tests
				result.TestUsages["testOnlyFunc"] = []UsageInfo{{
					Pos:      token.Pos(100),
					FilePath: "test_file.go",
					IsTest:   true,
				}}

				// Mark dualUsageFunc as used in both test and production
				result.TestUsages["dualUsageFunc"] = []UsageInfo{{
					Pos:      token.Pos(200),
					FilePath: "test_file.go",
					IsTest:   true,
				}}
				result.Usages["dualUsageFunc"] = []UsageInfo{{
					Pos:      token.Pos(300),
					FilePath: "main.go",
					IsTest:   false,
				}}

				// Setup call graph
				result.CallGraph["testOnlyFunc"] = []string{"dualUsageFunc"}
			},
			expectedPropagation: map[string]bool{
				// dualUsageFunc should not be in expectedPropagation as it's already used in production
			},
		},
		{
			name: "should skip non-function declarations",
			setupResult: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["testOnlyFunc"] = DeclInfo{DeclType: DeclFunction}
				result.Declarations["someType"] = DeclInfo{DeclType: DeclTypeDecl}
				result.Declarations["someConst"] = DeclInfo{DeclType: DeclConstant}

				// Mark all as used in tests only
				result.TestUsages["testOnlyFunc"] = []UsageInfo{{IsTest: true}}
				result.TestUsages["someType"] = []UsageInfo{{IsTest: true}}
				result.TestUsages["someConst"] = []UsageInfo{{IsTest: true}}

				// No production usages set

				// Setup call graph (though type and const shouldn't be processed)
				result.CallGraph["testOnlyFunc"] = []string{"someOtherFunc"}
			},
			expectedPropagation: map[string]bool{
				// Only functions should be processed, someType and someConst should be skipped
			},
		},
		{
			name: "should handle methods correctly",
			setupResult: func(result *AnalysisResult) {
				// Setup declarations
				result.Declarations["TestStruct.TestMethod"] = DeclInfo{
					DeclType:     DeclMethod,
					IsMethod:     true,
					ReceiverType: "TestStruct",
				}
				result.Declarations["helperFunc"] = DeclInfo{DeclType: DeclFunction}

				// Mark method as used in tests
				result.TestUsages["TestStruct.TestMethod"] = []UsageInfo{{IsTest: true}}

				// Setup call graph
				result.CallGraph["TestStruct.TestMethod"] = []string{"helperFunc"}
			},
			expectedPropagation: map[string]bool{
				"helperFunc": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test analysis result
			result := NewAnalysisResult()

			// Setup the test result
			tt.setupResult(result)

			// Execute the function
			config := &Config{Debug: false}
			propagateCallDependencies(result, config)

			// Verify the test usages were correctly propagated
			for funcName, expected := range tt.expectedPropagation {
				if expected {
					usages, exists := result.TestUsages[funcName]
					if !exists {
						t.Errorf("Expected function %s to be marked as used in tests", funcName)
						continue
					}
					if len(usages) == 0 {
						t.Errorf("Expected at least one test usage for %s", funcName)
						continue
					}

					// Check the last added test usage is a synthetic one
					lastUsage := usages[len(usages)-1]
					if lastUsage.Pos != token.NoPos {
						t.Errorf("Expected NoPos for synthetic usage, got %v", lastUsage.Pos)
					}
					if lastUsage.FilePath != "" {
						t.Errorf("Expected empty file path for synthetic usage, got %s", lastUsage.FilePath)
					}
					if !lastUsage.IsTest {
						t.Errorf("Expected IsTest to be true for synthetic usage")
					}
				}
			}

			// For functions that aren't in the expected propagation list and aren't already marked in tests before the test,
			// make sure they weren't incorrectly marked
			for funcName, declInfo := range result.Declarations {
				if declInfo.DeclType == DeclFunction || declInfo.DeclType == DeclMethod {
					// Skip functions that were already marked as used in tests before the test
					wasMarkedBefore := false
					if _, ok := result.TestUsages[funcName]; ok && len(result.TestUsages[funcName]) > 0 &&
						result.TestUsages[funcName][0].Pos != token.NoPos {
						wasMarkedBefore = true
					}

					if !wasMarkedBefore {
						// If not in the expected propagation map, it shouldn't be marked
						if !tt.expectedPropagation[funcName] {
							_, marked := result.TestUsages[funcName]
							if marked && len(result.TestUsages[funcName]) > 0 {
								t.Errorf("Function %s should not be marked as used in tests", funcName)
							}
						}
					}
				}
			}
		})
	}
}

func TestAnalyzeFunctionCalls(t *testing.T) {
	// Create a simple AST for testing
	src := `
package example

func main() {
    helper()
    fmt.Println("Hello")
    myStruct.Method()
}

func helper() {
    another()
}

func another() {
    // Do something
}

type myStruct struct{}

func (m *myStruct) Method() {
    // Do something
}
`

	// Parse the source into an AST
	fset := token.NewFileSet()
	file, err := parseSource(src, fset)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Create test analysis result
	result := NewAnalysisResult()

	// Set up declarations
	result.Declarations["helper"] = DeclInfo{DeclType: DeclFunction}
	result.Declarations["another"] = DeclInfo{DeclType: DeclFunction}
	result.Declarations["myStruct"] = DeclInfo{DeclType: DeclTypeDecl}
	result.Declarations["myStruct.Method"] = DeclInfo{DeclType: DeclMethod, IsMethod: true, ReceiverType: "myStruct"}

	// Set up imports
	result.ImportedPkgs["fmt"] = "fmt"

	// Run the function to test
	config := &Config{Debug: false}
	analyzeFunctionCalls(file, result, config)

	// Check if the call graph was populated correctly
	if !containsString(result.CallGraph["main"], "helper") {
		t.Errorf("main should call helper")
	}
	if !containsString(result.CalledBy["helper"], "main") {
		t.Errorf("helper should be called by main")
	}

	if !containsString(result.CallGraph["helper"], "another") {
		t.Errorf("helper should call another")
	}
	if !containsString(result.CalledBy["another"], "helper") {
		t.Errorf("another should be called by helper")
	}

	if !containsString(result.CallGraph["main"], "myStruct.Method") {
		t.Errorf("main should call myStruct.Method")
	}
	if !containsString(result.CalledBy["myStruct.Method"], "main") {
		t.Errorf("myStruct.Method should be called by main")
	}
}

// Helper function to parse Go source code into an AST
func parseSource(src string, fset *token.FileSet) (*ast.File, error) {
	return parser.ParseFile(fset, "example.go", src, 0)
}

// Helper function to check if a string slice contains a given string
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
