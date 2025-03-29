package intestonly

import (
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestNewAnalysisResult(t *testing.T) {
	// Call the function under test
	result := NewAnalysisResult()

	// Check that all maps are initialized
	if result.Declarations == nil {
		t.Error("Declarations map should be initialized, but it's nil")
	}

	if result.TestUsages == nil {
		t.Error("TestUsages map should be initialized, but it's nil")
	}

	if result.Usages == nil {
		t.Error("Usages map should be initialized, but it's nil")
	}

	if result.DeclPositions == nil {
		t.Error("DeclPositions map should be initialized, but it's nil")
	}

	if result.ImportRefs == nil {
		t.Error("ImportRefs map should be initialized, but it's nil")
	}

	if result.ImportedPkgs == nil {
		t.Error("ImportedPkgs map should be initialized, but it's nil")
	}

	if result.CallGraph == nil {
		t.Error("CallGraph map should be initialized, but it's nil")
	}

	if result.CalledBy == nil {
		t.Error("CalledBy map should be initialized, but it's nil")
	}

	if result.Interfaces == nil {
		t.Error("Interfaces map should be initialized, but it's nil")
	}

	if result.Implementations == nil {
		t.Error("Implementations map should be initialized, but it's nil")
	}

	if result.MethodsOfType == nil {
		t.Error("MethodsOfType map should be initialized, but it's nil")
	}

	if result.ExportedDecls == nil {
		t.Error("ExportedDecls map should be initialized, but it's nil")
	}

	if result.CrossPackageTestRefs == nil {
		t.Error("CrossPackageTestRefs map should be initialized, but it's nil")
	}

	if result.CrossPackageRefs == nil {
		t.Error("CrossPackageRefs map should be initialized, but it's nil")
	}

	// Check that all maps are empty
	if len(result.Declarations) != 0 {
		t.Errorf("Declarations map should be empty, but it has %d elements", len(result.Declarations))
	}

	if len(result.TestUsages) != 0 {
		t.Errorf("TestUsages map should be empty, but it has %d elements", len(result.TestUsages))
	}

	if len(result.Usages) != 0 {
		t.Errorf("Usages map should be empty, but it has %d elements", len(result.Usages))
	}

	if len(result.DeclPositions) != 0 {
		t.Errorf("DeclPositions map should be empty, but it has %d elements", len(result.DeclPositions))
	}

	if len(result.ImportRefs) != 0 {
		t.Errorf("ImportRefs map should be empty, but it has %d elements", len(result.ImportRefs))
	}

	if len(result.ImportedPkgs) != 0 {
		t.Errorf("ImportedPkgs map should be empty, but it has %d elements", len(result.ImportedPkgs))
	}

	if len(result.CallGraph) != 0 {
		t.Errorf("CallGraph map should be empty, but it has %d elements", len(result.CallGraph))
	}

	if len(result.CalledBy) != 0 {
		t.Errorf("CalledBy map should be empty, but it has %d elements", len(result.CalledBy))
	}

	if len(result.Interfaces) != 0 {
		t.Errorf("Interfaces map should be empty, but it has %d elements", len(result.Interfaces))
	}

	if len(result.Implementations) != 0 {
		t.Errorf("Implementations map should be empty, but it has %d elements", len(result.Implementations))
	}

	if len(result.MethodsOfType) != 0 {
		t.Errorf("MethodsOfType map should be empty, but it has %d elements", len(result.MethodsOfType))
	}

	if len(result.ExportedDecls) != 0 {
		t.Errorf("ExportedDecls map should be empty, but it has %d elements", len(result.ExportedDecls))
	}

	if len(result.CrossPackageTestRefs) != 0 {
		t.Errorf("CrossPackageTestRefs map should be empty, but it has %d elements", len(result.CrossPackageTestRefs))
	}

	if len(result.CrossPackageRefs) != 0 {
		t.Errorf("CrossPackageRefs map should be empty, but it has %d elements", len(result.CrossPackageRefs))
	}
}

func TestIssueToAnalysisIssue(t *testing.T) {
	// Create a test issue
	issue := &Issue{
		Pos:     token.Pos(123),
		Message: "Test issue message",
	}

	// Create a mock pass
	pass := &analysis.Pass{
		// We don't need to set any fields for this test
	}

	// Call the function under test
	diagnostic := issue.ToAnalysisIssue(pass)

	// Check the result
	if diagnostic.Pos != token.Pos(123) {
		t.Errorf("Expected Pos to be 123, but got %v", diagnostic.Pos)
	}

	if diagnostic.Message != "Test issue message" {
		t.Errorf("Expected Message to be 'Test issue message', but got '%s'", diagnostic.Message)
	}

	if diagnostic.Category != "intestonly" {
		t.Errorf("Expected Category to be 'intestonly', but got '%s'", diagnostic.Category)
	}
}