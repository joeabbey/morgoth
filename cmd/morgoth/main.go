package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "run" {
		fmt.Fprintf(os.Stderr, "usage: morgoth run <file.mor>\n")
		os.Exit(1)
	}

	filename := os.Args[2]
	_, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// TODO: lex, parse, eval
	fmt.Fprintf(os.Stderr, "morgoth: not yet implemented\n")
	os.Exit(1)
}
