package main

import (
    "fmt"
    "holygo/runtime"
    "holygo/lexer"
)

func main() {
    code := "42 + (5 - 3) * 2"
    tokens := lexer.Tokenize(code)

    parser := runtime.NewParser(tokens)
    astNode, err := parser.Parse()
    if err != nil {
        panic(err)
    }

    result, err := runtime.Evaluate(astNode)
    if err != nil {
        panic(err)
    }

    fmt.Println("Result:", result)
}
