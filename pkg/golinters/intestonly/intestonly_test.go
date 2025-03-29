package intestonly_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly"
	"golang.org/x/tools/go/analysis/analysistest"
)

// Set up test data before tests
func setupTestdata() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(wd))), "testdata")
}

// getTestDirectories returns all test directories in testdata/src
// that contain Go files and are suitable for individual test runs
func getTestDirectories() ([]string, error) {
	testdata := setupTestdata()
	srcDir := filepath.Join(testdata, "src")

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		panic(err)
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(srcDir, entry.Name())
		hasGoFiles := false

		files, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %v", dirPath, err)
		}

		for _, file := range files {
			if !file.IsDir() {
				if filepath.Ext(file.Name()) == ".go" {
					hasGoFiles = true
				}
			}
		}

		// Directory must have Go files
		if hasGoFiles {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// TestAllCases runs all test cases by automatically iterating through all test directories
func TestAllCases(t *testing.T) {
	testdata := setupTestdata()

	// Get all test directories
	testDirs, err := getTestDirectories()
	if err != nil {
		t.Fatalf("failed to get test directories: %v", err)
	}

	for _, dir := range testDirs {
		t.Run(dir, func(t *testing.T) {
			// Reset global result before each test to ensure a fresh state
			intestonly.ResetGlobalResult()
			analysistest.Run(t, testdata, intestonly.Analyzer, dir)
		})
	}
}

// TestConfiguration tests that configuration works correctly
func TestConfiguration(t *testing.T) {
	// Test default configuration
	cfg := intestonly.DefaultConfig()

	// Check default values for user-configurable options
	if cfg.Debug {
		t.Error("Default config should have Debug=false")
	}
	if len(cfg.OverrideIsCodeFiles) == 0 {
		t.Error("Default config should have non-empty OverrideIsCodeFiles")
	}
	if len(cfg.OverrideIsTestFiles) != 0 {
		t.Error("Default config should have empty OverrideIsTestFiles")
	}

	// Check default values for hardcoded options
	if !cfg.CheckMethods {
		t.Error("Default config should have CheckMethods=true")
	}
	if !cfg.ExcludeTestHelpers {
		t.Error("Default config should have ExcludeTestHelpers=true")
	}
	if !cfg.EnableContentBasedDetection {
		t.Error("Default config should have EnableContentBasedDetection=true")
	}
	if len(cfg.TestHelperPatterns) == 0 {
		t.Error("Default config should have non-empty TestHelperPatterns")
	}

	// Test convertSettings function with user-configurable options
	settings := &intestonly.IntestOnlySettings{
		Debug:               intestonly.BoolPtr(true),
		OverrideIsCodeFiles: []string{"custom_ignore"},
		OverrideIsTestFiles: []string{"custom_test.go"},
	}

	customCfg := intestonly.ConvertSettings(settings)

	// Check that user-configurable options are respected
	if !customCfg.Debug {
		t.Error("Custom config should have Debug=true")
	}
	if len(customCfg.OverrideIsCodeFiles) != 1 || customCfg.OverrideIsCodeFiles[0] != "custom_ignore" {
		t.Errorf("Custom config should have OverrideIsCodeFiles=[\"custom_ignore\"], got %v", customCfg.OverrideIsCodeFiles)
	}
	if len(customCfg.OverrideIsTestFiles) != 1 || customCfg.OverrideIsTestFiles[0] != "custom_test.go" {
		t.Errorf("Custom config should have OverrideIsTestFiles=[\"custom_test.go\"], got %v", customCfg.OverrideIsTestFiles)
	}

	// Check that hardcoded options remain unchanged
	if !customCfg.CheckMethods {
		t.Error("Custom config should have CheckMethods=true")
	}
	if !customCfg.ExcludeTestHelpers {
		t.Error("Custom config should have ExcludeTestHelpers=true")
	}
	if !customCfg.EnableContentBasedDetection {
		t.Error("Custom config should have EnableContentBasedDetection=true")
	}
}
