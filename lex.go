package main

import (
	"os"
	"flag"
	"io"
	"fmt"
	"go/ast"
	"go/parser"
	//"go/printer"
	"go/token"
	"log"
	"runtime"
	"strconv"
	//"thequux/dsview"
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

	{
		temp := "a⍨c"
		fmt.Printf("Len: %v\n", len(temp))
		for i := 0; i < len(temp); i++ {
			fmt.Printf("%3d: %02x\n", i, temp[i])
		}
	}
	
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
	src_ast, err = parser.ParseFile(fset, filename, input, parser.ParseComments)// | parser.Trace)
	if err != nil {
		panic(err)
	}
	
	Transform(src_ast)
	//dfa := CreateAutomata(source)
	
	outfile, err := os.Create(*output_file)
	if err != nil {
		fmt.Print("Failed to open output file: ", err)
		os.Exit(1)
	}
	//if  err := printer.Fprint(outfile, fset, src_ast); err != nil {
	//fmt.Printf("%#v\n", src_ast)
	if  _, err := ast.Fprint(outfile , fset, src_ast, nil); err != nil {
		fmt.Print("Write failed: ", err)
		os.Exit(1)
	}
	outfile.Close()
}

func Transform(src *ast.File) *ast.File {
	// find the golex import...
	var toplevel ToplevelVisitor
	
	for _, spec := range src.Imports {
		ast.Print(nil, spec)
		if conv, err := strconv.Unquote(spec.Path.Value); err != nil || conv != "golex" {
			continue
		}
		if spec.Name != nil {
			toplevel.golexPackage = spec.Name.Name
		} else {
			toplevel.golexPackage = "golex"
		}
		break
	}
	if toplevel.golexPackage == "" {
		log.Print("Not a lex input")

		return src
	}

	ast.Walk(&toplevel, src)
	// Now, find switch statemnts...
	return src
}


type ToplevelVisitor struct {
	golexPackage string
}

type golexIdentifyingVisitor struct {
	golexPackage string
	istream *ast.Expr
	call *ast.CallExpr
	varName *ast.Expr // identifier that gets bound to the result token
}

func (self *golexIdentifyingVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AssignStmt:
		if len(n.Rhs) != 1 && len(n.Lhs) != len(n.Rhs) {
			return nil
		}
		for i, expr := range n.Rhs {
			defer func() {
				if err := recover(); err != nil {
					_ = err.(*runtime.TypeAssertionError);
					return
				}
			}()
			if call, ok := expr.(*ast.CallExpr); ok {
				fun := call.Fun.(*ast.SelectorExpr)
				if fun.X.(*ast.Ident).Name != self.golexPackage ||
					fun.Sel.Name != "Token" ||
					len(call.Args) != 1 { // is RHS not golex.Token(.) ?
					continue
				}

				self.call = call // substitutable select fn
				self.varName = &n.Lhs[i] // LHS identifier
				self.istream = &call.Args[0] // token stream
				return nil
			}
			
		}
	}
	return nil
}

func (self *ToplevelVisitor) walkStmtList(list []ast.Stmt) {
	res := make([]ast.Stmt, 0, cap(list))
	for _, x := range list {
		switch stmt := x.(type) {
		case *ast.SwitchStmt:
 			is_golex := &golexIdentifyingVisitor{golexPackage: self.golexPackage}
			ast.Walk(is_golex, stmt.Init)
			if is_golex.call != nil {
				// this should be converted...
				patterns := make([]string, 0, len(stmt.Body.List))
				for i := range stmt.Body.List {
					clause := stmt.Body.List[i].(*ast.CaseClause)
					if clause.List == nil {
						// default case... not handled...
						continue
					}
					for _, lit := range clause.List {
						str, err := strconv.Unquote(lit.(*ast.BasicLit).Value)
						if err != nil {
							log.Printf("WTF is %s?", lit.(*ast.BasicLit).Value)
							continue
						}
						patterns = append(patterns, str)
					}
				}
				for i, pat := range patterns {
					log.Printf("%d: (in) %v", i, strconv.Quote(pat))
					log.Printf("%d: %v", i, strconv.Quote(ParseRegex(pat).StringPrec(0)))
				}
			}
				
		default:
			res = append(res, x)
		}
	}
}
	

func (self *ToplevelVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.BlockStmt:
		self.walkStmtList(node.List)
	case *ast.SwitchStmt:
	}
	return self
}