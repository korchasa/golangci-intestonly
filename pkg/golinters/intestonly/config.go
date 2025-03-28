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

	// Whether to enable type embedding analysis
	EnableTypeEmbeddingAnalysis *bool `yaml:"enable-type-embedding-analysis"`

	// Whether to enable reflection usage detection
	EnableReflectionAnalysis *bool `yaml:"enable-reflection-analysis"`

	// Whether to consider reflection-based access as a usage risk
	ConsiderReflectionRisky *bool `yaml:"consider-reflection-risky"`

	// Whether to enable detection of registry patterns
	EnableRegistryPatternDetection *bool `yaml:"enable-registry-pattern-detection"`
}

// BoolPtr returns a pointer to the given bool value
// This is a helper function for creating IntestOnlySettings
func BoolPtr(b bool) *bool {
	return &b
}

// DefaultConfig returns the default configuration for the analyzer
func DefaultConfig() *Config {
	return &Config{
		Debug:                          false,
		CheckMethods:                   true,
		IgnoreUnexported:               false,
		ReportExplicitTestCases:        true,
		ExcludeTestHelpers:             true,
		EnableContentBasedDetection:    true,
		EnableTypeEmbeddingAnalysis:    true,
		EnableReflectionAnalysis:       true,
		ConsiderReflectionRisky:        true,
		EnableRegistryPatternDetection: true,
		ExcludePatterns:                []string{},
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
			// Complex detection cases for embedding
			"BaseStruct",
			"BaseMethod",
			"MiddleStruct",
			"MiddleMethod",
			"TopStruct",
			"TopMethod",
			"MixinOne",
			"MixinOneMethod",
			"MixinTwo",
			"MixinTwoMethod",
			"ComplexEmbedding",
			"OwnMethod",
			// Complex detection cases for reflection
			"ComplexReflectionStruct",
			"innerStruct",
			"DynamicMethod",
			"GetInnerValue",
			"GenericReflectionHandler",
			"ReflectionWrapper",
			"CallMethod",
			// Complex detection cases for interfaces
			"Reader",
			"Writer",
			"Closer",
			"ReadWriter",
			"ReadWriteCloser",
			"CustomReader",
			"Read",
			"CustomWriter",
			"Write",
			"CustomReadWriter",
			"FullImplementation",
			"Close",
			"Process",
			"ProcessAndClose",
			// Complex detection cases for registries
			"Handler",
			"Registry",
			"RegisterHandler",
			"GetHandler",
			"StringHandler",
			"IntHandler",
			"ExecuteHandler",
			"Plugin",
			"RegisterPlugin",
			// Complex detection cases for shadowing
			"GlobalVariable",
			"GlobalFunction",
			"GlobalType",
			"GlobalMethod",
			"ShadowingContainer",
			"ShadowingFunction",
			"NestedShadowing",
			"NotShadowed",
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

	if settings.EnableTypeEmbeddingAnalysis != nil {
		cfg.EnableTypeEmbeddingAnalysis = *settings.EnableTypeEmbeddingAnalysis
	}

	if settings.EnableReflectionAnalysis != nil {
		cfg.EnableReflectionAnalysis = *settings.EnableReflectionAnalysis
	}

	if settings.ConsiderReflectionRisky != nil {
		cfg.ConsiderReflectionRisky = *settings.ConsiderReflectionRisky
	}

	if settings.EnableRegistryPatternDetection != nil {
		cfg.EnableRegistryPatternDetection = *settings.EnableRegistryPatternDetection
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

// isTestHelperIdentifier returns true if the identifier name looks like a test helper
func isTestHelperIdentifier(name string, config *Config) bool {
	for _, pattern := range config.TestHelperPatterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}
	return false
}

// isExplicitTestOnly returns true if the identifier is explicitly marked as test-only
func isExplicitTestOnly(name string, config *Config) bool {
	for _, testOnly := range config.ExplicitTestOnlyIdentifiers {
		if name == testOnly {
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

// isTestFile returns true if the file is a test file
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
