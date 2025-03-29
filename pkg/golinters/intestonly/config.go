package intestonly

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// IntestOnlySettings defines the external configuration structure for integration with golangci-lint
type IntestOnlySettings struct {
	// Debug mode - enables output of debug information
	Debug *bool `yaml:"debug"`

	// Patterns for files to override as code files - these files will definitely not be considered as test files
	OverrideIsCodeFiles []string `yaml:"override-is-code-files"`

	// Patterns for files to override as test files - these files will definitely be considered as test files
	OverrideIsTestFiles []string `yaml:"override-is-test-files"`
}

// BoolPtr returns a pointer to the given bool value
// This is a helper function for creating IntestOnlySettings
func BoolPtr(b bool) *bool {
	return &b
}

// DefaultConfig returns the default configuration for the analyzer
func DefaultConfig() *Config {
	config := &Config{
		// User-configurable options
		Debug:              false,
		OverrideIsCodeFiles: defaultIgnoreFilePatterns(),
		OverrideIsTestFiles:    []string{},

		// Hardcoded options (not configurable by users)
		CheckMethods:                           true,
		IgnoreUnexported:                       false,
		EnableContentBasedDetection:            true,
		ExcludeTestHelpers:                     true,
		TestHelperPatterns:                     defaultTestHelperPatterns(),
		ExcludePatterns:                        []string{},
		ExplicitTestOnlyIdentifiers:            []string{},
		ReportExplicitTestCases:                false,
		EnableTypeEmbeddingAnalysis:            true,
		EnableReflectionAnalysis:               true,
		ConsiderReflectionRisky:                true,
		EnableRegistryPatternDetection:         true,
		EnableCallGraphAnalysis:                true,
		EnableInterfaceImplementationDetection: true,
		EnableRobustCrossPackageAnalysis:       true,
		EnableExportedIdentifierHandling:       true,
		ConsiderExportedConstantsUsed:          true,
		IgnoreFiles:                            []string{},
		IgnoreDirectories:                      []string{},
		ExplicitTestCases:                      []string{},
		IgnoreDirPatterns:                      []string{},
	}
	return config
}

// ConvertSettings converts golangci-lint settings to internal configuration
func ConvertSettings(settings *IntestOnlySettings) *Config {
	if settings == nil {
		return DefaultConfig()
	}

	config := DefaultConfig()

	// Convert only the user-configurable options
	if settings.Debug != nil {
		config.Debug = *settings.Debug
	}
	if settings.OverrideIsCodeFiles != nil {
		config.OverrideIsCodeFiles = settings.OverrideIsCodeFiles
	}
	if settings.OverrideIsTestFiles != nil {
		config.OverrideIsTestFiles = settings.OverrideIsTestFiles
	}

	return config
}

// getConfig creates a config based on analyzer flags and external settings
func getConfig(pass *analysis.Pass) *Config {
	// In a real golangci-lint integration, we would get settings from the linter context
	// For now, just return the default configuration with some modifications for tests
	config := DefaultConfig()

	// Disable ConsiderExportedConstantsUsed for tests to ensure we correctly identify
	// constants that are only used in tests
	config.ConsiderExportedConstantsUsed = false

	return config
}

// shouldIgnoreFile returns true if the file should be ignored for analysis
func shouldIgnoreFile(filename string, config *Config) bool {
	// Ignore files that are named like test helpers
	base := filepath.Base(filename)

	// Check against configured file patterns to override as code files
	for _, pattern := range config.OverrideIsCodeFiles {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	return false
}

// isTestHelperIdentifier returns true if the identifier name looks like a test helper
func isTestHelperIdentifier(name string, config *Config) bool {
	// Check for specific test prefixes
	if strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "test") {
		return true
	}

	// Check for specific test suffixes
	if strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "test") {
		return true
	}

	// Check against configured patterns
	lowerName := strings.ToLower(name)
	for _, pattern := range config.TestHelperPatterns {
		if strings.Contains(lowerName, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// shouldExcludeFromReport determines if an identifier should be excluded from reporting
func shouldExcludeFromReport(name string, info DeclInfo, config *Config) bool {
	// Always exclude specific patterns if configured
	for _, pattern := range config.ExcludePatterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}

	// Skip methods if configured to do so
	if info.IsMethod && !config.CheckMethods {
		return true
	}

	// Skip unexported identifiers if configured to do so
	if !isExported(name) && config.IgnoreUnexported {
		return true
	}

	// Skip test helpers if configured to do so
	if config.ExcludeTestHelpers && isTestHelperIdentifier(name, config) {
		return true
	}

	// Check if reflection usage is risky and this identifier might be used via reflection
	if config.ConsiderReflectionRisky {
		// These are heuristics for identifiers that might be used via reflection:
		// - Exported methods of struct types (might be called via reflection)
		// - Fields of struct types (might be accessed via reflection)
		// - Types that might be instantiated via reflection
		if isExported(name) &&
			(info.IsMethod || strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "Set") ||
				strings.HasSuffix(name, "Type") || strings.HasSuffix(name, "Handler")) {
			return true
		}
	}

	return false
}

// isExported returns true if the identifier is exported
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// matchesPattern checks if name contains the given pattern
func matchesPattern(name, pattern string) bool {
	return strings.Contains(strings.ToLower(name), strings.ToLower(pattern))
}

// isTestFile returns true if the file is a test file based on its base name ending with '_test.go'
func isTestFile(filename string, config *Config) bool {
	base := filepath.Base(filename)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// Check patterns for files to override as test files from configuration
	if config != nil {
		for _, pattern := range config.OverrideIsTestFiles {
			if matchWildcard(filename, pattern) || matchWildcard(base, pattern) {
				return true
			}
		}
	}

	return false
}

// matchWildcard performs simple wildcard matching.
// It supports the '*' character to match any sequence of characters.
func matchWildcard(s, pattern string) bool {
	// If the pattern doesn't contain a wildcard, use simple contains check
	if !strings.Contains(pattern, "*") {
		return strings.Contains(s, pattern)
	}

	// Split the pattern by '*' to get parts
	parts := strings.Split(pattern, "*")

	// Special case: pattern starts with '*'
	if pattern[0] == '*' {
		// If pattern is just "*", it matches everything
		if len(parts) == 2 && parts[1] == "" {
			return true
		}

		// Check if the string ends with the part after '*'
		return strings.HasSuffix(s, parts[1])
	}

	// Special case: pattern ends with '*'
	if pattern[len(pattern)-1] == '*' {
		// Check if the string starts with the part before '*'
		return strings.HasPrefix(s, parts[0])
	}

	// Pattern has '*' in the middle
	// Check if the string starts with the first part and ends with the last part
	return strings.HasPrefix(s, parts[0]) && strings.HasSuffix(s, parts[len(parts)-1])
}

// defaultTestHelperPatterns returns the default patterns for identifying test helpers
func defaultTestHelperPatterns() []string {
	return []string{
		"assert",
		"mock",
		"fake",
		"stub",
		"setup",
		"cleanup",
		"testhelper",
		"mockdb",
		"helper",
		"fixture",
		"util",
		"wrapper",
		"prepare",
		"test",
		"env",
		"equal",
		"expectation",
		"harness",
	}
}

// defaultIgnoreFilePatterns returns the default patterns for files to override as code files
func defaultIgnoreFilePatterns() []string {
	return []string{
		"test_helper",
		"test_util",
		"testutil",
		"testhelper",
	}
}
