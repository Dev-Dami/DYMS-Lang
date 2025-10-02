package main

import (
	"fmt"
	"DYMS/lexer"
	"DYMS/parser"
	"DYMS/runtime"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hg <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]
	sourceCode, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	tokens := lexer.Tokenize(string(sourceCode))
	parser := parser.New(tokens)
	program := parser.ParseProgram()

	// Create a new environment for the program
	env := runtime.GlobalEnv

	_, err = runtime.Evaluate(program, env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}


