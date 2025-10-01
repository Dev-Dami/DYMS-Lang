package main

import (
    "fmt"
    "holygo/interpreter"
    "holygo/lexer"
)

func main() {
    code := "42 + (5 - 3) * 2"
    tokens := lexer.Tokenize(code)

    parser := interpreter.NewParser(tokens)
    astNode, err := parser.Parse()
    if err != nil {
        panic(err)
    }

    result, err := interpreter.Evaluate(astNode)
    if err != nil {
        panic(err)
    }

    fmt.Println("Result:", result)
}
