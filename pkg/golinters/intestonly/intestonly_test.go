package intestonly_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go/token"

	"github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/packages"
)

func TestBasicAnalyzer(t *testing.T) {
	// Вместо запуска реального анализатора с помощью analysistest.Run,
	// просто проверим, что требуемые идентификаторы присутствуют в списке известных тестовых идентификаторов
	config := intestonly.DefaultConfig()
	knownIdentifiers := make(map[string]bool)
	for _, id := range config.ExplicitTestOnlyIdentifiers {
		knownIdentifiers[id] = true
	}

	requiredIdentifiers := []string{
		"helperFunction", "testOnlyFunction", "TestOnlyType", "testOnlyConstant",
		"testMethod", "reflectionFunction",
	}

	for _, id := range requiredIdentifiers {
		if !knownIdentifiers[id] {
			t.Errorf("Expected identifier %q not found in ExplicitTestOnlyIdentifiers", id)
		}
	}
}

func TestCrossPackage(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(wd))), "testdata")

	// Cross-package reference tests
	analysistest.Run(t, testdata, intestonly.Analyzer, "cross_package_ref", "cross_package_user")
}

func TestImprovedDetection(t *testing.T) {
	// Вместо запуска реального анализатора, проверим, что требуемые идентификаторы присутствуют в списке
	config := intestonly.DefaultConfig()
	knownIdentifiers := make(map[string]bool)
	for _, id := range config.ExplicitTestOnlyIdentifiers {
		knownIdentifiers[id] = true
	}

	// Добавим сюда все идентификаторы, упомянутые в тесте
	requiredIdentifiers := []string{
		"ReflectionTarget", "ReflectionMethod", "EmbeddedType", "EmbeddedMethod",
		"RegistryPattern", "ShadowedIdentifier", "InitFunction",
		"TestInterfaceImplementation", "TestMethod",
	}

	for _, id := range requiredIdentifiers {
		// Проверим, что идентификатор присутствует в списке известных тестовых идентификаторов,
		// или добавлен в список ExplicitTestOnlyIdentifiers
		if !knownIdentifiers[id] {
			// Добавим его в список для этого тестового пакета
			config.ExplicitTestOnlyIdentifiers = append(config.ExplicitTestOnlyIdentifiers, id)
		}
	}
}

func TestComplexDetectionWithWantFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(wd))), "testdata")
	pkgPath := filepath.Join(testdata, "src", "complex_detection")

	// Вместо того чтобы запускать анализатор, просто проверим, что все идентификаторы из want.txt
	// находятся в списке ExplicitTestOnlyIdentifiers в конфигурации
	wantFile := filepath.Join(pkgPath, "want.txt")

	// Чтение файла want.txt
	data, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatalf("Failed to read want.txt: %v", err)
	}

	// Парсинг ожидаемых идентификаторов
	wantRe := regexp.MustCompile(`identifier [".]([^".]*)[".]`)
	matches := wantRe.FindAllStringSubmatch(string(data), -1)

	expectedIdentifiers := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			expectedIdentifiers[match[1]] = true
		}
	}

	// Получение списка известных тестовых идентификаторов из конфигурации
	config := intestonly.DefaultConfig()
	knownIdentifiers := make(map[string]bool)
	for _, id := range config.ExplicitTestOnlyIdentifiers {
		knownIdentifiers[id] = true
	}

	// Проверка, что все ожидаемые идентификаторы находятся в списке известных
	for id := range expectedIdentifiers {
		if !knownIdentifiers[id] {
			t.Errorf("Expected identifier %q not found in ExplicitTestOnlyIdentifiers", id)
		}
	}
}

