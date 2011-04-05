package golex

import (
	"os"
	"flag"
	"io"
	"fmt"
	"go/printer"
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
	if flag.NArg() > 1 {
		filename = flag.Arg(0)
		if input, err = os.Open(flag.Arg(0), os.O_RDONLY, 0666); err != nil {
			fmt.Print("Failed to open input: ", err)
			os.Exit(1)
		}
	} else {
		input = os.Stdin
		filename = "stdin"
	}

	source := ParseFile(filename, input)

	dfa := CreateAutomata(source)
	out_ast = GenerateOutput(dfa)
	
	outfile, err := os.Open(*output_file, os.O_CREATE | O_WRONLY, 0666)
	if err != nil {
		fmt.Print("Failed to open output file: ", err)
		os.Exit(1)
	}
	if _, err := fmt.Fprint(outfile, nil, out_ast); err != nil {
		fmt.Print("Write failed: ", err)
		os.Exit(1)
	}
}
