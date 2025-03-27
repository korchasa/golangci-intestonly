# 6. Improved Identifier Usage Detection in Intestonly Linter

## Status
Proposed

## Date
2024-03-27

## Context
The current implementation of the intestonly linter has several critical flaws in how it detects identifier usage:

1. **Primitive Usage Detection via String Matching**: The linter uses a basic `strings.Contains()` approach to detect potential usage in files. This inevitably leads to false negatives (missed actual usages) and false positives (reported non-usages) because:
   - Identifiers may appear in strings or comments
   - Identifiers may be part of larger identifiers (e.g., `testFunc` inside `runTestFunc`)
   - This approach doesn't respect scope or actual references

2. **Limited AST Analysis**: The current AST analysis does not fully track all usage patterns, including:
   - Variable shadowing
   - Type embedding
   - Indirect references through interfaces
   - Reflection-based usage

3. **Inability to Detect Implicit Usage**: Many Go patterns involve implicit usage that's not directly visible in the AST:
   - Init functions used for side effects
   - Registry/plugin patterns where functions register themselves
   - Runtime type identification (RTTI) use cases

These limitations result in both false positives (incorrectly flagged test-only code) and false negatives (missed actual test-only code), reducing the usefulness and trustworthiness of the linter.

## Decision
Implement a comprehensive and robust identifier usage detection system that accurately tracks all usage patterns in both test and non-test code.

## Implementation Details

### AST-Based Usage Analysis

1. Replace string-based detection with proper AST traversal and type checking:

```go
// analyzeIdentifierUsage performs a comprehensive analysis of identifier usage
func analyzeIdentifierUsage(pass *analysis.Pass, result *AnalysisResult, config *Config) {
    // Process each file
    for _, file := range pass.Files {
        fileName := pass.Fset.File(file.Pos()).Name()
        isTest := isTestFile(fileName)

        // Create a set of all identifiers in this file's scope to detect shadowing
        fileScope := make(map[string]bool)

        // First pass: collect all local declarations to handle shadowing
        ast.Inspect(file, func(n ast.Node) bool {
            switch node := n.(type) {
            case *ast.FuncDecl:
                if node.Name != nil {
                    fileScope[node.Name.Name] = true
                }
            case *ast.GenDecl:
                for _, spec := range node.Specs {
                    if valueSpec, ok := spec.(*ast.ValueSpec); ok {
                        for _, name := range valueSpec.Names {
                            fileScope[name.Name] = true
                        }
                    } else if typeSpec, ok := spec.(*ast.TypeSpec); ok {
                        if typeSpec.Name != nil {
                            fileScope[typeSpec.Name.Name] = true
                        }
                    }
                }
            case *ast.AssignStmt:
                if node.Tok == token.DEFINE {
                    for _, lhs := range node.Lhs {
                        if ident, ok := lhs.(*ast.Ident); ok {
                            fileScope[ident.Name] = true
                        }
                    }
                }
            }
            return true
        })

        // Second pass: analyze actual usages
        ast.Inspect(file, func(n ast.Node) bool {
            switch node := n.(type) {
            case *ast.Ident:
                // Skip declarations and local shadowed identifiers
                if fileScope[node.Name] {
                    return true
                }

                // Look up the object this identifier refers to
                obj := pass.TypesInfo.Uses[node]
                if obj == nil {
                    return true
                }

                // Check if this is a reference to a declaration we're tracking
                declInfo, exists := result.Declarations[obj.Name()]
                if !exists {
                    return true
                }

                // Verify this is the same object as our declaration (not shadowed)
                if obj.Pos() != declInfo.Pos {
                    return true
                }

                // Record usage based on whether this is a test file
                if isTest {
                    result.AddTestUsage(obj.Name())
                } else {
                    result.AddNonTestUsage(obj.Name())
                }

            case *ast.SelectorExpr:
                // Handle qualified references (pkg.Func or x.Method)
                if x, ok := node.X.(*ast.Ident); ok {
                    sel := node.Sel

                    // Check if this is a package-qualified reference
                    pkgObj := pass.TypesInfo.Uses[x]
                    if pkgObj != nil && pkgObj.Pkg() != nil {
                        // Package-qualified reference, check against imports
                        importPath := pkgObj.Pkg().Path()
                        fullName := importPath + "." + sel.Name

                        // Check if this matches one of our tracked declarations
                        for declName, info := range result.Declarations {
                            if info.ImportRef == fullName {
                                if isTest {
                                    result.AddTestUsage(declName)
                                } else {
                                    result.AddNonTestUsage(declName)
                                }
                            }
                        }
                    } else {
                        // Check if this is a method call on a known type
                        xType := pass.TypesInfo.Types[node.X].Type
                        if xType != nil {
                            // Get the method being called
                            method, exists := getMethodFromType(xType, sel.Name, pass)
                            if exists {
                                // Check if this method is one we're tracking
                                for declName, info := range result.Declarations {
                                    if info.Name == method.Name() && info.Pos == method.Pos() {
                                        if isTest {
                                            result.AddTestUsage(declName)
                                        } else {
                                            result.AddNonTestUsage(declName)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            return true
        })
    }
}

// getMethodFromType gets a method from a type
func getMethodFromType(t types.Type, methodName string, pass *analysis.Pass) (*types.Func, bool) {
    // Get the method set for this type
    methodSet := types.NewMethodSet(t)

    // Look for the method by name
    for i := 0; i < methodSet.Len(); i++ {
        m := methodSet.At(i)
        if m.Obj().Name() == methodName {
            if fn, ok := m.Obj().(*types.Func); ok {
                return fn, true
            }
        }
    }

    return nil, false
}
```

