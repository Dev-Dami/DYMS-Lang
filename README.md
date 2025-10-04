# DYMS — Dynamic Yet Minimal System Interpreter

**DYMS** is a lightweight, embeddable interpreter for a dynamic scripting language, implemented in Go. It emphasizes simplicity, extensibility, and seamless integration.

**Status:** Demo 0.5
**License:** [MIT](./LICENSE)
**Requirements:** Go ≥ 1.24

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Command-Line Usage](#command-line-usage)
- [Language Overview](#language-overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Pretty Printing](#pretty-printing)
- [Demos](#demos)
- [Roadmap](#roadmap)
- [Further Reading & Inspiration](#further-reading--inspiration)

---

## Features

- **Variables**: `let` (immutable by default), `var` (mutable), `const` (immutable with strict enforcement)
- **Data Types**: Number (float64), String, Boolean, Array (heterogeneous), Map (string keys)
- **Functions**:
  - User-defined with `funct name(params) { body }`
  - Supports closures and lexical environment capture
  - Return statements: `return value`
  - Automatic null-padding for missing arguments

- **Module System**:
  - Import modules with aliasing: `import "module" as alias`
  - Built-in `time` library: `now()`, `millis()`, `sleep()`
  - Built-in `fmaths` library: Advanced mathematical functions and constants

- **Operators**:
  - Arithmetic: `+`, `-`, `*`, `/`, `%` (handles division by zero)
  - Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
  - Logical: `&&`, `||`
  - Increment/Decrement: `++var`, `var++`, `--var`, `var--`
  - String concatenation with automatic type conversion

- **Control Flow**:
  - Conditional: `if/else`
  - Loops: `while` and `for range(i, N)` with `break` and `continue` support
  - Exception handling: `try/catch` blocks

- **Advanced Features**:
  - **Hybrid execution engine**: Smart routing between VM and interpreter based on code complexity
  - **High-performance bytecode VM**: 20+ optimized opcodes for common operations
  - **Compiler optimizations**: Peephole optimization, constant folding, dead code elimination
  - **Ultra-fast loops**: Specialized fast paths for common loop patterns (sub-150ms performance)
  - **Memory optimization**: Object pooling and reuse for runtime values
  - Property access via dot notation for maps
  - String escaping: `\n`, `\t`, `\r\n`, `\\`, `\"`
  - Single-line comments: `//`

- **Built-in Functions**:
  - I/O: `println`, `printf`, `systemout`, `logln`
  - Formatting: `pretty(v)`, `prettyml(v)`, `printlnml(v)`
  - All built-ins support variadic arguments

- **Robust Error Handling**:
  - Line/column-aware parser and runtime errors
  - Context-sensitive error reporting

---

## Quick Start

### Run a script

```powershell
go run . test.dy
```

### Build and run as binary

```powershell
# Using build script (Windows)
.\build.bat build
.\build\dyms.exe test\01_basic_features.dy

# Manual build
go build -o dyms.exe .
.\dyms.exe test\01_basic_features.dy
```

---

## Command-Line Usage

```text
dyms <filename>
```

**Examples:**

```powershell
# Run directly with Go
go run . test/01_basic_features.dy

# Build and run
.\build.bat build
.\build\dyms.exe test/21_simple_test.dy

# Run all tests
.\build.bat test
```

---

## Language Overview

### Core Syntax

```hg
let x = 10
var y = 20
const who = "DYMS"
let ok = true
let arr = [1, 2, 3, "mixed"]
let m = {"name": "DYMS", "version": 0.5, "stable": ok}

if (x > 5) { println("x > 5") } else { println("x <= 5") }
for range(i, 10) {
    if (i == 5) { break }
    if (i % 2 == 0) { continue }
    println(i)
}
while (y > 0) { --y }
println(m.name)
```

### Functions

```hg
funct greet(name) {
    return "Hello, " + name + "!"
}

let message = greet("World")
println(message)

funct makeCounter() {
    let count = 0
    funct increment() {
        ++count
        return count
    }
    return increment
}

// Exception handling
funct safeDivide(a, b) {
    try {
        if (b == 0) {
            return "Division by zero"
        }
        return a / b
    } catch(e) {
        println("Error:", e)
        return null
    }
}
```

### Module System

```hg
import "time" as t
import "fmaths" as math

let t0 = t.now()
let start = t.millis()

for range(i, 1000000) {}

let t1 = t.now()
println("Elapsed: " + (t1 - t0) + " seconds")
t.sleep(1.5)
```

### Math Library

```hg
import "fmaths" as math

// Mathematical constants
println("π = " + math.pi())
println("e = " + math.e())

// Basic functions
println("sqrt(16) = " + math.sqrt(16))     // 4
println("pow(2, 8) = " + math.pow(2, 8))  // 256
println("abs(-42) = " + math.abs(-42))    // 42

// Trigonometric functions
println("sin(π/2) = " + math.sin(math.pi() / 2))  // 1
println("cos(0) = " + math.cos(0))               // 1
println("tan(π/4) = " + math.tan(math.pi() / 4)) // 1

// Logarithmic and exponential
println("log(e) = " + math.log(math.e()))  // 1
println("log10(100) = " + math.log10(100)) // 2
println("log2(8) = " + math.log2(8))       // 3
println("exp(1) = " + math.exp(1))         // e

// Rounding and utility functions
println("ceil(3.2) = " + math.ceil(3.2))   // 4
println("floor(3.8) = " + math.floor(3.8)) // 3
println("round(3.6) = " + math.round(3.6)) // 4
println("min(5, 3) = " + math.min(5, 3))   // 3
println("max(5, 3) = " + math.max(5, 3))   // 5
```

---

## Architecture

### Execution Pipeline

1. **Lexer → Tokens**: [lexer/lexer.go](./lexer/lexer.go)
2. **Parser → AST**: [parser/parser.go](./parser/parser.go)
3. **Hybrid Runtime System**: [runtime/hybrid.go](./runtime/hybrid.go)
   - Smart routing between VM and interpreter based on code complexity
   - AST → Compiler → Bytecode → VM (optimized performance)
   - AST → Interpreter (flexible evaluation)
   - Performance tracking and adaptive execution

### Core Components

- **AST System**: Node definitions, pretty printing, function and import support
- **VM & Compiler**: High-performance stack-based VM with 20+ fast opcodes, peephole optimization, constant deduplication, call frame management
- **Runtime Environment**: Lexical scoping, dynamic value system, interpreter, pretty printing
- **Error System**: Line/column-aware parser and runtime errors
- **Module System**: Built-in libraries with aliasing support

---

## Project Structure

- **Entry Point**: [main.go](./main.go)
- **Core**: lexer, parser, AST
- **Runtime**: compiler, VM, interpreter, value system, environment, error handling, pretty printing
- **Libraries**: Built-in modules like `time`
- **Tests / Demos**: Comprehensive `.dy` scripts demonstrating all features

---

## Pretty Printing

- `pretty(v)` — inline, single-line formatting
- `prettyml(v)` — multi-line, indented, stable output with sorted keys
- `printlnml(v)` — prints `prettyml` output with newline

---

## Demos

All test files are organized in the `test/` directory for easy access:

- **Basic Features**: `go run . test/01_basic_features.dy`
- **Type System**: `go run . test/02_types_demo.dy`
- **Performance Tests**: `go run . test/03_performance_basic.dy`
- **VM Optimizations**: `go run . test/04_vm_optimization_benchmark.dy`
- **Bytecode Speed**: `go run . test/05_fast_bytecode_test.dy`
- **Time Module**: `go run . test/06_time_module.dy`
- **Time & Math Demo**: `go run . test/07_time_math_demo.dy`
- **Algorithm Patterns**: `go run . test/08_algorithm_patterns.dy`
- **Map Operations**: `go run . test/09_map_operations.dy`
- **Basic Math**: `go run . test/10_math_basic.dy`
- **Math Optimization**: `go run . test/11_math_optimization.dy`
- **Comprehensive Math Benchmark**: `go run . test/12_math_comprehensive_benchmark.dy`
- **New Language Features**: `go run . test/21_simple_test.dy`

---

## Roadmap

### Completed (v0.4)

- **High-performance bytecode VM**: 20+ specialized opcodes for common operations
- **Compiler optimizations**: Peephole optimization, constant deduplication, automatic optimization passes
- **Fast execution paths**: Optimized opcodes for constants (0, 1, true, false, null), loops, string operations
- **Memory optimizations**: Pre-allocated stacks, constant pooling, improved garbage collection
- User-defined functions with closures and return statements
- Enhanced time library with `now()`, `millis()`, `sleep()`
- Property access via dot notation for maps
- Advanced pretty printing with stable output
- Comprehensive error handling with line/column tracking

### Completed (v0.5)

- **Hybrid execution engine**: Smart VM/interpreter routing with performance tracking
- **Loop control flow**: `break` and `continue` statements with proper scoping
- **Increment/decrement operators**: Pre/post `++` and `--` with identifier support
- **Exception handling**: `try/catch` blocks with error variable binding
- **Performance optimizations**: Ultra-fast loops with sub-150ms execution
- **Memory improvements**: Object pooling and variable reuse patterns
- Array and map bracket indexing
- Expanded standard library with advanced `fmaths` module
- Modulo operator (`%`) support
- Enhanced identifier support (underscores allowed)
- **Function expressions**: Anonymous functions and lambda support
- **File extension change**: Now uses `.dy` and `.dx` extensions

### Future Enhancements

- `switch/case` statements
- File I/O functions
- Regular expressions
- Debugging tools
- User-defined modules
- Advanced VM optimizations

---

## Further Reading & Inspiration

- **Books**: Crafting Interpreters, Dragon Book, PLAI
- **Languages**: Lua, Wren, Go
- **Other**: SICP, Awesome Compilers

---

© 2025 DYMS. Licensed under the [MIT License](./LICENSE).
