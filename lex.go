package main

import (
	"os"
	"flag"
	"io"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"thequux/dsview"
)

var (
	package_name = flag.String("package", "main", "Output package name")
	output_file  = flag.String("output", "l.go", "Output filename")
	name_prefix  = flag.String("prefix", "yy", "Prefix prepended to generated names")
)

func init() {
	flag.Parse()
}

func main() {
	var input io.Reader
	var err os.Error
	var filename string
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
		if input, err = os.Open(flag.Arg(0)); err != nil {
			fmt.Print("Failed to open input: ", err)
			os.Exit(1)
		}
	} else {
		input = os.Stdin
		filename = "stdin"
	}
	
	var src_ast *ast.File
	fset := token.NewFileSet()
	src_ast, err = parser.ParseFile(fset, filename, input, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	
	//Transform(src_ast)
	//dfa := CreateAutomata(source)
	dsview.QuickServe(src_ast)
	outfile, err := os.Create(*output_file)
	if err != nil {
		fmt.Print("Failed to open output file: ", err)
		os.Exit(1)
	}
	if  err := printer.Fprint(outfile, fset, src_ast); err != nil {
		fmt.Print("Write failed: ", err)
		os.Exit(1)
	}
}

/*
type ToplevelVisitor int

func walkStmtList(v Visitor, list []ast.Stmt) {
	res = make([]ast.Stmt, 0, cap(list)
	for i, x := 

func (*ToplevelVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.BlockStmt:
		a
	case *ast.SwitchStmt:
		if node.Init == 
*/