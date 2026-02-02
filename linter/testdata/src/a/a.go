package a

import (
	"log"
	"os"
)

func f() {
	panic("boom")    // want "usage of panic"
	log.Fatal("die") // want "log.Fatal is allowed only in main.main"
	os.Exit(1)       // want "os.Exit is allowed only in main.main"
}
