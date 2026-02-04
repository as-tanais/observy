package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitInMain запрещающет использовать прямой вызов os.Exit в функции main пакета main
//
// Пример ошибки:
//
//	package main
//
//	import "os"
//
//	func main() {
//	    os.Exit(1)
//	}
var NoOsExitInMain = &analysis.Analyzer{
	Name: "noosexitinmain",
	Doc:  "forbids os.Exit calls inside main function of main package",
	Run:  runNoOsExitInMain,
}

func runNoOsExitInMain(pass *analysis.Pass) (interface{}, error) {

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok || pkgIdent.Name != "os" {
					return true
				}

				if sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit is forbidden in main function; return error instead")
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
