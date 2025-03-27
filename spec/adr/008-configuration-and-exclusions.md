# 8. Comprehensive Configuration System for Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter has significant limitations in configurability:

1. **Hardcoded Exclusion Rules**: The functions `shouldIgnoreFile`, `isTestHelperIdentifier`, and `shouldExcludeFromReport` contain hardcoded patterns and rules that cannot be customized by users.

2. **Lack of User Configuration**: There is no mechanism for users to customize the linter's behavior to match their project's specific conventions or needs.

3. **Inconsistent Treatment of Edge Cases**: Without configurable rules, the linter treats all projects the same, regardless of their unique naming conventions or architecture.

4. **No Documentation of Available Options**: Users have no way to understand what options might be available or how to configure the linter.

These limitations reduce the usefulness of the linter across different projects and force users to modify their code to match the linter's expectations rather than adapting the linter to their codebase.

## Decision
Implement a comprehensive configuration system for the intestonly linter that allows users to customize its behavior through a configuration file, command-line flags, or in-code directives.

## Implementation Details

### Configuration Structure

1. Define a comprehensive configuration structure:

```go
// Config holds the configuration for the intestonly linter
type Config struct {
    // Whether to enable debug output
    Debug bool `yaml:"debug" json:"debug"`

    // Whether to check methods (functions with receivers)
    CheckMethods bool `yaml:"check-methods" json:"check_methods"`

    // Whether to ignore unexported identifiers
    IgnoreUnexported bool `yaml:"ignore-unexported" json:"ignore_unexported"`

    // Whether to respect exported status (skip reporting exported identifiers)
    ConsiderExportedStatus bool `yaml:"consider-exported-status" json:"consider_exported_status"`

    // Whether to enable the content-based usage detection
    EnableContentBasedDetection bool `yaml:"enable-content-based-detection" json:"enable_content_based_detection"`

    // Whether to exclude test helpers from reporting
    ExcludeTestHelpers bool `yaml:"exclude-test-helpers" json:"exclude_test_helpers"`

    // Consider reflection as potentially risky (mark reflected methods as used)
    ConsiderReflectionRisky bool `yaml:"consider-reflection-risky" json:"consider_reflection_risky"`

    // Enable incremental analysis to improve performance on repeated runs
    EnableIncrementalAnalysis bool `yaml:"enable-incremental-analysis" json:"enable_incremental_analysis"`

    // Maximum number of worker goroutines for parallel processing
    MaxWorkers int `yaml:"max-workers" json:"max_workers"`

    // Patterns for identifying test helper identifiers
    TestHelperPatterns []string `yaml:"test-helper-patterns" json:"test_helper_patterns"`

    // Patterns for files to ignore (glob patterns)
    IgnoreFilePatterns []string `yaml:"ignore-file-patterns" json:"ignore_file_patterns"`

    // Patterns for identifiers to exclude from reporting
    ExcludePatterns []string `yaml:"exclude-patterns" json:"exclude_patterns"`

    // List of explicit test-only identifiers that should always be reported
    ExplicitTestOnlyIdentifiers []string `yaml:"explicit-test-only-identifiers" json:"explicit_test_only_identifiers"`

    // Directories to scan (relative to project root)
    IncludeDirs []string `yaml:"include-dirs" json:"include_dirs"`

    // Directories to exclude from scanning
    ExcludeDirs []string `yaml:"exclude-dirs" json:"exclude_dirs"`

    // Whether to include vendor directory in analysis
    IncludeVendor bool `yaml:"include-vendor" json:"include_vendor"`
}
```

### Default Configuration

1. Provide sensible defaults:

```go
// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
    return &Config{
        Debug:                       false,
        CheckMethods:                true,
        IgnoreUnexported:            false,
        ConsiderExportedStatus:      true,
        EnableContentBasedDetection: true,
        ExcludeTestHelpers:          true,
        ConsiderReflectionRisky:     true,
        EnableIncrementalAnalysis:   true,
        MaxWorkers:                  runtime.NumCPU(),
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
        ExcludePatterns:             []string{},
        ExplicitTestOnlyIdentifiers: []string{},
        IncludeDirs:                 []string{"."},
        ExcludeDirs:                 []string{"vendor"},
        IncludeVendor:               false,
    }
}
```

### Configuration Loading

1. Implement configuration loading from multiple sources:

