package intestonly_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(wd))), "testdata")
	analysistest.Run(t, testdata, intestonly.Analyzer, "p")

	// Test for cross-package references
	analysistest.Run(t, testdata, intestonly.Analyzer, "cross_package_ref", "cross_package_user")

	// Test for implicit usage in string literals and comments
	analysistest.Run(t, testdata, intestonly.Analyzer, "string_reference")

	// Test for implicit usage detection
	analysistest.Run(t, testdata, intestonly.Analyzer, "implicit_usage")
}

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
	if len(cfg.ExplicitTestOnlyIdentifiers) == 0 {
		t.Error("Default config should have non-empty ExplicitTestOnlyIdentifiers")
	}

	// Test convertSettings function
	settings := &intestonly.IntestOnlySettings{
		CheckMethods:                intestonly.BoolPtr(false),
		IgnoreUnexported:            intestonly.BoolPtr(true),
		EnableContentBasedDetection: intestonly.BoolPtr(false),
		TestHelperPatterns:          []string{"custom_pattern"},
		IgnoreFilePatterns:          []string{"custom_ignore"},
		ExcludePatterns:             []string{"custom_exclude"},
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
	if len(customCfg.TestHelperPatterns) != 1 || customCfg.TestHelperPatterns[0] != "custom_pattern" {
		t.Errorf("Custom config should have TestHelperPatterns=[\"custom_pattern\"], got %v", customCfg.TestHelperPatterns)
	}
	if len(customCfg.IgnoreFilePatterns) != 1 || customCfg.IgnoreFilePatterns[0] != "custom_ignore" {
		t.Errorf("Custom config should have IgnoreFilePatterns=[\"custom_ignore\"], got %v", customCfg.IgnoreFilePatterns)
	}
	if len(customCfg.ExcludePatterns) != 1 || customCfg.ExcludePatterns[0] != "custom_exclude" {
		t.Errorf("Custom config should have ExcludePatterns=[\"custom_exclude\"], got %v", customCfg.ExcludePatterns)
	}
}
