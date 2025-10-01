package main

import (
    "fmt"
    "holygo/runtime"
    "holygo/lexer"
    "holygo/parser"
    "io/ioutil"
    "os"
)

func main() {
    var filename string
    if len(os.Args) > 1 {
        filename = os.Args[1]
    } else {
        filename = "test.hg"
    }

    code, err := ioutil.ReadFile(filename)
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
