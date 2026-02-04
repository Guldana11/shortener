package linter

import (
	"go/ast"
	"go/token"
	"go/types"

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

			// panic(...)
			if isPanicCall(call) {
				pass.Reportf(call.Pos(), "usage of panic is forbidden")
				return true
			}

			// log.Fatal(...) / os.Exit(...)
			pkgPath, selName, ok := selectorPkgPathAndName(pass, call)
			if !ok {
				return true
			}

			// интересуют только log.Fatal и os.Exit
			if !(pkgPath == "log" && selName == "Fatal") && !(pkgPath == "os" && selName == "Exit") {
				return true
			}

			// разрешено только внутри main.main пакета main
			if file.Name != nil && file.Name.Name == "main" && insideMainFunc(file, call.Pos()) {
				return true
			}

			pass.Reportf(call.Pos(), "%s.%s is allowed only in main.main", pkgPath, selName)
			return true
		})
	}
	return nil, nil
}

func insideMainFunc(file *ast.File, pos token.Pos) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name == nil || fn.Name.Name != "main" {
			continue
		}
		return pos >= fn.Body.Pos() && pos <= fn.Body.End()
	}
	return false
}

func isPanicCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "panic"
}

// selectorPkgPathAndName возвращает путь пакета ("os"/"log") и имя селектора ("Exit"/"Fatal"),
func selectorPkgPathAndName(pass *analysis.Pass, call *ast.CallExpr) (pkgPath string, selName string, ok bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}

	obj := pass.TypesInfo.Uses[pkgIdent]
	if obj == nil {
		return "", "", false
	}

	pkgName, ok := obj.(*types.PkgName)
	if !ok || pkgName.Imported() == nil {
		return "", "", false
	}

	return pkgName.Imported().Path(), sel.Sel.Name, true
}
