package intestonly

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFileContent(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_file.go")
	
	// Test content with references to declarations
	fileContent := []byte(`
package example

// This file contains references to SomeFunction and AnotherFunction
func main() {
	// Call SomeFunction here
	result := SomeFunction()
	
	// Also use AnotherFunction
	AnotherFunction(result)
}
`)
	
	err := os.WriteFile(tempFile, fileContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create test result and config
	result := NewAnalysisResult()
	config := &Config{Debug: false}
	
	// Add declarations to the result
	result.Declarations["SomeFunction"] = DeclInfo{
		Name:     "SomeFunction",
		FilePath: "some_other_file.go", // Different file to avoid being skipped
		DeclType: DeclFunction,
	}
	
	result.Declarations["AnotherFunction"] = DeclInfo{
		Name:     "AnotherFunction",
		FilePath: "another_file.go", // Different file to avoid being skipped
		DeclType: DeclFunction,
	}
	
	result.Declarations["UnusedFunction"] = DeclInfo{
		Name:     "UnusedFunction",
		FilePath: "unused_file.go",
		DeclType: DeclFunction,
	}
	
	// Run the function under test
	analyzeFileContent(tempFile, result, config)
	
	// Check if the functions were detected in the file content
	if len(result.Usages["SomeFunction"]) == 0 {
		t.Errorf("Expected SomeFunction to be detected in file content, but it wasn't")
	}
	
	if len(result.Usages["AnotherFunction"]) == 0 {
		t.Errorf("Expected AnotherFunction to be detected in file content, but it wasn't")
	}
	
	if len(result.Usages["UnusedFunction"]) > 0 {
		t.Errorf("UnusedFunction should not be detected in file content, but it was")
	}
}

func TestAnalyzeFileContent_SkipsSameFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_file.go")
	
	// Test content with references to declarations
	fileContent := []byte(`
package example

func SameFileFunction() {
	// This function is defined in this file
}

func main() {
	// Call SameFileFunction
	SameFileFunction()
}
`)
	
	err := os.WriteFile(tempFile, fileContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create test result and config
	result := NewAnalysisResult()
	config := &Config{Debug: false}
	
	// Add declaration to the result with the same file path
	result.Declarations["SameFileFunction"] = DeclInfo{
		Name:     "SameFileFunction",
		FilePath: tempFile, // Same file, should be skipped
		DeclType: DeclFunction,
	}
	
	// Run the function under test
	analyzeFileContent(tempFile, result, config)
	
	// Check that the function was not detected as a usage since it's in the same file
	if len(result.Usages["SameFileFunction"]) > 0 {
		t.Errorf("SameFileFunction should not be detected as a usage since it's in the same file")
	}
}

func TestAnalyzeFileContent_HandlesFileReadError(t *testing.T) {
	// Create a non-existent file path
	nonExistentFile := "/path/to/nonexistent/file.go"
	
	// Create test result and config
	result := NewAnalysisResult()
	config := &Config{Debug: false}
	
	// Add a declaration
	result.Declarations["SomeFunction"] = DeclInfo{
		Name:     "SomeFunction",
		FilePath: "some_other_file.go",
		DeclType: DeclFunction,
	}
	
	// Run the function under test with a non-existent file
	// This should not panic or cause an error
	analyzeFileContent(nonExistentFile, result, config)
	
	// No assertions needed - we're just checking that it doesn't panic
}