### Type Embedding and Interface Analysis

1. Add special handling for embedded fields and interface implementation:

```go
// analyzeTypeEmbedding tracks usage through type embedding
func analyzeTypeEmbedding(pass *analysis.Pass, result *AnalysisResult) {
    for _, file := range pass.Files {
        ast.Inspect(file, func(n ast.Node) bool {
            // Look for type definitions with embedded fields
            typeSpec, ok := n.(*ast.TypeSpec)
            if !ok || typeSpec.Name == nil {
                return true
            }

            structType, ok := typeSpec.Type.(*ast.StructType)
            if !ok || structType.Fields == nil {
                return true
            }

            // Find embedded fields
            for _, field := range structType.Fields.List {
                if len(field.Names) == 0 { // Embedded field
                    // Get the type of the embedded field
                    typeInfo, ok := pass.TypesInfo.Types[field.Type]
                    if !ok {
                        continue
                    }

                    // Get methods of the embedded type
                    methodSet := types.NewMethodSet(typeInfo.Type)
                    for i := 0; i < methodSet.Len(); i++ {
                        method := methodSet.At(i).Obj()
                        methodName := method.Name()

                        // Check if this method is in our declarations
                        for declName, info := range result.Declarations {
                            if info.Name == methodName {
                                // The containing type is using this method through embedding
                                isTestFile := isTestFile(pass.Fset.File(typeSpec.Pos()).Name())
                                if isTestFile {
                                    result.AddTestUsage(declName)
                                } else {
                                    result.AddNonTestUsage(declName)
                                }
                            }
                        }
                    }
                }
            }

            return true
        })
    }
}
```

### Reflection and Implicit Usage Detection

1. Add patterns for detecting reflection-based usage:

