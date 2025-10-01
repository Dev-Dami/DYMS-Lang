# HolyGo Interpreter

HolyGo is a lightweight, embeddable interpreter for a simple, dynamically-typed scripting language. It is written in Go and is designed to be easily integrated into other Go applications.

## Features

*   **Dynamic Typing**: The language is dynamically typed, which means that you don't have to declare the type of a variable before you use it.
*   **First-Class Functions**: Functions are first-class citizens, which means that they can be passed as arguments to other functions, returned from functions, and assigned to variables.
*   **Closures**: The language supports closures, which means that a function can access the variables of its enclosing scope, even after the enclosing function has returned.
*   **Garbage Collection**: The language has a built-in garbage collector, which automatically frees memory that is no longer in use.

## Architecture

The interpreter is divided into the following components:

*   **Lexer**: The lexer is responsible for breaking the source code into a stream of tokens.
*   **Parser**: The parser takes the tokens from the lexer and builds an Abstract Syntax Tree (AST).
*   **Compiler**: The compiler takes the AST and compiles it into a series of bytecode instructions.
*   **Virtual Machine**: The virtual machine executes the bytecode instructions.

## Getting Started

To get started with HolyGo, you will need to have Go installed on your system. You can then build the interpreter by running the following command:

```bash
go build
```

This will create an executable file called `holygo` in the current directory. You can then run the interpreter by running the following command:

```bash
./holygo
```

This will start the interactive Read-Eval-Print Loop (REPL), where you can enter and execute HolyGo code.

## Roadmap

*   **Add support for more data types**: The language currently only supports numbers and strings. We plan to add support for more data types, such as booleans, arrays, and maps.
*   **Improve the error handling**: The error handling is currently very basic. We plan to improve the error handling to make it more user-friendly.
*   **Add a standard library**: We plan to add a standard library that will provide a set of useful functions for working with strings, numbers, and other data types.