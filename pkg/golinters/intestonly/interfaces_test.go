package intestonly

import (
	"testing"
)

func TestImplementsInterface(t *testing.T) {
	tests := []struct {
		name             string
		typeName         string
		interfaceMethods []string
		methodsOfType    map[string][]string
		expected         bool
	}{
		{
			name:             "Type implements interface with all methods",
			typeName:         "MyType",
			interfaceMethods: []string{"Method1", "Method2"},
			methodsOfType: map[string][]string{
				"MyType": {"Method1", "Method2", "ExtraMethod"},
			},
			expected: true,
		},
		{
			name:             "Type implements interface with exact methods",
			typeName:         "MyType",
			interfaceMethods: []string{"Method1", "Method2"},
			methodsOfType: map[string][]string{
				"MyType": {"Method1", "Method2"},
			},
			expected: true,
		},
		{
			name:             "Type does not implement interface - missing method",
			typeName:         "MyType",
			interfaceMethods: []string{"Method1", "Method2", "Method3"},
			methodsOfType: map[string][]string{
				"MyType": {"Method1", "Method2"},
			},
			expected: false,
		},
		{
			name:             "Type does not implement interface - no methods",
			typeName:         "MyType",
			interfaceMethods: []string{"Method1"},
			methodsOfType: map[string][]string{
				"MyType": {},
			},
			expected: false,
		},
		{
			name:             "Type does not exist",
			typeName:         "NonExistentType",
			interfaceMethods: []string{"Method1"},
			methodsOfType: map[string][]string{
				"MyType": {"Method1"},
			},
			expected: false,
		},
		{
			name:             "Empty interface is implemented by any type with methods",
			typeName:         "MyType",
			interfaceMethods: []string{},
			methodsOfType: map[string][]string{
				"MyType": {"Method1"},
			},
			expected: true,
		},
		{
			name:             "Empty interface is implemented by type with no methods",
			typeName:         "MyType",
			interfaceMethods: []string{},
			methodsOfType: map[string][]string{
				"MyType": {},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test result
			result := NewAnalysisResult()

			// Set up methods of type
			for typeName, methods := range tt.methodsOfType {
				result.MethodsOfType[typeName] = methods
			}

			// Run the function under test
			actual := implementsInterface(tt.typeName, tt.interfaceMethods, result)

			// Check the result
			if actual != tt.expected {
				t.Errorf("implementsInterface(%s, %v) = %v, want %v",
					tt.typeName, tt.interfaceMethods, actual, tt.expected)
			}
		})
	}
}

func TestMethodExists(t *testing.T) {
	tests := []struct {
		name            string
		qualifiedMethod string
		declarations    map[string]DeclInfo
		methodsOfType   map[string][]string
		expected        bool
	}{
		{
			name:            "Method exists in declarations",
			qualifiedMethod: "MyType.Method1",
			declarations: map[string]DeclInfo{
				"MyType.Method1": {
					Name:     "Method1",
					FilePath: "some_file.go",
					DeclType: DeclMethod,
				},
			},
			methodsOfType: map[string][]string{},
			expected:      true,
		},
		{
			name:            "Method exists in methodsOfType",
			qualifiedMethod: "MyType.Method1",
			declarations:    map[string]DeclInfo{},
			methodsOfType: map[string][]string{
				"MyType": {"Method1", "Method2"},
			},
			expected: true,
		},
		{
			name:            "Method does not exist",
			qualifiedMethod: "MyType.NonExistentMethod",
			declarations: map[string]DeclInfo{
				"MyType.Method1": {
					Name:     "Method1",
					FilePath: "some_file.go",
					DeclType: DeclMethod,
				},
			},
			methodsOfType: map[string][]string{
				"MyType": {"Method1", "Method2"},
			},
			expected: false,
		},
		{
			name:            "Type does not exist",
			qualifiedMethod: "NonExistentType.Method1",
			declarations:    map[string]DeclInfo{},
			methodsOfType: map[string][]string{
				"MyType": {"Method1"},
			},
			expected: false,
		},
		{
			name:            "Invalid qualified method name format",
			qualifiedMethod: "InvalidFormat",
			declarations:    map[string]DeclInfo{},
			methodsOfType:   map[string][]string{},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test result
			result := NewAnalysisResult()

			// Set up declarations
			for name, info := range tt.declarations {
				result.Declarations[name] = info
			}

			// Set up methods of type
			for typeName, methods := range tt.methodsOfType {
				result.MethodsOfType[typeName] = methods
			}

			// Run the function under test
			actual := methodExists(tt.qualifiedMethod, result)

			// Check the result
			if actual != tt.expected {
				t.Errorf("methodExists(%s) = %v, want %v",
					tt.qualifiedMethod, actual, tt.expected)
			}
		})
	}
}