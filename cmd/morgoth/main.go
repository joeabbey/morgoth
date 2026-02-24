package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joeabbey/morgoth/internal/eval"
	"github.com/joeabbey/morgoth/internal/lexer"
	"github.com/joeabbey/morgoth/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: morgoth <command> [args]\ncommands: run <file.mor>, repl\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: morgoth run <file.mor>\n")
			os.Exit(1)
		}
		runFile(os.Args[2])
	case "repl":
		runRepl()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: morgoth <command> [args]\ncommands: run <file.mor>, repl\n", os.Args[1])
		os.Exit(1)
	}
}

func runFile(filename string) {
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

func runRepl() {
	scanner := bufio.NewScanner(os.Stdin)
	ev := eval.New()

	fmt.Println("Morgoth REPL (type 'exit' or Ctrl+D to quit)")
	for {
		fmt.Print("morgoth> ")
		if !scanner.Scan() {
			fmt.Println()
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.Parse()

		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
			}
			continue
		}

		result, err := ev.Eval(program)
		if err != nil {
			if doomErr, ok := err.(*eval.DoomError); ok {
				fmt.Fprintf(os.Stderr, "doom: %s\n", doomErr.Message)
			} else {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			continue
		}

		// Print non-nil results for expression evaluation feedback
		if result != nil && result.Kind != eval.ValNil {
			fmt.Println(result.String())
		}
	}
}
