package intestonly

import (
	"strings"
	"testing"
)

func TestConvertSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *IntestOnlySettings
		want     *Config
	}{
		{
			name:     "nil settings should return default config",
			settings: nil,
			want:     DefaultConfig(),
		},
		{
			name: "debug setting should be respected",
			settings: &IntestOnlySettings{
				Debug: BoolPtr(true),
			},
			want: func() *Config {
				config := DefaultConfig()
				config.Debug = true
				return config
			}(),
		},
		{
			name: "file pattern settings should be respected",
			settings: &IntestOnlySettings{
				OverrideIsCodeFiles: []string{"mock_", "fake_"},
				OverrideIsTestFiles: []string{"*_fixture.go"},
			},
			want: func() *Config {
				config := DefaultConfig()
				config.OverrideIsCodeFiles = []string{"mock_", "fake_"}
				config.OverrideIsTestFiles = []string{"*_fixture.go"}
				return config
			}(),
		},
		{
			name: "all settings combined should be respected",
			settings: &IntestOnlySettings{
				Debug:               BoolPtr(true),
				OverrideIsCodeFiles: []string{"mock_", "fake_"},
				OverrideIsTestFiles: []string{"*_fixture.go"},
			},
			want: func() *Config {
				config := DefaultConfig()
				config.Debug = true
				config.OverrideIsCodeFiles = []string{"mock_", "fake_"}
				config.OverrideIsTestFiles = []string{"*_fixture.go"}
				return config
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSettings(tt.settings)

			// Check user-configurable options
			if got.Debug != tt.want.Debug {
				t.Errorf("Debug = %v, want %v", got.Debug, tt.want.Debug)
			}

			// Check slice equality for user-configurable options
			assertStringSliceEqual(t, "OverrideIsCodeFiles", got.OverrideIsCodeFiles, tt.want.OverrideIsCodeFiles)
			assertStringSliceEqual(t, "OverrideIsTestFiles", got.OverrideIsTestFiles, tt.want.OverrideIsTestFiles)

			// Check that hardcoded options have the expected default values
			if got.CheckMethods != tt.want.CheckMethods {
				t.Errorf("CheckMethods = %v, want %v", got.CheckMethods, tt.want.CheckMethods)
			}
			if got.IgnoreUnexported != tt.want.IgnoreUnexported {
				t.Errorf("IgnoreUnexported = %v, want %v", got.IgnoreUnexported, tt.want.IgnoreUnexported)
			}
			// ... and so on for other hardcoded options
		})
	}
}

func assertStringSliceEqual(t *testing.T, name string, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("%s length = %d, want %d", name, len(got), len(want))
		return
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %s, want %s", name, i, got[i], want[i])
		}
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  bool
	}{
		{
			name:  "true value",
			value: true,
			want:  true,
		},
		{
			name:  "false value",
			value: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := BoolPtr(tt.value)
			if ptr == nil {
				t.Fatal("Expected non-nil pointer, got nil")
			}
			if *ptr != tt.want {
				t.Errorf("BoolPtr(%v) = %v, want %v", tt.value, *ptr, tt.want)
			}
		})
	}
}

func TestShouldIgnoreFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		config   *Config
		want     bool
	}{
		{
			name:     "should ignore files matching pattern",
			filename: "path/to/mock_user.go",
			config: &Config{
				OverrideIsCodeFiles: []string{"mock_"},
			},
			want: true,
		},
		{
			name:     "should ignore files matching pattern in middle of name",
			filename: "path/to/some_mock_file.go",
			config: &Config{
				OverrideIsCodeFiles: []string{"mock"},
			},
			want: true,
		},
		{
			name:     "should not ignore files not matching any pattern",
			filename: "path/to/user.go",
			config: &Config{
				OverrideIsCodeFiles: []string{"mock_", "test_", "fake_"},
			},
			want: false,
		},
		{
			name:     "should handle empty patterns",
			filename: "path/to/file.go",
			config: &Config{
				OverrideIsCodeFiles: []string{},
			},
			want: false,
		},
		{
			name:     "should handle empty filename",
			filename: "",
			config: &Config{
				OverrideIsCodeFiles: []string{"mock_"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIgnoreFile(tt.filename, tt.config)
			if got != tt.want {
				t.Errorf("shouldIgnoreFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsTestHelperIdentifier(t *testing.T) {
	tests := []struct {
		name   string
		ident  string
		config *Config
		want   bool
	}{
		{
			name:   "should identify Test prefix",
			ident:  "TestHelper",
			config: &Config{TestHelperPatterns: []string{"helper"}},
			want:   true,
		},
		{
			name:   "should identify test prefix (lowercase)",
			ident:  "testFunction",
			config: &Config{TestHelperPatterns: []string{"helper"}},
			want:   true,
		},
		{
			name:   "should identify Test suffix",
			ident:  "RunnerTest",
			config: &Config{TestHelperPatterns: []string{"helper"}},
			want:   true,
		},
		{
			name:   "should identify test suffix (lowercase)",
			ident:  "helpertest",
			config: &Config{TestHelperPatterns: []string{"helper"}},
			want:   true,
		},
		{
			name:   "should match custom pattern",
			ident:  "MockServer",
			config: &Config{TestHelperPatterns: []string{"mock", "fake", "stub"}},
			want:   true,
		},
		{
			name:   "should be case insensitive for patterns",
			ident:  "mockFunction",
			config: &Config{TestHelperPatterns: []string{"MOCK"}},
			want:   true,
		},
		{
			name:   "should not match if no patterns are matched",
			ident:  "ProductionCode",
			config: &Config{TestHelperPatterns: []string{"mock", "fake", "stub"}},
			want:   false,
		},
		{
			name:   "should handle empty patterns",
			ident:  "SomeIdentifier",
			config: &Config{TestHelperPatterns: []string{}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTestHelperIdentifier(tt.ident, tt.config)
			if got != tt.want {
				t.Errorf("isTestHelperIdentifier(%s) = %v, want %v", tt.ident, got, tt.want)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		config   *Config
		want     bool
	}{
		{
			name:     "should identify standard test file",
			filename: "user_test.go",
			config:   &Config{OverrideIsTestFiles: []string{}},
			want:     true,
		},
		{
			name:     "should not identify production file",
			filename: "user.go",
			config:   &Config{OverrideIsTestFiles: []string{}},
			want:     false,
		},
		{
			name:     "should identify additional test pattern",
			filename: "user_fixture.go",
			config:   &Config{OverrideIsTestFiles: []string{"*_fixture.go"}},
			want:     true,
		},
		{
			name:     "should identify additional test pattern with full path",
			filename: "/path/to/testdata/file.go",
			config:   &Config{OverrideIsTestFiles: []string{"testdata/"}},
			want:     true,
		},
		{
			name:     "should not identify if not matching any pattern",
			filename: "helper.go",
			config:   &Config{OverrideIsTestFiles: []string{"*_test.go", "*_fixture.go"}},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTestFile(tt.filename, tt.config)
			if got != tt.want {
				t.Errorf("isTestFile(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name    string
		ident   string
		pattern string
		want    bool
	}{
		{
			name:    "exact match",
			ident:   "TestName",
			pattern: "TestName",
			want:    true,
		},
		{
			name:    "case insensitive match",
			ident:   "TestName",
			pattern: "testname",
			want:    true,
		},
		{
			name:    "substring match",
			ident:   "TestHelperFunction",
			pattern: "Helper",
			want:    true,
		},
		{
			name:    "no match",
			ident:   "ProductionCode",
			pattern: "Test",
			want:    false,
		},
		{
			name:    "empty pattern",
			ident:   "SomeIdentifier",
			pattern: "",
			want:    false,
		},
		{
			name:    "empty identifier",
			ident:   "",
			pattern: "Test",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPattern(tt.ident, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%s, %s) = %v, want %v",
					tt.ident, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestDefaultTestHelperPatterns(t *testing.T) {
	patterns := defaultTestHelperPatterns()

	// Test that some expected patterns are included
	expectedPatterns := []string{"mock", "stub", "fake", "test", "fixture"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if strings.ToLower(pattern) == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern '%s' not found in default patterns", expected)
		}
	}

	// Test that there are a reasonable number of patterns
	if len(patterns) < 5 {
		t.Errorf("Expected at least 5 default test helper patterns, got %d", len(patterns))
	}
}

func TestDefaultIgnoreFilePatterns(t *testing.T) {
	patterns := defaultIgnoreFilePatterns()

	// Test that some expected patterns are included
	expectedPatterns := []string{"test_helper", "test_util", "testutil"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(pattern), expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern containing '%s' not found in default patterns", expected)
		}
	}

	// Test that there are a reasonable number of patterns
	if len(patterns) < 3 {
		t.Errorf("Expected at least 3 default override-is-code-files patterns, got %d", len(patterns))
	}
}
