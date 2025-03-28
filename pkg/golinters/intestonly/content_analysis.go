package intestonly

import (
	"fmt"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzeContentBasedUsages performs additional file content analysis
func analyzeContentBasedUsages(pass *analysis.Pass, result *AnalysisResult, config *Config) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()

		// Skip test files
		if isTestFile(fileName, config) || shouldIgnoreFile(fileName, config) {
			continue
		}

		// Analyze file content for potential usages
		analyzeFileContent(fileName, result, config)
	}
}

// analyzeFileContent scans a file's content for potential declaration usages
func analyzeFileContent(fileName string, result *AnalysisResult, config *Config) {
	// Read file content
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		if config.Debug {
			fmt.Printf("Error reading file %s: %v\n", fileName, err)
		}
		return
	}

	contentStr := string(fileContent)

	// Check for declaration names in the content
	for name := range result.Declarations {
		if strings.Contains(contentStr, name) {
			// Mark as potentially used in non-test code
			if len(result.NonTestUsages[name]) == 0 {
				result.NonTestUsages[name] = append(result.NonTestUsages[name], token.NoPos)
			}
		}
	}
}
