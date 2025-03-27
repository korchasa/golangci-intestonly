# Intestonly Linter

A Go static code analyzer that identifies declarations in non-test files that are only used in test files. This helps improve code quality by detecting code that can be safely removed.

## Overview

The Intestonly linter follows the principle that code exclusively used for testing represents unnecessary bloat in production code. It analyzes your Go codebase to find declarations (functions, types, constants, variables) that are candidates for deletion because they're not used in production.

```go
// Non-test file with code only used in tests - CAN BE REMOVED!
func unusedFunction() string { // This is dead code for production
  return "test"
}

// In test file
func TestSomething(t *testing.T) {
  result := unusedFunction() // Only usage is here
  assert.Equal(t, "test", result)
}
```

## Installation

```bash
go install github.com/korchasa/golangci-intestonly/cmd/go-intestonly@latest
```

## golangci-lint Integration

Add to your `.golangci.yml`:

```yaml
linters:
  enable:
    - intestonly
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
