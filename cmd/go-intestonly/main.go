package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/korchasa/golangci-intestonly/pkg/golinters/intestonly"
)

func main() {
	singlechecker.Main(intestonly.Analyzer)
}
