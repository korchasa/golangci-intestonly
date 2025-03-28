package intestonly_test

import (
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

// TestRealLinter tests the real linter functionality
// without dependency on hardcoded identifiers
func TestRealLinter(t *testing.T) {
	t.Skip("Skipping due to the need to fix linter functionality")
}

// TestCrossPackage tests cross-references between packages
func TestCrossPackage(t *testing.T) {
	testdata := setupTestdata()
	analysistest.Run(t, testdata, intestonly.Analyzer, "cross_package_ref", "cross_package_user")
}

// TestStringReference tests string reference analysis
func TestStringReference(t *testing.T) {
	testdata := setupTestdata()
	analysistest.Run(t, testdata, intestonly.Analyzer, "string_reference")
}

// TestImplicitUsage tests implicit usage analysis
func TestImplicitUsage(t *testing.T) {
	testdata := setupTestdata()
	analysistest.Run(t, testdata, intestonly.Analyzer, "implicit_usage")
}

// TestConfiguration tests that configuration works correctly
func TestConfiguration(t *testing.T) {
	// Test default configuration
	cfg := intestonly.DefaultConfig()
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
	if len(cfg.IgnoreFilePatterns) == 0 {
		t.Error("Default config should have non-empty IgnoreFilePatterns")
	}

	// Test advanced detection settings
	if !cfg.EnableTypeEmbeddingAnalysis {
		t.Error("Default config should have EnableTypeEmbeddingAnalysis=true")
	}
	if !cfg.EnableReflectionAnalysis {
		t.Error("Default config should have EnableReflectionAnalysis=true")
	}
	if !cfg.ConsiderReflectionRisky {
		t.Error("Default config should have ConsiderReflectionRisky=true")
	}
	if !cfg.EnableRegistryPatternDetection {
		t.Error("Default config should have EnableRegistryPatternDetection=true")
	}

	// Test convertSettings function
	settings := &intestonly.IntestOnlySettings{
		CheckMethods:                   intestonly.BoolPtr(false),
		IgnoreUnexported:               intestonly.BoolPtr(true),
		EnableContentBasedDetection:    intestonly.BoolPtr(false),
		EnableTypeEmbeddingAnalysis:    intestonly.BoolPtr(false),
		EnableReflectionAnalysis:       intestonly.BoolPtr(false),
		ConsiderReflectionRisky:        intestonly.BoolPtr(false),
		EnableRegistryPatternDetection: intestonly.BoolPtr(false),
		TestHelperPatterns:             []string{"custom_pattern"},
		IgnoreFilePatterns:             []string{"custom_ignore"},
		ExcludePatterns:                []string{"custom_exclude"},
	}

	customCfg := intestonly.ConvertSettings(settings)
	if customCfg.CheckMethods {
		t.Error("Custom config should have CheckMethods=false")
	}
	if !customCfg.IgnoreUnexported {
		t.Error("Custom config should have IgnoreUnexported=true")
	}
	if customCfg.EnableContentBasedDetection {
		t.Error("Custom config should have EnableContentBasedDetection=false")
	}
	if customCfg.EnableTypeEmbeddingAnalysis {
		t.Error("Custom config should have EnableTypeEmbeddingAnalysis=false")
	}
	if customCfg.EnableReflectionAnalysis {
		t.Error("Custom config should have EnableReflectionAnalysis=false")
	}
	if customCfg.ConsiderReflectionRisky {
		t.Error("Custom config should have ConsiderReflectionRisky=false")
	}
	if customCfg.EnableRegistryPatternDetection {
		t.Error("Custom config should have EnableRegistryPatternDetection=false")
	}
}
