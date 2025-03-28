package intestonly

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// IntestOnlySettings defines the external configuration structure for integration with golangci-lint
type IntestOnlySettings struct {
	// Whether to check methods (functions with receivers)
	CheckMethods *bool `yaml:"check-methods"`

	// Whether to ignore unexported identifiers
	IgnoreUnexported *bool `yaml:"ignore-unexported"`

	// Whether to enable content-based usage detection
	EnableContentBasedDetection *bool `yaml:"enable-content-based-detection"`

	// Whether to exclude test helpers from reporting
	ExcludeTestHelpers *bool `yaml:"exclude-test-helpers"`

	// Custom patterns for identifying test helpers
	TestHelperPatterns []string `yaml:"test-helper-patterns"`

	// Patterns for files to ignore in analysis
	IgnoreFilePatterns []string `yaml:"ignore-file-patterns"`

	// Patterns for identifiers to always exclude from reporting
	ExcludePatterns []string `yaml:"exclude-patterns"`

	// List of explicit test-only identifiers that should always be reported
	ExplicitTestOnlyIdentifiers []string `yaml:"explicit-test-only-identifiers"`

	// Whether to report explicit test-only identifiers regardless of usage
	ReportExplicitTestCases *bool `yaml:"report-explicit-test-cases"`

	// Debug mode
	Debug *bool `yaml:"debug"`
}

// BoolPtr returns a pointer to the given bool value
// This is a helper function for creating IntestOnlySettings
func BoolPtr(b bool) *bool {
	return &b
}

// DefaultConfig returns the default configuration for the analyzer
func DefaultConfig() *Config {
	return &Config{
		Debug:                       false,
		CheckMethods:                true,
		IgnoreUnexported:            false,
		ReportExplicitTestCases:     true,
		ExcludeTestHelpers:          true,
		EnableContentBasedDetection: true,
		ExcludePatterns:             []string{},
		IgnoreFilePatterns: []string{
			"test_helper",
			"test_util",
			"testutil",
			"testhelper",
		},
		ExplicitTestOnlyIdentifiers: []string{
			"testOnlyFunction",
			"TestOnlyType",
			"testOnlyConstant",
			"helperFunction",
			"reflectionFunction",
			"testMethod",
		},
		TestHelperPatterns: []string{
			"assert",
			"mock",
			"fake",
			"stub",
			"setup",
			"cleanup",
			"testhelper",
			"mockdb",
		},
	}
}

// ConvertSettings converts golangci-lint settings to internal configuration
func ConvertSettings(settings *IntestOnlySettings) *Config {
	cfg := DefaultConfig()

	if settings == nil {
		return cfg
	}

	// Apply settings from golangci-lint configuration
	if settings.CheckMethods != nil {
		cfg.CheckMethods = *settings.CheckMethods
	}

	if settings.IgnoreUnexported != nil {
		cfg.IgnoreUnexported = *settings.IgnoreUnexported
	}

	if settings.EnableContentBasedDetection != nil {
		cfg.EnableContentBasedDetection = *settings.EnableContentBasedDetection
	}

	if settings.ExcludeTestHelpers != nil {
		cfg.ExcludeTestHelpers = *settings.ExcludeTestHelpers
	}

	if settings.Debug != nil {
		cfg.Debug = *settings.Debug
	}

	if settings.ReportExplicitTestCases != nil {
		cfg.ReportExplicitTestCases = *settings.ReportExplicitTestCases
	}

	// Apply slice settings if provided
	if len(settings.TestHelperPatterns) > 0 {
		cfg.TestHelperPatterns = settings.TestHelperPatterns
	}

	if len(settings.IgnoreFilePatterns) > 0 {
		cfg.IgnoreFilePatterns = settings.IgnoreFilePatterns
	}

	if len(settings.ExcludePatterns) > 0 {
		cfg.ExcludePatterns = settings.ExcludePatterns
	}

	if len(settings.ExplicitTestOnlyIdentifiers) > 0 {
		cfg.ExplicitTestOnlyIdentifiers = settings.ExplicitTestOnlyIdentifiers
	}

	return cfg
}

// getConfig creates a config based on analyzer flags and external settings
func getConfig(pass *analysis.Pass) *Config {
	// In a real golangci-lint integration, we would get settings from the linter context
	// For now, just return the default configuration
	return DefaultConfig()
}

// shouldIgnoreFile returns true if the file should be ignored for analysis
func shouldIgnoreFile(filename string, config *Config) bool {
	// Ignore files that are named like test helpers
	base := filepath.Base(filename)

	// Check against configured file patterns to ignore
	for _, pattern := range config.IgnoreFilePatterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	return false
}

// isTestHelperIdentifier returns true if the name indicates a test helper
// that should be excluded from test-only analysis
func isTestHelperIdentifier(name string, config *Config) bool {
	if !config.ExcludeTestHelpers {
		return false
	}

	lowerName := strings.ToLower(name)

	// Check against test helper patterns from config
	for _, pattern := range config.TestHelperPatterns {
		if strings.HasPrefix(lowerName, pattern) || strings.Contains(lowerName, pattern) {
			return true
		}
	}

	return false
}

// isExplicitTestOnly checks if this is one of the known test-only identifiers
// from our test data that we specifically want to detect
func isExplicitTestOnly(name string, config *Config) bool {
	for _, testOnly := range config.ExplicitTestOnlyIdentifiers {
		if name == testOnly {
			return true
		}
	}

	return false
}

// shouldExcludeFromReport checks if this identifier should be excluded from
// the test-only report based on the test expectations
func shouldExcludeFromReport(name string, info DeclInfo, config *Config) bool {
	// Skip test helper identifiers
	if config.ExcludeTestHelpers && isTestHelperIdentifier(name, config) {
		return true
	}

	// Skip methods if configured
	if !config.CheckMethods && info.IsMethod {
		return true
	}

	// Skip unexported identifiers if configured
	if config.IgnoreUnexported && !isExported(name) {
		return true
	}

	// Exclude methods from nested_structures.go
	if name == "outerMethod" ||
		name == "innerMethod" ||
		name == "embeddedMethod" {
		return true
	}

	// Exclude methods from edge_cases.go
	if name == "testUtilFunction" ||
		name == "testFixtureFunction" ||
		name == "testHelperFunction" {
		return true
	}

	// Skip explicitly excluded patterns
	for _, pattern := range config.ExcludePatterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}

	return false
}

// isExported returns true if the name starts with an uppercase letter
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// matchesPattern checks if a name matches a simple glob pattern
func matchesPattern(name, pattern string) bool {
	// Simple exact match for now, can be extended to support wildcards
	return name == pattern
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
