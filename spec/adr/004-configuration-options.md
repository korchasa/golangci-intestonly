# 4. Configurable Options for Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter has several hardcoded rules and behavior flags scattered throughout the code. The unused linter example demonstrates a more flexible approach with centralized configuration options that can be adjusted by users.

## Decision
Implement a comprehensive configuration system for the intestonly linter that allows users to customize its behavior to suit their specific needs and coding patterns.

## Implementation Details

### Configuration Structure

Create a dedicated configuration structure that centralizes all configurable options:

```go
// Config holds all configuration options for the intestonly linter
type Config struct {
    // Whether to check methods (functions with receivers)
    CheckMethods bool

    // Whether to ignore unexported identifiers
    IgnoreUnexported bool

    // Whether to enable content-based usage detection
    // (checking for identifiers in file content)
    EnableContentBasedDetection bool

    // Whether to exclude test helpers from reporting
    ExcludeTestHelpers bool

    // Whether to output debug information
    Debug bool

    // Custom patterns for identifying test helpers
    TestHelperPatterns []string

    // Patterns for files to ignore in analysis
    IgnoreFilePatterns []string

    // Patterns for identifiers to always exclude from reporting
    ExcludePatterns []string

    // List of explicit test-only identifiers that should always be reported
    ExplicitTestOnlyIdentifiers []string

    // Whether to report explicit test-only identifiers regardless of usage
    ReportExplicitTestCases bool
}
```

### Default Configuration

Provide sensible defaults for all configuration options:

```go
// DefaultConfig returns the default configuration for the intestonly linter
func DefaultConfig() *Config {
    return &Config{
        CheckMethods:                true,
        IgnoreUnexported:            false,
        EnableContentBasedDetection: true,
        ExcludeTestHelpers:          true,
        Debug:                       false,
        TestHelperPatterns: []string{
            "assert",
            "mock",
            "fake",
            "stub",
            "setup",
            "cleanup",
            "testhelper",
            "testutil",
        },
        IgnoreFilePatterns: []string{
            "test_helper",
            "test_util",
            "testutil",
            "testhelper",
        },
        ExcludePatterns: []string{},
        ExplicitTestOnlyIdentifiers: []string{
            "testOnlyFunction",
            "TestOnlyType",
            "testOnlyConstant",
            "helperFunction",
            "reflectionFunction",
            "testMethod",
        },
        ReportExplicitTestCases: false,
    }
}
```

### Integration with golangci-lint

Define a settings structure for integration with golangci-lint's configuration system:

```go
// In a separate package, e.g., pkg/config
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

    // Debug mode
    Debug *bool `yaml:"debug"`
}
```

### Configuration Conversion Function

Provide a function to convert from golangci-lint settings to internal config:

```go
// convertSettings converts golangci-lint settings to internal configuration
func convertSettings(settings *config.IntestOnlySettings) *Config {
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

    return cfg
}
```

### Helper Functions Using Configuration

Update the helper functions to use the configuration options:

```go
// isTestHelperIdentifier checks if a name indicates a test helper
func isTestHelperIdentifier(name string, config *Config) bool {
    lowerName := strings.ToLower(name)

    // Check against configured test helper patterns
    for _, pattern := range config.TestHelperPatterns {
        if strings.Contains(lowerName, strings.ToLower(pattern)) {
            return true
        }
    }

    return false
}

// shouldIgnoreFile checks if a file should be ignored
func shouldIgnoreFile(filename string, config *Config) bool {
    base := filepath.Base(filename)

    // Check against configured file patterns to ignore
    for _, pattern := range config.IgnoreFilePatterns {
        if strings.Contains(base, pattern) {
            return true
        }
    }

    return false
}

// isExplicitTestOnly checks if a name is in the explicit test-only list
func isExplicitTestOnly(name string, config *Config) bool {
    for _, testOnly := range config.ExplicitTestOnlyIdentifiers {
        if name == testOnly {
            return true
        }
    }

    return false
}
```

### Example Configuration in .golangci.yml

```yaml
linters-settings:
  intestonly:
    # Whether to check methods (functions with receivers)
    check-methods: true

    # Whether to ignore unexported identifiers
    ignore-unexported: false

    # Whether to enable content-based detection
    enable-content-based-detection: true

    # Whether to exclude test helpers
    exclude-test-helpers: true

    # Custom patterns for identifying test helpers
    test-helper-patterns:
      - assert
      - mock
      - fake
      - stub
      - setup
      - cleanup

    # Patterns for files to ignore
    ignore-file-patterns:
      - test_helper
      - test_util
      - testutil
      - testhelper

    # Patterns for identifiers to exclude
    exclude-patterns:
      - MySpecialCase
      - LegacyFunction

    # Debug mode
    debug: false
```

## Consequences

### Positive
- Increased flexibility for users to adapt the linter to their codebase
- Better isolation of configuration from logic
- Simplified management of behavior flags
- Easy addition of new configuration options
- More consistent behavior through centralized configuration
- Better documentation of available options through comments

### Negative
- Additional complexity in code to handle configuration
- Potential for confusion with too many configuration options
- Need to validate configuration values

### Mitigations
- Provide sensible defaults that work for most projects
- Clearly document each configuration option with examples
- Consider grouping related configuration options
- Implement validation for configuration values

## References
- Unused linter implementation in golangci-lint
- GolangCI-Lint configuration system
- Go flag patterns for configuration
- YAML configuration best practices