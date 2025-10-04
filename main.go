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
	sourceCode, readErr := ioutil.ReadFile(filename)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", readErr)
		os.Exit(1)
	}

	tokens := lexer.Tokenize(string(sourceCode))
	p := parser.New(tokens)
	program, perr := p.ParseProgram()
	if perr != nil {
		fmt.Fprintln(os.Stderr, perr.Error())
		os.Exit(1)
	}

	// env -> for program
	env := runtime.GlobalEnv

	// Use hybrid execution engine (VM for functions, interpreter for complex operations)
	hybrid := runtime.NewHybridEngine(env)
	_, rerr := hybrid.Execute(program)
	if rerr != nil {
		fmt.Fprintln(os.Stderr, rerr.Error())
		os.Exit(1)
	}
}