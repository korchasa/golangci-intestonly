package intestonly

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestGetFileName(t *testing.T) {
	// Create a new file set
	fset := token.NewFileSet()

	// Parse a simple file to get a valid token.Pos
	src := "package example\n\nfunc main() {}"
	file, err := parser.ParseFile(fset, "example.go", src, 0)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Test getFileName function
	fileName := getFileName(fset, file.Pos())

	if fileName != "example.go" {
		t.Errorf("getFileName() = %s, want example.go", fileName)
	}
}

func TestRecordUsage(t *testing.T) {
	tests := []struct {
		name      string
		identName string
		isTest    bool
	}{
		{
			name:      "record usage in test file",
			identName: "testFunction",
			isTest:    true,
		},
		{
			name:      "record usage in non-test file",
			identName: "prodFunction",
			isTest:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new analysis result
			result := NewAnalysisResult()

			// Create a sample usage
			usage := UsageInfo{
				Pos:      token.Pos(100),
				FilePath: "example.go",
				IsTest:   tt.isTest,
			}

			// Record the usage
			recordUsage(result, tt.identName, usage, tt.isTest)

			// Check if the usage was recorded in the correct map
			if tt.isTest {
				usages, ok := result.TestUsages[tt.identName]
				if !ok {
					t.Errorf("Usage not recorded in TestUsages")
					return
				}
				if len(usages) != 1 {
					t.Errorf("Expected 1 test usage, got %d", len(usages))
					return
				}
				if usages[0].Pos != token.Pos(100) || usages[0].FilePath != "example.go" || !usages[0].IsTest {
					t.Errorf("Test usage recorded incorrectly: %+v", usages[0])
				}

				// Verify no usage in non-test map
				if nonTestUsages, ok := result.Usages[tt.identName]; ok && len(nonTestUsages) > 0 {
					t.Errorf("Unexpected non-test usages: %+v", nonTestUsages)
				}
			} else {
				usages, ok := result.Usages[tt.identName]
				if !ok {
					t.Errorf("Usage not recorded in Usages")
					return
				}
				if len(usages) != 1 {
					t.Errorf("Expected 1 non-test usage, got %d", len(usages))
					return
				}
				if usages[0].Pos != token.Pos(100) || usages[0].FilePath != "example.go" || usages[0].IsTest {
					t.Errorf("Non-test usage recorded incorrectly: %+v", usages[0])
				}

				// Verify no usage in test map
				if testUsages, ok := result.TestUsages[tt.identName]; ok && len(testUsages) > 0 {
					t.Errorf("Unexpected test usages: %+v", testUsages)
				}
			}
		})
	}
}

