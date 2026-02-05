package linter

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "panicexit",
	Doc:  "reports usage of panic and log.Fatal/os.Exit outside main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(call.Pos(), "usage of panic is forbidden")
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok {
					if (pkg.Name == "log" && sel.Sel.Name == "Fatal") ||
						(pkg.Name == "os" && sel.Sel.Name == "Exit") {

						if file.Name.Name != "main" || !insideMainFunc(file, call.Pos()) {
							pass.Reportf(
								call.Pos(),
								"%s.%s is allowed only in main.main",
								pkg.Name,
								sel.Sel.Name,
							)
						}
					}
				}
			}

			return true
		})
	}
	return nil, nil
}

func insideMainFunc(file *ast.File, pos token.Pos) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Name.Name == "main" &&
			pos >= fn.Body.Pos() && pos <= fn.Body.End() {
			return true
		}
	}
	return false
}
