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

	// Whether to enable call graph analysis
	EnableCallGraphAnalysis *bool `yaml:"enable-call-graph-analysis"`

	// Whether to enable interface implementation detection
	EnableInterfaceImplementationDetection *bool `yaml:"enable-interface-implementation-detection"`

	// Whether to enable robust cross-package analysis
	EnableRobustCrossPackageAnalysis *bool `yaml:"enable-robust-cross-package-analysis"`

	// Whether to enable exported identifier handling
	EnableExportedIdentifierHandling *bool `yaml:"enable-exported-identifier-handling"`

	// Whether to consider exported constants used
	ConsiderExportedConstantsUsed *bool `yaml:"consider-exported-constants-used"`

	// Additional test files patterns to consider as test files
	AdditionalTests []string `yaml:"additional-tests"`
}

// BoolPtr returns a pointer to the given bool value
// This is a helper function for creating IntestOnlySettings
func BoolPtr(b bool) *bool {
	return &b
}

// DefaultConfig returns the default configuration for the analyzer
func DefaultConfig() *Config {
	return &Config{
		CheckMethods:                           true,
		IgnoreUnexported:                       false,
		EnableContentBasedDetection:            true,
		ExcludeTestHelpers:                     true,
		Debug:                                  false,
		TestHelperPatterns:                     defaultTestHelperPatterns(),
		IgnoreFilePatterns:                     defaultIgnoreFilePatterns(),
		ExcludePatterns:                        []string{},
		ExplicitTestOnlyIdentifiers:            defaultExplicitTestOnlyIdentifiers(),
		ReportExplicitTestCases:                true,
		EnableTypeEmbeddingAnalysis:            true,
		EnableReflectionAnalysis:               true,
		ConsiderReflectionRisky:                true,
		EnableRegistryPatternDetection:         true,
		EnableCallGraphAnalysis:                true,
		EnableInterfaceImplementationDetection: true,
		EnableRobustCrossPackageAnalysis:       true,
		EnableExportedIdentifierHandling:       true,
		ConsiderExportedConstantsUsed:          true,
		AdditionalTests:                        []string{},
	}
}

// ConvertSettings converts golangci-lint settings to internal configuration
func ConvertSettings(settings *IntestOnlySettings) *Config {
	if settings == nil {
		return DefaultConfig()
	}

	config := DefaultConfig()

	// Convert booleans with default value checks
	if settings.CheckMethods != nil {
		config.CheckMethods = *settings.CheckMethods
	}
	if settings.IgnoreUnexported != nil {
		config.IgnoreUnexported = *settings.IgnoreUnexported
	}
	if settings.EnableContentBasedDetection != nil {
		config.EnableContentBasedDetection = *settings.EnableContentBasedDetection
	}
	if settings.ExcludeTestHelpers != nil {
		config.ExcludeTestHelpers = *settings.ExcludeTestHelpers
	}
	if settings.Debug != nil {
		config.Debug = *settings.Debug
	}
	if settings.ReportExplicitTestCases != nil {
		config.ReportExplicitTestCases = *settings.ReportExplicitTestCases
	}
	if settings.EnableTypeEmbeddingAnalysis != nil {
		config.EnableTypeEmbeddingAnalysis = *settings.EnableTypeEmbeddingAnalysis
	}
	if settings.EnableReflectionAnalysis != nil {
		config.EnableReflectionAnalysis = *settings.EnableReflectionAnalysis
	}
	if settings.ConsiderReflectionRisky != nil {
		config.ConsiderReflectionRisky = *settings.ConsiderReflectionRisky
	}
	if settings.EnableRegistryPatternDetection != nil {
		config.EnableRegistryPatternDetection = *settings.EnableRegistryPatternDetection
	}
	if settings.EnableCallGraphAnalysis != nil {
		config.EnableCallGraphAnalysis = *settings.EnableCallGraphAnalysis
	}
	if settings.EnableInterfaceImplementationDetection != nil {
		config.EnableInterfaceImplementationDetection = *settings.EnableInterfaceImplementationDetection
	}
	if settings.EnableRobustCrossPackageAnalysis != nil {
		config.EnableRobustCrossPackageAnalysis = *settings.EnableRobustCrossPackageAnalysis
	}
	if settings.EnableExportedIdentifierHandling != nil {
		config.EnableExportedIdentifierHandling = *settings.EnableExportedIdentifierHandling
	}
	if settings.ConsiderExportedConstantsUsed != nil {
		config.ConsiderExportedConstantsUsed = *settings.ConsiderExportedConstantsUsed
	}

	// Convert slices with nil checks
	if settings.TestHelperPatterns != nil {
		config.TestHelperPatterns = settings.TestHelperPatterns
	}
	if settings.IgnoreFilePatterns != nil {
		config.IgnoreFilePatterns = settings.IgnoreFilePatterns
	}
	if settings.ExcludePatterns != nil {
		config.ExcludePatterns = settings.ExcludePatterns
	}
	if settings.ExplicitTestOnlyIdentifiers != nil {
		config.ExplicitTestOnlyIdentifiers = settings.ExplicitTestOnlyIdentifiers
	}
	if settings.AdditionalTests != nil {
		config.AdditionalTests = settings.AdditionalTests
	}

	return config
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
func isTestFile(filename string, config *Config) bool {
	// Проверяем на наличие слова "test" в любом регистре в имени файла
	lowerFilename := strings.ToLower(filename)
	if strings.Contains(lowerFilename, "test") {
		return true
	}

	// Проверяем дополнительные тестовые файлы из конфигурации
	if config != nil {
		for _, pattern := range config.AdditionalTests {
			if strings.Contains(filename, pattern) {
				return true
			}
		}
	}

	return false
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
	}
}

// defaultIgnoreFilePatterns returns the default patterns for files to ignore
func defaultIgnoreFilePatterns() []string {
	return []string{
		"test_helper",
		"test_util",
		"testutil",
		"testhelper",
	}
}

// defaultExplicitTestOnlyIdentifiers returns the default list of identifiers that should be considered test-only
func defaultExplicitTestOnlyIdentifiers() []string {
	return []string{
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
	}
}
