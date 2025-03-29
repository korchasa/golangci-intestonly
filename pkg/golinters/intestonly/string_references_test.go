package intestonly

import (
	"testing"
)

func TestFindFunctionReferencesInString(t *testing.T) {
	tests := []struct {
		name           string
		str            string
		declarations   map[string]DeclInfo
		importRefs     map[string]string
		isTest         bool
		expectedUsages map[string]bool
	}{
		{
			name: "Function call pattern with parentheses",
			str:  "Call SomeFunction() to get results",
			declarations: map[string]DeclInfo{
				"SomeFunction": {
					Name:     "SomeFunction",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
				"UnusedFunction": {
					Name:     "UnusedFunction",
					FilePath: "unused_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{"SomeFunction": true},
		},
		{
			name: "Function call pattern with space before parentheses",
			str:  "Call SomeFunction () to get results",
			declarations: map[string]DeclInfo{
				"SomeFunction": {
					Name:     "SomeFunction",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{"SomeFunction": true},
		},
		{
			name: "Function call pattern with arguments",
			str:  "Call SomeFunction(arg1, arg2) to get results",
			declarations: map[string]DeclInfo{
				"SomeFunction": {
					Name:     "SomeFunction",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{"SomeFunction": true},
		},
		{
			name: "Function call pattern with space before arguments",
			str:  "Call SomeFunction (arg1, arg2) to get results",
			declarations: map[string]DeclInfo{
				"SomeFunction": {
					Name:     "SomeFunction",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{"SomeFunction": true},
		},
		{
			name: "No function call pattern",
			str:  "This string doesn't contain any function calls",
			declarations: map[string]DeclInfo{
				"SomeFunction": {
					Name:     "SomeFunction",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{},
		},
		{
			name: "Short name skipped",
			str:  "Call Fn() to get results",
			declarations: map[string]DeclInfo{
				"Fn": {
					Name:     "Fn",
					FilePath: "some_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         false,
			expectedUsages: map[string]bool{},
		},
		{
			name: "Test file usage",
			str:  "Call TestFunction() in tests",
			declarations: map[string]DeclInfo{
				"TestFunction": {
					Name:     "TestFunction",
					FilePath: "test_file.go",
					DeclType: DeclFunction,
				},
			},
			importRefs:     map[string]string{},
			isTest:         true,
			expectedUsages: map[string]bool{"TestFunction": true},
		},
		{
			name: "Imported function reference",
			str:  "Call pkg.ImportedFunction() to use imported function",
			declarations: map[string]DeclInfo{
				"ImportedFunction": {
					Name:      "ImportedFunction",
					FilePath:  "imported_file.go",
					DeclType:  DeclFunction,
					ImportRef: "pkg.ImportedFunction",
				},
			},
			importRefs: map[string]string{
				"pkg.ImportedFunction": "ImportedFunction",
			},
			isTest:         false,
			expectedUsages: map[string]bool{"ImportedFunction": true},
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

			// Add import references
			for ref, name := range tt.importRefs {
				result.ImportRefs[ref] = name
			}

			// Run the function under test
			findFunctionReferencesInString(tt.str, result, "test_file.go", tt.isTest, config)

			// Check if the expected usages were detected
			for name, expected := range tt.expectedUsages {
				if expected {
					if tt.isTest {
						if len(result.TestUsages[name]) == 0 {
							t.Errorf("Expected %s to be detected in test string, but it wasn't", name)
						}
					} else {
						if len(result.Usages[name]) == 0 {
							t.Errorf("Expected %s to be detected in string, but it wasn't", name)
						}
					}
				}
			}

			// Check that unexpected usages were not detected
			for name := range tt.declarations {
				if !tt.expectedUsages[name] {
					if tt.isTest {
						if len(result.TestUsages[name]) > 0 {
							t.Errorf("%s should not be detected in test string, but it was", name)
						}
					} else {
						if len(result.Usages[name]) > 0 {
							t.Errorf("%s should not be detected in string, but it was", name)
						}
					}
				}
			}
		})
	}
}
