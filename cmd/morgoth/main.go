package main

import (
	"fmt"
	"os"

	"github.com/joeabbey/morgoth/internal/eval"
	"github.com/joeabbey/morgoth/internal/lexer"
	"github.com/joeabbey/morgoth/internal/parser"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "usage: morgoth run <file.mor>\n")
		os.Exit(1)
	}

	filename := os.Args[2]
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.Parse()

	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		os.Exit(1)
	}

	ev := eval.New()
	_, evalErr := ev.Eval(program)
	if evalErr != nil {
		if doomErr, ok := evalErr.(*eval.DoomError); ok {
			fmt.Fprintf(os.Stderr, "doom: %s\n", doomErr.Message)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", evalErr)
		os.Exit(1)
	}
}
