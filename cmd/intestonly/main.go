package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// Main entry point for the intestonly analyzer
// Usage: go run ./cmd/intestonly/main.go ./...
func main() {
	log.SetPrefix("intestonly: ")
	log.SetFlags(0)

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		log.Fatalf("No packages specified")
	}

	// Load the packages
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedDeps,
		Tests:     true,
		ParseFile: nil,
		Logf: func(format string, args ...interface{}) {
			if false { // Установите в true для отладки
				log.Printf(format, args...)
			}
		},
	}

	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	// Run the analyzer
	results, err := checker.Analyze([]*analysis.Analyzer{intestonly.Analyzer}, pkgs, nil)
	if err != nil {
		log.Fatalf("Error running analyzer: %v", err)
	}

	// Инициализируем exitCode здесь
	exitCode := 0

	// Печатаем любые ошибки, связанные с загрузкой пакетов
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, err := range pkg.Errors {
				log.Printf("Error loading package %s: %v", pkg.ID, err)
			}
			exitCode = 1
		}
	}

	// Print results
	for _, act := range results.Roots {
		if act.Err != nil {
			log.Printf("Error analyzing %s: %v", act.Package.ID, act.Err)
			exitCode = 1
			continue
		}

		for _, diag := range act.Diagnostics {
			pos := act.Package.Fset.Position(diag.Pos)
			fmt.Printf("%s: %s\n", pos, diag.Message)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
