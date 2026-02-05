package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Guldana11/shortener/linter"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