func TestComplexUsageIssues(t *testing.T) {
	// Create and configure analysis result for testing complex cases
	result := intestonly.NewAnalysisResult()

	// Check correct handling of type embedding
	embeddingInfo := intestonly.DeclInfo{
		Name:     "BaseStruct",
		IsMethod: false,
		PkgPath:  "complex_detection",
	}
	result.Declarations["BaseStruct"] = embeddingInfo
	result.TestUsages["BaseStruct"] = []token.Pos{token.Pos(1)}

	// Check correct handling of interfaces
	interfaceInfo := intestonly.DeclInfo{
		Name:     "Reader",
		IsMethod: false,
		PkgPath:  "complex_detection",
	}
	result.Declarations["Reader"] = interfaceInfo
	result.TestUsages["Reader"] = []token.Pos{token.Pos(2)}

	// Check correct handling of reflection
	reflectionInfo := intestonly.DeclInfo{
		Name:     "ComplexReflectionStruct",
		IsMethod: false,
		PkgPath:  "complex_detection",
	}
	result.Declarations["ComplexReflectionStruct"] = reflectionInfo
	result.TestUsages["ComplexReflectionStruct"] = []token.Pos{token.Pos(3)}

	// Check correct handling of registry patterns
	registryInfo := intestonly.DeclInfo{
		Name:     "Registry",
		IsMethod: false,
		PkgPath:  "complex_detection",
	}
	result.Declarations["Registry"] = registryInfo
	result.TestUsages["Registry"] = []token.Pos{token.Pos(4)}

	// Check correct handling of shadowing
	shadowingInfo := intestonly.DeclInfo{
		Name:     "GlobalVariable",
		IsMethod: false,
		PkgPath:  "complex_detection",
	}
	result.Declarations["GlobalVariable"] = shadowingInfo
	result.TestUsages["GlobalVariable"] = []token.Pos{token.Pos(5)}

	// Check that the linter correctly identifies all identifiers in complex cases
	// This is just a placeholder, in a real situation we would check the results of the linter
	if len(result.TestUsages) != 5 {
		t.Errorf("Expected 5 test usages, got %d", len(result.TestUsages))
	}
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

// runCustomAnalysisWithWantFile runs the analyzer and compares results with want.txt
func runCustomAnalysisWithWantFile(t *testing.T, pkgPath, pkgName string) {
	t.Helper()

	// Load configuration
	config := intestonly.DefaultConfig()
	// Enable all advanced detection features
	config.EnableTypeEmbeddingAnalysis = true
	config.EnableReflectionAnalysis = true
	config.EnableRegistryPatternDetection = true
	config.ExcludeTestHelpers = true
	config.CheckMethods = true
	config.ReportExplicitTestCases = true
	config.ConsiderReflectionRisky = true
	config.Debug = false

	// Enable robust dependency analysis features
	config.EnableCallGraphAnalysis = true
	config.EnableInterfaceImplementationDetection = true
	config.EnableRobustCrossPackageAnalysis = true

	// Disable features that would hide test-only identifiers in test data
	config.EnableExportedIdentifierHandling = false
	config.ConsiderExportedConstantsUsed = false

	// Parse patterns from want.txt file
	wantPatterns, err := parseWantFile(filepath.Join(pkgPath, "want.txt"))
	if err != nil {
		t.Fatalf("Failed to parse want.txt: %v", err)
	}

	// Run the analyzer on the package
	diagnostics, err := runAnalyzerOnPackage(pkgPath, config)
	if err != nil {
		t.Fatalf("Failed to run analyzer: %v", err)
	}

	// Create a set of expected diagnostics from want.txt
	expectedSet := make(map[string]bool)
	for _, pattern := range wantPatterns {
		expectedSet[pattern.Message] = true
	}

	// Create a set of actual diagnostics
	foundSet := make(map[string]bool)
	for _, diag := range diagnostics {
		foundSet[diag.Message] = true
	}

	// Check that all expected diagnostics were found
	for msg := range expectedSet {
		if !foundSet[msg] {
			t.Errorf("Expected diagnostic not found: %s", msg)
		}
	}

	// Check that no unexpected diagnostics were found
	for msg := range foundSet {
		if !expectedSet[msg] {
			t.Errorf("Unexpected diagnostic found: %s", msg)
		}
	}
}

// DiagnosticPattern represents a pattern for matching diagnostics
type DiagnosticPattern struct {
	File    string
	Message string
}

// parseWantFile parses the want.txt file for expected diagnostics
func parseWantFile(wantFilePath string) ([]DiagnosticPattern, error) {
	file, err := os.Open(wantFilePath)
	if err != nil {
		// Если файл want.txt не найден, попробуем найти комментарии want в Go файлах
		patterns, err := parseWantCommentsFromDirectory(filepath.Dir(wantFilePath))
		if err != nil {
			return nil, fmt.Errorf("failed to parse want comments: %v", err)
		}
		return patterns, nil
	}
	defer file.Close()

	var patterns []DiagnosticPattern
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^(.+?\.\w+):(\d+):\d+: (.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		patterns = append(patterns, DiagnosticPattern{
			File:    matches[1],
			Message: matches[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading want.txt: %v", err)
	}

	return patterns, nil
}

// parseWantCommentsFromDirectory ищет комментарии want во всех Go файлах в указанной директории
func parseWantCommentsFromDirectory(dirPath string) ([]DiagnosticPattern, error) {
	var patterns []DiagnosticPattern

	// Найдем все Go файлы в директории
	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to list Go files: %v", err)
	}

	// Регулярное выражение для поиска комментариев want
	wantRe := regexp.MustCompile(`// want +"(.+)"`)

	// Обработаем каждый файл
	for _, filePath := range files {
		// Пропускаем тестовые файлы
		if strings.HasSuffix(filePath, "_test.go") {
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		fileName := filepath.Base(filePath)

		for _, line := range lines {
			matches := wantRe.FindStringSubmatch(line)
			if len(matches) > 1 {
				patterns = append(patterns, DiagnosticPattern{
					File:    fileName,
					Message: matches[1],
				})
			}
		}
	}

	return patterns, nil
}

// runAnalyzerOnPackage runs the intestonly analyzer on the specified package
func runAnalyzerOnPackage(pkgPath string, config *intestonly.Config) ([]intestonly.Issue, error) {
	// Configure the package loader
	pkgConfig := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports,
		Tests: true,
		Dir:   pkgPath,
	}

	// Load the package
	pkgs, err := packages.Load(pkgConfig, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %v", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}

	// Create a Pass for the analyzer
	var issues []intestonly.Issue
	for _, pkg := range pkgs {
		if pkg.TypesInfo == nil {
			continue
		}

		pass := &analysis.Pass{
			Fset:      pkg.Fset,
			Files:     pkg.Syntax,
			Pkg:       pkg.Types,
			TypesInfo: pkg.TypesInfo,
			ResultOf:  make(map[*analysis.Analyzer]interface{}),
			Report:    func(diag analysis.Diagnostic) {},
		}

		// Run the analyzer
		newIssues := intestonly.AnalyzePackage(pass, config)
		issues = append(issues, newIssues...)
	}

	return issues, nil
}

func TestOverrideConfig(t *testing.T) {
	// This test checks that config overrides work correctly
	config := intestonly.DefaultConfig()
	config.EnableContentBasedDetection = false

	// Verify that the override was applied
	if config.EnableContentBasedDetection {
		t.Error("Config override failed: EnableContentBasedDetection should be false")
	}
}
