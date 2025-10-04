package main

import (
	"fmt"
	"DYMS/lexer"
	"DYMS/parser"
	"DYMS/runtime"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dyms <filename.dy>")
		os.Exit(1)
	}

	filename := os.Args[1]
	
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".dy" && ext != ".dx" {
		fmt.Fprintf(os.Stderr, "Error: Only .dy and .dx files are supported (got %s)\n", ext)
		os.Exit(1)
	}
	
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