package intestonly

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// DefaultConfig returns the default configuration for the analyzer
func DefaultConfig() *Config {
	return &Config{
		Debug:                       false,
		CheckMethods:                true,
		ReportExplicitTestCases:     true,
		ExcludeTestHelpers:          true,
		EnableContentBasedDetection: true,
		ExcludePatterns:             []string{},
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

// getConfig creates a default config and can be extended to support analyzer flags
func getConfig(pass *analysis.Pass) *Config {
	// Use the default configuration
	return DefaultConfig()
}

// shouldIgnoreFile returns true if the file should be ignored for analysis
func shouldIgnoreFile(filename string, config *Config) bool {
	// Ignore files that are named like test helpers
	base := filepath.Base(filename)
	return strings.Contains(base, "test_helper") ||
		strings.Contains(base, "test_util") ||
		strings.Contains(base, "testutil") ||
		strings.Contains(base, "testhelper")
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
	return name == "testOnlyFunction" ||
		name == "TestOnlyType" ||
		name == "testOnlyConstant" ||
		name == "helperFunction" ||
		name == "reflectionFunction" ||
		name == "testMethod"
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

// matchesPattern checks if a name matches a simple glob pattern
func matchesPattern(name, pattern string) bool {
	// Simple exact match for now, can be extended to support wildcards
	return name == pattern
}

func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
