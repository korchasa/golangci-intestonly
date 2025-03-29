package intestonly

import (
	"testing"
)

func TestProcessCrossPackageTestReferences(t *testing.T) {
	tests := []struct {
		name                 string
		declarations         map[string]DeclInfo
		crossPackageTestRefs map[string]bool
		crossPackageRefs     map[string]bool
		expectedTestUsages   map[string]bool
		expectedUsages       map[string]bool
	}{
		{
			name: "Declaration referenced in test file in another package",
			declarations: map[string]DeclInfo{
				"TestOnlyFunc": {
					Name:      "TestOnlyFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.TestOnlyFunc",
				},
			},
			crossPackageTestRefs: map[string]bool{
				"pkg.TestOnlyFunc": true,
			},
			crossPackageRefs: map[string]bool{},
			expectedTestUsages: map[string]bool{
				"TestOnlyFunc": true,
			},
			expectedUsages: map[string]bool{},
		},
		{
			name: "Declaration referenced in non-test file in another package",
			declarations: map[string]DeclInfo{
				"ProdFunc": {
					Name:      "ProdFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.ProdFunc",
				},
			},
			crossPackageTestRefs: map[string]bool{},
			crossPackageRefs: map[string]bool{
				"pkg.ProdFunc": true,
			},
			expectedTestUsages: map[string]bool{},
			expectedUsages: map[string]bool{
				"ProdFunc": true,
			},
		},
		{
			name: "Declaration referenced in both test and non-test files in another package",
			declarations: map[string]DeclInfo{
				"BothFunc": {
					Name:      "BothFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.BothFunc",
				},
			},
			crossPackageTestRefs: map[string]bool{
				"pkg.BothFunc": true,
			},
			crossPackageRefs: map[string]bool{
				"pkg.BothFunc": true,
			},
			expectedTestUsages: map[string]bool{
				"BothFunc": true,
			},
			expectedUsages: map[string]bool{
				"BothFunc": true,
			},
		},
		{
			name: "Declaration not referenced in any other package",
			declarations: map[string]DeclInfo{
				"UnusedFunc": {
					Name:      "UnusedFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.UnusedFunc",
				},
			},
			crossPackageTestRefs: map[string]bool{},
			crossPackageRefs:     map[string]bool{},
			expectedTestUsages:   map[string]bool{},
			expectedUsages:       map[string]bool{},
		},
		{
			name: "Declaration without import reference",
			declarations: map[string]DeclInfo{
				"LocalFunc": {
					Name:     "LocalFunc",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
					// No ImportRef
				},
			},
			crossPackageTestRefs: map[string]bool{
				"pkg.SomeOtherFunc": true,
			},
			crossPackageRefs: map[string]bool{
				"pkg.SomeOtherFunc": true,
			},
			expectedTestUsages: map[string]bool{},
			expectedUsages:     map[string]bool{},
		},
		{
			name: "Multiple declarations with different references",
			declarations: map[string]DeclInfo{
				"TestFunc": {
					Name:      "TestFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.TestFunc",
				},
				"ProdFunc": {
					Name:      "ProdFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.ProdFunc",
				},
				"UnusedFunc": {
					Name:      "UnusedFunc",
					FilePath:  "some_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.UnusedFunc",
				},
			},
			crossPackageTestRefs: map[string]bool{
				"pkg.TestFunc": true,
			},
			crossPackageRefs: map[string]bool{
				"pkg.ProdFunc": true,
			},
			expectedTestUsages: map[string]bool{
				"TestFunc": true,
			},
			expectedUsages: map[string]bool{
				"ProdFunc": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test result and config
			result := NewAnalysisResult()
			config := &Config{Debug: false}

			// Add declarations to the result
			for name, info := range tt.declarations {
				result.Declarations[name] = info
			}

			// Set up cross-package references
			for ref, val := range tt.crossPackageTestRefs {
				result.CrossPackageTestRefs[ref] = val
			}
			for ref, val := range tt.crossPackageRefs {
				result.CrossPackageRefs[ref] = val
			}

			// Run the function under test
			processCrossPackageTestReferences(result, config)

			// Check test usages
			for name, expected := range tt.expectedTestUsages {
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
			for name, expected := range tt.expectedUsages {
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