package main

import (
    "fmt"
    "holygo/runtime"
    "holygo/lexer"
    "holygo/parser"
)

func main() {
    code := "42 + (5 - 3) * 2"
    tokens := lexer.Tokenize(code)

    parser := parser.New(tokens)
    astNode := parser.ParseProgram()

    result, err := runtime.Evaluate(astNode)
    if err != nil {
        panic(err)
    }

    fmt.Println("Result:", result)
}
