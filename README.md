# DYMS — Dynamic Yet Minimal System Interpreter

DYMS is a lightweight, embeddable interpreter for a simple, dynamically-typed scripting language. It is written in Go and is designed to be easily integrated into other Go applications.

## Version

update 0.1 demo

## Features

- **Dynamic Typing**: The language is dynamically typed, which means that you don't have to declare the type of a variable before you use it.
- **Variables**: The language supports variable declarations using the `let`, `var`, and `const` keywords.
- **Data Types**: Supports numbers (float64) and strings.
- **Operators**: Basic arithmetic operators: `+`, `-`, `*`, `/`.
- **Comments**: Supports single-line comments using `//`.
- **Built-in Functions**:
  - `systemout(...)`: Prints arguments to the console using the log package.
  - `println(...)`: Prints arguments to the console.
  - `printf(format, ...)`: Prints formatted strings.
  - `logln(...)`: Prints arguments to the console using the log package.
- **Control Flow**: support for `if/else` statements and `for` loops.

## Architecture

The interpreter is divided into the following components:

- **Lexer**: The lexer is responsible for breaking the source code into a stream of tokens.
- **Parser**: The parser takes the tokens from the lexer and builds an Abstract Syntax Tree (AST).
- **Evaluator**: The evaluator traverses the AST and evaluates the code.

## Getting Started

To get started with DYMS, you will need to have Go installed on your system. You can then run the interpreter by running the following command:

```bash
go run . test.hg
```

This will execute the code in the `test.hg` file.

## Usage/Examples

Here is an example of a HolyGo program:

```go
let x = 10
var y = 20
const z = 30

// This is a comment
systemout(x + y + z)
println("Hello from println")
printf("x = %d, y = %d, z = %d", x, y, z)
logln("This is a log message")
logln(x + 2)
```

## Recent Changes

- Added support for `//` comments.
- Improved the output of the `println`, `printf`, and `logln` functions.
- Fixed a bug in the `for` loop where the loop variable was being redeclared on each iteration.
- Fixed a bug in the lexer where the division operator was not being handled correctly.

## Error Handling

The interpreter will report errors for various issues, including:

- **Syntax Errors**: Such as unrecognized characters or malformed statements.
- **Runtime Errors**: Such as division by zero, or trying to use a variable that has not been declared.

## Roadmap

- **Add support for more data types**: Booleans, arrays, and maps.
- **Improve the error handling**: Make error messages more user-friendly.
- **Add a standard library**: Provide a set of useful functions for working with strings, numbers, and other data types.
