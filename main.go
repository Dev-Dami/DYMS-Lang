package main

import (
	"fmt"
	"holygo/lexer"
	"holygo/parser"
	"holygo/ast"
)

func main() {
	code := "42 + (5 - 3)"
	tokens := lexer.Tokenize(code)

	fmt.Println("Tokens:")
	for _, t := range tokens {
		fmt.Printf("%s: %s\n", t.Type, t.Value)
	}

	p := parser.New(tokens)
	prog := p.ParseProgram()

	fmt.Println("\nAST Root:", prog.Kind())
	fmt.Println("Pretty AST:")
	for _, stmt := range prog.Body {
		if expr, ok := stmt.(ast.Expr); ok {
			fmt.Println(ast.PrettyPrint(expr))
		}
	}
}
