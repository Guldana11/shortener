package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := "." // корень проекта
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return processDir(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// processDir сканирует пакет и генерирует reset.gen.go
func processDir(dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_gen.go")
	}, parser.ParseComments)
	if err != nil {
		return err
	}

	for pkgName, pkg := range pkgs {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

		methodsGenerated := false
		hasMap := false

		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}

					// проверяем комментарий // generate:reset
					if !hasGenerateResetComment(gen.Doc) {
						continue
					}

					methodCode, usesMap := generateResetMethod(ts.Name.Name, st)
					if usesMap {
						hasMap = true
					}
					buf.WriteString(methodCode + "\n\n")
					methodsGenerated = true
				}
			}
		}

		if methodsGenerated {
			if hasMap {
				buf.WriteString(mapClearFunc() + "\n")
			}

			outFile := filepath.Join(dir, "reset.gen.go")
			if err := os.WriteFile(outFile, buf.Bytes(), 0644); err != nil {
				return err
			}
			fmt.Println("Generated:", outFile)
		}
	}

	return nil
}

func hasGenerateResetComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.TrimSpace(c.Text) == "// generate:reset" {
			return true
		}
	}
	return false
}

// generateResetMethod создаёт метод Reset() для структуры
func generateResetMethod(name string, st *ast.StructType) (string, bool) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("func (s *%s) Reset() {\n", name))
	buf.WriteString("\tif s == nil {\n\t\treturn\n\t}\n\n")

	usesMap := false

	for _, f := range st.Fields.List {
		for _, fname := range f.Names {
			fieldCode, isMap := generateFieldReset(fname.Name, f.Type)
			buf.WriteString("\t" + fieldCode + "\n")
			if isMap {
				usesMap = true
			}
		}
	}

	buf.WriteString("}")
	return buf.String(), usesMap
}

// generateFieldReset создаёт код сброса конкретного поля
func generateFieldReset(fieldName string, expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "complex64", "complex128":
			return fmt.Sprintf("s.%s = 0", fieldName), false
		case "string":
			return fmt.Sprintf("s.%s = \"\"", fieldName), false
		case "bool":
			return fmt.Sprintf("s.%s = false", fieldName), false
		default:
			return fmt.Sprintf("if r, ok := s.%s.(interface{ Reset() }); ok {\n\t\tr.Reset()\n\t}", fieldName), false
		}
	case *ast.StarExpr:
		// указатель на структуру
		switch inner := t.X.(type) {
		case *ast.Ident:
			// встроенный тип через указатель
			switch inner.Name {
			case "int", "int8", "int16", "int32", "int64",
				"uint", "uint8", "uint16", "uint32", "uint64",
				"float32", "float64", "complex64", "complex128":
				return fmt.Sprintf("*s.%s = 0", fieldName), false
			case "string":
				return fmt.Sprintf("*s.%s = \"\"", fieldName), false
			case "bool":
				return fmt.Sprintf("*s.%s = false", fieldName), false
			default:
				return fmt.Sprintf("s.%s.Reset()", fieldName), false
			}
		default:
			// вложенные структуры через указатель
			return fmt.Sprintf("s.%s.Reset()", fieldName), false
		}
	case *ast.ArrayType:
		return fmt.Sprintf("s.%s = s.%s[:0]", fieldName, fieldName), false
	case *ast.MapType:
		return fmt.Sprintf("clear(s.%s)", fieldName), true
	case *ast.SelectorExpr:
		return fmt.Sprintf("s.%s = %s{}", fieldName, exprString(expr)), false
	default:
		return fmt.Sprintf("// TODO reset %s", fieldName), false
	}
}

// exprString возвращает исходный код выражения
func exprString(expr ast.Expr) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), expr)
	return buf.String()
}

// mapClearFunc возвращает код функции clear для карт
func mapClearFunc() string {
	return `func clear[K comparable, V any](m map[K]V) {
	for k := range m {
		delete(m, k)
	}
}`
}

// Пример структуры
// generate:reset
type URLMapping struct {
	ID      int
	URL     string
	Tags    []string
	Options map[string]string
	Child   *URLMapping
}