func TestAnalyzeTypeEmbeddingForFile(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		isTest        bool
		embeddedTypes []string
	}{
		{
			name: "struct with embedded type",
			src: `
package example

type BaseType struct {}
type DerivedType struct {
	BaseType // embedded type
}`,
			isTest:        false,
			embeddedTypes: []string{"BaseType"},
		},
		{
			name: "struct with multiple embedded types",
			src: `
package example

type TypeA struct {}
type TypeB struct {}
type TypeC struct {}
type Combined struct {
	TypeA
	TypeB
	TypeC
	notEmbedded string
}`,
			isTest:        false,
			embeddedTypes: []string{"TypeA", "TypeB", "TypeC"},
		},
		{
			name: "struct with embedded pointer type",
			src: `
package example

type BaseType struct {}
type DerivedType struct {
	*BaseType // embedded pointer type
}`,
			isTest:        false,
			embeddedTypes: []string{"BaseType"},
		},
		{
			name: "struct with imported embedded type",
			src: `
package example

import (
	"fmt"
)

type MyStruct struct {
	fmt.Stringer // embedded interface from another package
}`,
			isTest:        false,
			embeddedTypes: []string{}, // We won't test cross-package embedding here
		},
		{
			name: "struct embedding in test file",
			src: `
package example

type TestBase struct {}
type TestDerived struct {
	TestBase // embedded type in test file
}`,
			isTest:        true,
			embeddedTypes: []string{"TestBase"},
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

			// Create a new analysis result
			result := NewAnalysisResult()

			// Add declarations for embedded types so they can be tracked
			for _, typeName := range tt.embeddedTypes {
				result.Declarations[typeName] = DeclInfo{
					Pos:      token.NoPos,
					Name:     typeName,
					FilePath: "example.go",
					DeclType: DeclTypeDecl,
				}
			}

			// Create a minimal config
			config := &Config{
				Debug: false,
			}

			// Run the embedding analysis
			analyzeTypeEmbeddingForFile(file, fset, tt.isTest, result, config)

			// Check that all expected types were recorded as used
			for _, typeName := range tt.embeddedTypes {
				if tt.isTest {
					usages, ok := result.TestUsages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in test, but wasn't", typeName)
					} else {
						if !usages[0].IsTest {
							t.Errorf("Usage for %s should have IsTest=true, got false", typeName)
						}
					}
				} else {
					usages, ok := result.Usages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in non-test, but wasn't", typeName)
					} else {
						if usages[0].IsTest {
							t.Errorf("Usage for %s should have IsTest=false, got true", typeName)
						}
					}
				}
			}

			// Check that no unexpected types were recorded as used
			for typeName := range result.Declarations {
				isExpected := false
				for _, expected := range tt.embeddedTypes {
					if typeName == expected {
						isExpected = true
						break
					}
				}

				if !isExpected {
					if tt.isTest {
						if usages, ok := result.TestUsages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in test", typeName)
						}
					} else {
						if usages, ok := result.Usages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in non-test", typeName)
						}
					}
				}
			}
		})
	}
}

func TestAnalyzeReflectionUsages(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		isTest   bool
		refTypes []string // Types used in reflection
	}{
		{
			name: "reflect.TypeOf usage",
			src: `
package example

import "reflect"

func useReflection() {
	myVar := MyType{}
	reflect.TypeOf(myVar)
}

type MyType struct{}
`,
			isTest:   false,
			refTypes: []string{"MyType"},
		},
		{
			name: "reflect.ValueOf usage",
			src: `
package example

import "reflect"

func useReflection() {
	myVar := MyType{}
	reflect.ValueOf(myVar)
}

type MyType struct{}
`,
			isTest:   false,
			refTypes: []string{"MyType"},
		},
		{
			name: "multiple reflection usages",
			src: `
package example

import "reflect"

func useReflection() {
	a := TypeA{}
	b := TypeB{}
	reflect.TypeOf(a)
	reflect.ValueOf(b)
}

type TypeA struct{}
type TypeB struct{}
`,
			isTest:   false,
			refTypes: []string{"TypeA", "TypeB"},
		},
		{
			name: "reflection usage in test file",
			src: `
package example

import "reflect"

func useReflection() {
	myVar := TestType{}
	reflect.TypeOf(myVar)
}

type TestType struct{}
`,
			isTest:   true,
			refTypes: []string{"TestType"},
		},
		{
			name: "no reflection usage",
			src: `
package example

import "fmt"

func noReflection() {
	myVar := MyType{}
	fmt.Println(myVar)
}

type MyType struct{}
`,
			isTest:   false,
			refTypes: []string{},
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

			// Create a new analysis result
			result := NewAnalysisResult()

			// Add declarations for reflected types so they can be tracked
			for _, typeName := range tt.refTypes {
				result.Declarations[typeName] = DeclInfo{
					Pos:      token.NoPos,
					Name:     typeName,
					FilePath: "example.go",
					DeclType: DeclTypeDecl,
				}
			}

			// Create a minimal config
			config := &Config{
				Debug:       false,
				CurrentFile: "example.go",
			}

			// Run the reflection analysis
			analyzeReflectionUsages(file, fset, tt.isTest, result, config)

			// Check that all expected types were recorded as used in reflection
			for _, typeName := range tt.refTypes {
				if tt.isTest {
					usages, ok := result.TestUsages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in test reflection, but wasn't", typeName)
					} else {
						if !usages[0].IsTest {
							t.Errorf("Reflection usage for %s should have IsTest=true, got false", typeName)
						}
					}
				} else {
					usages, ok := result.Usages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in non-test reflection, but wasn't", typeName)
					} else {
						if usages[0].IsTest {
							t.Errorf("Reflection usage for %s should have IsTest=false, got true", typeName)
						}
					}
				}
			}

			// Check that no unexpected types were recorded as used
			for typeName := range result.Declarations {
				isExpected := false
				for _, expected := range tt.refTypes {
					if typeName == expected {
						isExpected = true
						break
					}
				}

				if !isExpected {
					if tt.isTest {
						if usages, ok := result.TestUsages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in test reflection", typeName)
						}
					} else {
						if usages, ok := result.Usages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in non-test reflection", typeName)
						}
					}
				}
			}
		})
	}
}

