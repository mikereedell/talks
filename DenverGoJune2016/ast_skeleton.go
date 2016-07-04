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
	// token.FileSet is a set of token.File.
	// token.File is a lexical representation of the source file.
	fset = token.NewFileSet() // HL

	// Any additional global state like counters, etc.
)

// Entry point
func main() {
	flag.Parse()
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			exit(err)
		case dir.IsDir():
			walkDir(path) // HL
		default:
			if err := processFile(path); err != nil { // HL
				exit(err)
			}
		}
	}

	// Report findings here.
	os.Exit(0)
}

func exit(err error) {
	fmt.Printf("Error occurred: %s\n", err.Error())
	os.Exit(2)
}

// Walk the directory structure
func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

// Function used when walking the directory structure.
func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isGoFile(f) {
		err = processFile(path)
	}
	if err != nil {
		exit(err)
	}
	return nil
}

// Filter out source files. Only process '.go' files that aren't tests.
func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") &&
		strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

// Function called when a suitable '.go' file is visited.
func processFile(filename string) error {
	var f *os.File
	var err error

	// Open source file
	f, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the entire source file into a []byte
	src, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// Use the go/parser package to generate the AST
	// Args are:
	//  - state of files already parsed (*token.FileSet)
	//  - filename of file to parse (string)
	//  - source of file to parse ({}interface but one of string, []byte, io.Reader)
	//  - parsing mode (package only, import only, comments, trace, etc)
	file, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		exit(err)
		return err
	}

	// Walk the AST using a custom node visitor, in depth-first order.
	ast.Walk(new(CustomVisitor), file)

	return nil
}

// Create a custom ast.Visitor to hold any state needed while walking the AST.
type CustomVisitor struct {
	//State
}

// Implement the Visit(node Node) method on the ast.Visit interface.
// Invoked for every node encountered by ast.Walk(...)
func (v *CustomVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node.(type) {
	case *ast.SelectStmt:
		// Process 'select' statements.
	case *ast.IfStmt:
		// Process 'if' statements.
		// etc.
	}
	return v
}
