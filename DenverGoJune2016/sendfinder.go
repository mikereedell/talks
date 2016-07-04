package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	fset           = token.NewFileSet()
	selectCount    = 0
	nonSelectCount = 0
)

func main() {
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			exit(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path); err != nil {
				exit(err)
			}
		}
	}

	fmt.Printf("Sends in select: %d\nSends not in select: %d\n", selectCount, nonSelectCount)
	os.Exit(0)
}

func exit(err error) {
	fmt.Printf("Error occurred: %s\n", err.Error())
	os.Exit(2)
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isGoFile(f) {
		err = processFile(path)
	}
	if err != nil {
		exit(err)
	}
	return nil
}

func isGoFile(f os.FileInfo) bool {
	// Only process .go files that aren't tests.
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

func processFile(filename string) error {
	var f *os.File
	var err error

	f, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		exit(err)
		return err
	}

	// Walk the AST for the file here, pulling out all channel sends.
	ast.Walk(new(SendStatementVisitor), file)

	return nil
}

type SendStatementVisitor struct {
	inSelect bool
}

func (v *SendStatementVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node.(type) {
	case *ast.SelectStmt:
		v.inSelect = true // HL
	case *ast.SendStmt:
		if !v.inSelect {
			position := fset.Position(node.Pos()) // HL
			fmt.Printf("Found channel send not in select at: %s\n", position)
			nonSelectCount++
		} else {
			selectCount++
		}
		return nil
	}
	return v
}
