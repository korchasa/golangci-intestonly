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

### Pre-commit Hook

Create a pre-commit hook to check before committing:

```bash
#!/bin/sh
go-intestonly ./...
```

### Using with Go Commands

Run as part of your test workflow:

```bash
go test ./... && go-intestonly ./...
```

## Practical Examples

### Example Output

When the linter detects code only used in tests, it produces output like this:

```
/path/to/your/project/utils.go:15:1: function 'helperFunction' is only used in tests
/path/to/your/project/models.go:42:1: type 'TestModel' is only used in tests
/path/to/your/project/constants.go:8:5: const 'testOnlyConstant' is only used in tests
```

### Real-world Scenario

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

The analyzer uses a two-pass strategy:
1. **Collect declarations** from non-test files (functions, types, constants, variables)
2. **Track identifier usage** in both test and non-test contexts
3. **Report** identifiers that appear only in test usage contexts as candidates for removal

### Smart Detection

The analyzer includes special handling to:
- Detect test helper patterns by naming conventions
- Skip test utility files entirely
- Handle method calls through selector expressions
- Process type usages and embedded types

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

## License

[MIT License](LICENSE)
