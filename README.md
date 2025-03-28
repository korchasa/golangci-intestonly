# Intestonly Linter

A Go static code analyzer that identifies declarations in non-test files that are only used in test files. This helps improve code quality by detecting code that can be safely removed from your production codebase.

## Overview

The Intestonly linter follows the principle that code exclusively used for testing represents unnecessary bloat in production. It analyzes your Go codebase to find declarations (functions, types, constants, variables) that are candidates for deletion because they're not used in actual production code.

```go
// Non-test file (main.go) with code only used in tests
func unusedFunction() string { // This is dead code for production
  return "test"
}

// In test file (main_test.go)
func TestSomething(t *testing.T) {
  result := unusedFunction() // Only usage is here
  assert.Equal(t, "test", result)
}
```

## Installation

```bash
go install github.com/korchasa/golangci-intestonly/cmd/go-intestonly@latest
```

## Usage

The Intestonly linter can be used in multiple ways depending on your workflow preferences:

### Direct Usage

Run the linter directly on your codebase:

```bash
# Analyze a specific package
go-intestonly ./pkg/...

# Analyze the current directory
go-intestonly .

# Show more details with context (3 lines)
go-intestonly -c=3 ./...

# Output in JSON format
go-intestonly -json ./...
```

### golangci-lint Integration

Intestonly is not yet included in the standard golangci-lint distribution. To integrate it, use the plugin approach:

1. Create a plugin file in your project:

```go
// tools/golangci/intestonly/plugin.go
package main

import (
	_ "github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly" // Import the linter
)
```

2. Configure `.golangci.yml` to load the plugin:

```yaml
linters-settings:
  custom:
    intestonly:
      path: tools/golangci/intestonly/plugin.so
      description: Checks for code that is only used in tests
      original-url: github.com/korchasa/golangci-intestonly

linters:
  enable:
    - intestonly
```

3. Build the plugin:

```bash
cd tools/golangci/intestonly
go build -buildmode=plugin -o plugin.so plugin.go
```

After setup, run as usual:

```bash
golangci-lint run
```

### CI/CD Pipeline Integration

Add to your GitHub Actions workflow:

```yaml
- name: Check for test-only code
  run: |
    go install github.com/korchasa/golangci-intestonly/cmd/go-intestonly@latest
    go-intestonly ./...
```

Or in your GitLab CI:

```yaml
lint:
  script:
    - go install github.com/korchasa/golangci-intestonly/cmd/go-intestonly@latest
    - go-intestonly ./...
```

## Configuration

### Configuring in .golangci.yml

Intestonly offers several configuration options to customize its behavior for your specific codebase:

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

    # List of explicit test-only identifiers that should always be reported
    explicit-test-only-identifiers:
      - testOnlyFunction
      - TestOnlyType
      - helperFunction

    # Whether to report explicit test cases regardless of usage
    report-explicit-test-cases: true

    # Debug mode
    debug: false
```

#### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `check-methods` | Check methods (functions with receivers) | `true` |
| `ignore-unexported` | Ignore unexported identifiers | `false` |
| `enable-content-based-detection` | Enable detection based on file contents | `true` |
| `exclude-test-helpers` | Exclude functions that look like test helpers | `true` |
| `test-helper-patterns` | Patterns for identifying test helpers | `[assert, mock, fake, stub, ...]` |
| `ignore-file-patterns` | Patterns for files to ignore | `[test_helper, test_util, ...]` |
| `exclude-patterns` | Identifiers to always exclude | `[]` |
| `explicit-test-only-identifiers` | Identifiers that should always be reported | `[testOnlyFunction, ...]` |
| `report-explicit-test-cases` | Report explicit test cases regardless of usage | `true` |
| `debug` | Enable debug output | `false` |

## Example Output

When the linter detects code only used in tests, it produces output like this:

```
/path/to/your/project/utils.go:15:1: function 'helperFunction' is only used in tests
/path/to/your/project/models.go:42:1: type 'TestModel' is only used in tests
/path/to/your/project/constants.go:8:5: const 'testOnlyConstant' is only used in tests
```

## Example Scenario

Consider this example where a utility function is only referenced in tests:

```go
// main.go
package main

func main() {
    // Production code
}

func formatData(data string) string {
    // This function is never used in production code
    return "[" + data + "]"
}

// main_test.go
package main

import "testing"

func TestFormatting(t *testing.T) {
    result := formatData("test")
    if result != "[test]" {
        t.Errorf("Expected [test], got %s", result)
    }
}
```

Running the linter will identify `formatData` as test-only code:

```bash
$ go-intestonly .
/path/to/your/project/main.go:8:1: function 'formatData' is only used in tests
```

## Implementation Details

### Core Algorithm

The analyzer uses a multi-pass strategy:
1. **Collect declarations** from non-test files (functions, types, constants, variables)
2. **Track identifier usage** in both test and non-test contexts
3. **Analyze cross-package references** to detect usage across package boundaries
4. **Perform content-based analysis** to detect implicit usage in strings or comments
5. **Report** identifiers that appear only in test usage contexts as candidates for removal

### Smart Detection

The analyzer includes special handling to:
- Detect test helper patterns by naming conventions (configurable)
- Skip test utility files entirely (configurable patterns)
- Handle method calls through selector expressions
- Process type usages and embedded types
- Analyze cross-package references
- Detect usage in string literals and comments

### Test Coverage

The analyzer has been tested against various scenarios:

| Category | Description | Handling |
|----------|-------------|----------|
| True Positives | Functions, types, and constants only used in tests | Detected ✓ |
| False Positives | Items used in both test and non-test files | Excluded ✓ |
| Edge Cases | Test helpers and utility functions | Intelligently filtered ✓ |
| False Negatives | Items used through reflection or type assertions | Detected ✓ |
| Nested Structures | Inner types, methods, embedded types | Properly analyzed ✓ |
| Interfaces | Interface types, methods, and type aliases | Correctly handled ✓ |
| Cross-Package | References across package boundaries | Tracked and analyzed ✓ |

## Project Structure

```
golangci-intestonly/
├── cmd/go-intestonly/    # CLI entry point
├── pkg/golinters/        # Core implementation
├── testdata/             # Comprehensive test cases
└── README.md             # Documentation
```

## Requirements

- Go 1.21+
- Uses golang.org/x/tools/go/analysis framework
- Integrates with golangci-lint

## Benefits

- **Code Cleanliness**: Remove dead code that's only used in tests
- **Reduced Maintenance**: Fewer lines of code means less to maintain
- **Improved Build Times**: Remove unnecessary code from compilation
- **Better Documentation**: Clarify what code is actually used in production
- **Customizable**: Adapt to your codebase's specific needs via configuration

## License

[MIT License](LICENSE)