func TestAnalyzeRegistryPatternUsages(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		isTest   bool
		regTypes []string // Types used in registry pattern
	}{
		{
			name: "simple registration function",
			src: `
package example

func Register(item interface{}) {}

func init() {
	Register(MyType{})
}

type MyType struct{}
`,
			isTest:   false,
			regTypes: []string{"MyType"},
		},
		{
			name: "registry with method call",
			src: `
package example

type Registry struct{}

func (r *Registry) Register(item interface{}) {}

func init() {
	reg := &Registry{}
	reg.Register(MyType{})
}

type MyType struct{}
`,
			isTest:   false,
			regTypes: []string{"MyType"},
		},
		{
			name: "multiple registrations",
			src: `
package example

func Register(item interface{}) {}

func init() {
	Register(TypeA{})
	Register(TypeB{})
}

type TypeA struct{}
type TypeB struct{}
`,
			isTest:   false,
			regTypes: []string{"TypeA", "TypeB"},
		},
		{
			name: "registry in test file",
			src: `
package example

func RegisterTest(item interface{}) {}

func TestInit() {
	RegisterTest(TestType{})
}

type TestType struct{}
`,
			isTest:   true,
			regTypes: []string{"TestType"},
		},
		{
			name: "no registry pattern",
			src: `
package example

func noRegistration() {
	myVar := MyType{}
	_ = myVar
}

type MyType struct{}
`,
			isTest:   false,
			regTypes: []string{},
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

			// Create a new analysis result
			result := NewAnalysisResult()

			// Add declarations for registry types so they can be tracked
			for _, typeName := range tt.regTypes {
				result.Declarations[typeName] = DeclInfo{
					Pos:      token.NoPos,
					Name:     typeName,
					FilePath: "example.go",
					DeclType: DeclTypeDecl,
				}
			}

			// Create a minimal config
			config := &Config{
				Debug:       false,
				CurrentFile: "example.go",
			}

			// Run the registry pattern analysis
			analyzeRegistryPatternUsages(file, fset, tt.isTest, result, config)

			// Check that all expected types were recorded as used in registry pattern
			for _, typeName := range tt.regTypes {
				if tt.isTest {
					usages, ok := result.TestUsages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in test registry, but wasn't", typeName)
					} else if !usages[0].IsTest {
						t.Errorf("Registry usage for %s should have IsTest=true, got false", typeName)
					}
				} else {
					usages, ok := result.Usages[typeName]
					if !ok || len(usages) == 0 {
						t.Errorf("Expected type %s to be recorded as used in non-test registry, but wasn't", typeName)
					} else if usages[0].IsTest {
						t.Errorf("Registry usage for %s should have IsTest=false, got true", typeName)
					}
				}
			}

			// Check that no unexpected types were recorded as used
			for typeName := range result.Declarations {
				isExpected := false
				for _, expected := range tt.regTypes {
					if typeName == expected {
						isExpected = true
						break
					}
				}

				if !isExpected {
					if tt.isTest {
						if usages, ok := result.TestUsages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in test registry", typeName)
						}
					} else {
						if usages, ok := result.Usages[typeName]; ok && len(usages) > 0 {
							t.Errorf("Unexpected type %s was recorded as used in non-test registry", typeName)
						}
					}
				}
			}
		})
	}
}