```go
// analyzeReflectionUsage detects usage through reflection
func analyzeReflectionUsage(pass *analysis.Pass, result *AnalysisResult, config *Config) {
    for _, file := range pass.Files {
        fileName := pass.Fset.File(file.Pos()).Name()
        isTest := isTestFile(fileName)

        ast.Inspect(file, func(n ast.Node) bool {
            // Look for reflection-based usage patterns
            // Such as reflect.TypeOf(), reflect.ValueOf() followed by .Method() or .FieldByName()

            call, ok := n.(*ast.CallExpr)
            if !ok {
                return true
            }

            // Check for reflect.TypeOf() or reflect.ValueOf()
            if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                if x, ok := sel.X.(*ast.Ident); ok && x.Name == "reflect" {
                    if sel.Sel.Name == "TypeOf" || sel.Sel.Name == "ValueOf" {
                        // Look for string literals in method or field access
                        // that match our declarations
                        for _, arg := range call.Args {
                            // If the argument is a reference to a type or variable we're tracking,
                            // mark it as used
                            if ident, ok := arg.(*ast.Ident); ok {
                                for declName := range result.Declarations {
                                    if declName == ident.Name {
                                        if isTest {
                                            result.AddTestUsage(declName)
                                        } else {
                                            result.AddNonTestUsage(declName)
                                        }
                                    }
                                }
                            }
                        }

                        // Also check for method calls on the result
                        // Look at parent expressions for e.g., reflect.TypeOf(x).Method(0).Name
                        parent := getParentNode(file, call)
                        if parent != nil {
                            if parentSel, ok := parent.(*ast.SelectorExpr); ok {
                                if parentSel.Sel.Name == "Method" || parentSel.Sel.Name == "FieldByName" {
                                    // This is a reflection-based method or field access
                                    // Mark all methods as potentially used
                                    for declName, info := range result.Declarations {
                                        if info.IsMethod && config.ConsiderReflectionRisky {
                                            // Mark all methods as potentially used under reflection
                                            if isTest {
                                                result.AddTestUsage(declName)
                                            } else {
                                                result.AddNonTestUsage(declName)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            return true
        })
    }
}

// getParentNode finds the parent expression containing the given node
func getParentNode(file *ast.File, target ast.Node) ast.Node {
    var parent ast.Node

    ast.Inspect(file, func(n ast.Node) bool {
        if n == target {
            return false
        }

        switch expr := n.(type) {
        case *ast.SelectorExpr:
            if expr.X == target {
                parent = expr
                return false
            }
        case *ast.CallExpr:
            if expr.Fun == target {
                parent = expr
                return false
            }
        }

        return true
    })

    return parent
}
```

### Registry and Plugin Pattern Detection

1. Detect common registry and plugin patterns:

```go
// analyzeRegistryPatterns detects usage in registry and plugin patterns
func analyzeRegistryPatterns(pass *analysis.Pass, result *AnalysisResult, config *Config) {
    for _, file := range pass.Files {
        fileName := pass.Fset.File(file.Pos()).Name()
        isTest := isTestFile(fileName)

        ast.Inspect(file, func(n ast.Node) bool {
            // Look for common registration patterns
            // 1. init() functions with side effects
            if funcDecl, ok := n.(*ast.FuncDecl); ok {
                if funcDecl.Name != nil && funcDecl.Name.Name == "init" && funcDecl.Recv == nil {
                    // Assume all declarations used in init() are used by the package
                    ast.Inspect(funcDecl, func(m ast.Node) bool {
                        if ident, ok := m.(*ast.Ident); ok {
                            for declName := range result.Declarations {
                                if declName == ident.Name {
                                    if isTest {
                                        result.AddTestUsage(declName)
                                    } else {
                                        result.AddNonTestUsage(declName)
                                    }
                                }
                            }
                        }
                        return true
                    })
                }
            }

            // 2. Map assignments that may be registries
            if assign, ok := n.(*ast.AssignStmt); ok {
                for _, rhs := range assign.Rhs {
                    ast.Inspect(rhs, func(m ast.Node) bool {
                        if ident, ok := m.(*ast.Ident); ok {
                            for declName := range result.Declarations {
                                if declName == ident.Name {
                                    // Check if this is a map assignment
                                    for _, lhs := range assign.Lhs {
                                        if idxExpr, ok := lhs.(*ast.IndexExpr); ok {
                                            // This looks like a map registration pattern
                                            if isTest {
                                                result.AddTestUsage(declName)
                                            } else {
                                                result.AddNonTestUsage(declName)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        return true
                    })
                }
            }

            return true
        })
    }
}
```

## Consequences

### Positive
- Dramatically improved accuracy in identifier usage detection
- Greatly reduced false positives and false negatives
- Better handling of complex Go idioms and patterns
- More accurate identification of truly test-only identifiers
- Increased user trust in the linter's results

### Negative
- Increased complexity in the analysis logic
- Potential performance impact from deeper AST traversal
- More memory usage for comprehensive type information
- Difficulty in detecting some extremely dynamic patterns (e.g., runtime codegen)

### Mitigations
- Implement efficient data structures and algorithms
- Add configuration to selectively enable/disable certain analysis passes
- Add safeguards to prevent excessive resource consumption
- Provide documentation about known edge cases and limitations
- Implement unit tests to validate detection across various patterns

## References
- Go AST Package: https://golang.org/pkg/go/ast/
- Go Types Package: https://golang.org/pkg/go/types/
- Reflection in Go: https://blog.golang.org/laws-of-reflection
- Go Static Analysis Tools: https://staticcheck.io/
- Common Go Patterns: https://github.com/golang/go/wiki/CodeReviewComments