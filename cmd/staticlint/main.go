// Command staticlint запускает кастомный статический анализатор кода.
//
// Использование:
//
//	staticlint ./...
//	staticlint ./internal/...
//
// Анализатор состои из:
//
// shadow.Analyzer пакета golang.org/x/tools/go/analysis/passes;
// всех анализаторов класса SA пакета staticcheck.io;
// ST1000 пакета staticcheck.io;
// NoOsExitInMain - кастомная реализация которая запрещает использование os.Exit в функции main пакета main
package main

import (
	ma "github.com/as-tanais/observy/cmd/staticlint/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
			break
		}
	}

	analyzers = append(analyzers, shadow.Analyzer)

	analyzers = append(analyzers, ma.NoOsExitInMain)

	multichecker.Main(analyzers...)
}