```go
// ConfigLoader loads configuration from various sources
type ConfigLoader struct {
    // Paths to look for configuration files
    ConfigPaths []string

    // Command-line flags that override config file settings
    FlagOverrides map[string]string

    // Environment variables that override config file settings
    EnvOverrides map[string]string
}

// NewConfigLoader creates a new config loader with default paths
func NewConfigLoader() *ConfigLoader {
    return &ConfigLoader{
        ConfigPaths: []string{
            ".intestonly.yml",
            ".intestonly.yaml",
            ".golangci.yml",
            ".golangci.yaml",
        },
        FlagOverrides: make(map[string]string),
        EnvOverrides:  make(map[string]string),
    }
}

// LoadConfig loads configuration from all available sources
func (l *ConfigLoader) LoadConfig() (*Config, error) {
    // Start with default configuration
    config := DefaultConfig()

    // Try loading from configuration files
    for _, path := range l.ConfigPaths {
        if err := l.loadFromFile(path, config); err == nil {
            // Successfully loaded from file
            break
        }
    }

    // Apply environment variable overrides
    l.applyEnvOverrides(config)

    // Apply command-line flag overrides
    l.applyFlagOverrides(config)

    return config, nil
}

// Implementation of file loading, env var processing, etc.
// ...
```

### Using the Configuration

1. Update helper functions to use the configuration:

```go
// shouldIgnoreFile checks if a file should be ignored based on configuration
func shouldIgnoreFile(filename string, config *Config) bool {
    base := filepath.Base(filename)

    // Check against configured patterns
    for _, pattern := range config.IgnoreFilePatterns {
        matched, err := filepath.Match(pattern, base)
        if err == nil && matched {
            return true
        }
        if strings.Contains(base, pattern) {
            return true
        }
    }

    // Check if this is in an excluded directory
    for _, dir := range config.ExcludeDirs {
        if strings.HasPrefix(filename, dir) {
            return true
        }
    }

    // Check if this is in vendor and vendor is excluded
    if !config.IncludeVendor && strings.Contains(filename, "/vendor/") {
        return true
    }

    return false
}

// isTestHelperIdentifier checks if an identifier should be considered a test helper
func isTestHelperIdentifier(name string, config *Config) bool {
    if !config.ExcludeTestHelpers {
        return false
    }

    lowerName := strings.ToLower(name)

    // Check against configured patterns
    for _, pattern := range config.TestHelperPatterns {
        if strings.HasPrefix(lowerName, strings.ToLower(pattern)) {
            return true
        }
    }

    return false
}

// shouldExcludeFromReport checks if an identifier should be excluded from the report
func shouldExcludeFromReport(name string, info *DeclInfo, config *Config) bool {
    // Check if this is unexported and we're ignoring unexported
    if config.IgnoreUnexported && !ast.IsExported(name) {
        return true
    }

    // Check if this is a method and we're not checking methods
    if !config.CheckMethods && info.IsMethod {
        return true
    }

    // Check if this is a test helper
    if isTestHelperIdentifier(name, config) {
        return true
    }

    // Check against explicit exclusion patterns
    for _, pattern := range config.ExcludePatterns {
        matched, err := filepath.Match(pattern, name)
        if err == nil && matched {
            return true
        }
    }

    // Consider exported status if configured
    if config.ConsiderExportedStatus && ast.IsExported(name) {
        // Additional logic for exported identifiers could go here
    }

    return false
}
```

### In-Code Directives

1. Implement support for in-code directives to control the linter:

```go
// directiveMarker is the comment prefix that indicates a linter directive
const directiveMarker = "intestonly:"

// Directive represents a linter directive found in the code
type Directive struct {
    Command string   // The directive command
    Args    []string // Arguments to the directive
    Pos     token.Pos // Position in the source
}

// parseDirectives extracts linter directives from comments
func parseDirectives(file *ast.File, fset *token.FileSet) []Directive {
    var directives []Directive

    for _, commentGroup := range file.Comments {
        for _, comment := range commentGroup.List {
            text := comment.Text

            // Check if this is a linter directive
            if !strings.Contains(text, directiveMarker) {
                continue
            }

            // Extract the directive part
            parts := strings.SplitN(text, directiveMarker, 2)
            if len(parts) != 2 {
                continue
            }

            // Parse the directive
            directive := strings.TrimSpace(parts[1])
            cmdAndArgs := strings.Fields(directive)
            if len(cmdAndArgs) == 0 {
                continue
            }

            directives = append(directives, Directive{
                Command: cmdAndArgs[0],
                Args:    cmdAndArgs[1:],
                Pos:     comment.Pos(),
            })
        }
    }

    return directives
}

// processFileDirectives handles directives in a file
func processFileDirectives(file *ast.File, fset *token.FileSet, config *Config) {
    directives := parseDirectives(file, fset)

    for _, directive := range directives {
        switch directive.Command {
        case "ignore": // Ignore specific identifiers
            if len(directive.Args) > 0 {
                for _, arg := range directive.Args {
                    config.ExcludePatterns = append(config.ExcludePatterns, arg)
                }
            }

        case "ignore-file": // Ignore the entire file
            fileName := fset.File(file.Pos()).Name()
            config.IgnoreFilePatterns = append(config.IgnoreFilePatterns, filepath.Base(fileName))

        // Other directive types...
        }
    }
}
```

### Documentation Generation

1. Add a command to generate documentation for the configuration options:

```go
// generateConfigDocumentation creates documentation for the configuration options
func generateConfigDocumentation() string {
    var doc strings.Builder

    doc.WriteString("# Intestonly Linter Configuration\n\n")
    doc.WriteString("This document describes the configuration options for the intestonly linter.\n\n")

    doc.WriteString("## Configuration File\n\n")
    doc.WriteString("The linter looks for configuration in the following files (in order of precedence):\n\n")
    doc.WriteString("- `.intestonly.yml`\n")
    doc.WriteString("- `.intestonly.yaml`\n")
    doc.WriteString("- `.golangci.yml` (under the `linters-settings.intestonly` section)\n")
    doc.WriteString("- `.golangci.yaml` (under the `linters-settings.intestonly` section)\n\n")

    doc.WriteString("## Available Options\n\n")

    // List all options with descriptions
    doc.WriteString("| Option | Type | Default | Description |\n")
    doc.WriteString("|--------|------|---------|-------------|\n")
    doc.WriteString("| `debug` | boolean | `false` | Enable debug output |\n")
    doc.WriteString("| `check-methods` | boolean | `true` | Whether to check methods (functions with receivers) |\n")
    // ... add all other options

    doc.WriteString("\n## Example Configuration\n\n")
    doc.WriteString("```yaml\n")
    doc.WriteString("# .intestonly.yml\n")
    doc.WriteString("debug: false\n")
    doc.WriteString("check-methods: true\n")
    doc.WriteString("ignore-unexported: false\n")
    // ... example for all options
    doc.WriteString("```\n\n")

    doc.WriteString("## In-Code Directives\n\n")
    doc.WriteString("You can use the following directives in your code comments:\n\n")
    doc.WriteString("- `intestonly:ignore identifier1 identifier2 ...` - Ignore specific identifiers\n")
    doc.WriteString("- `intestonly:ignore-file` - Ignore the entire file\n")
    // ... add all directive types

    return doc.String()
}
```

### Configuration in golangci-lint

1. Add configuration integration with golangci-lint:

```go
// In golangci-lint integration package

// IntestOnlySettings represents the configuration for the intestonly linter
// in golangci-lint's configuration format
type IntestOnlySettings struct {
    Debug                       *bool    `yaml:"debug"`
    CheckMethods                *bool    `yaml:"check-methods"`
    IgnoreUnexported            *bool    `yaml:"ignore-unexported"`
    ConsiderExportedStatus      *bool    `yaml:"consider-exported-status"`
    EnableContentBasedDetection *bool    `yaml:"enable-content-based-detection"`
    ExcludeTestHelpers          *bool    `yaml:"exclude-test-helpers"`
    ConsiderReflectionRisky     *bool    `yaml:"consider-reflection-risky"`
    EnableIncrementalAnalysis   *bool    `yaml:"enable-incremental-analysis"`
    MaxWorkers                  *int     `yaml:"max-workers"`
    TestHelperPatterns          []string `yaml:"test-helper-patterns"`
    IgnoreFilePatterns          []string `yaml:"ignore-file-patterns"`
    ExcludePatterns             []string `yaml:"exclude-patterns"`
    ExplicitTestOnlyIdentifiers []string `yaml:"explicit-test-only-identifiers"`
    IncludeDirs                 []string `yaml:"include-dirs"`
    ExcludeDirs                 []string `yaml:"exclude-dirs"`
    IncludeVendor               *bool    `yaml:"include-vendor"`
}

// ToInternalConfig converts golangci-lint settings to the internal config format
func (s *IntestOnlySettings) ToInternalConfig() *intestonly.Config {
    config := intestonly.DefaultConfig()

    // Apply settings from golangci-lint config
    if s.Debug != nil {
        config.Debug = *s.Debug
    }

    if s.CheckMethods != nil {
        config.CheckMethods = *s.CheckMethods
    }

    // ... apply other settings

    return config
}
```

## Consequences

### Positive
- Users can adapt the linter to their project's specific needs
- Greater flexibility in handling edge cases and special patterns
- Improved user experience through documented configuration
- Better integration with golangci-lint and other tools
- Reduced false positives through customizable exclusion rules
- Support for larger teams with different code organization patterns

### Negative
- Increased implementation complexity
- More testing required to ensure all configuration options work correctly
- Potential for users to misconfigure and reduce the effectiveness of the linter
- Need to maintain backward compatibility for configuration in future versions

### Mitigations
- Comprehensive documentation for all configuration options
- Validation of configuration to prevent common mistakes
- Clear error messages for configuration problems
- Example configurations for common use cases
- Default values that work well for most projects

## References
- GolangCI-Lint configuration system: https://golangci-lint.run/usage/configuration/
- Go flag patterns: https://golang.org/pkg/flag/
- YAML configuration best practices: https://yaml.org/spec/1.2/spec.html
- Compiler directive patterns in Go: https://golang.org/cmd/compile/