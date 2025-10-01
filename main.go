package main

import (
    "fmt"
    "holygo/runtime"
    "holygo/lexer"
    "holygo/parser"
    "io/ioutil"
)

func main() {
    code, err := ioutil.ReadFile("test.hg")
    if err != nil {
        panic(err)
    }
    tokens := lexer.Tokenize(string(code))

    parser := parser.New(tokens)
    astNode := parser.ParseProgram()

    result, err := runtime.Evaluate(astNode)
    if err != nil {
        panic(err)
    }

    if result != nil {
        fmt.Println("Result:", result)
    }
}
