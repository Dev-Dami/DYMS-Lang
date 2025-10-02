# DYMS — Dynamic Yet Minimal System Interpreter

**DYMS** is a lightweight, embeddable interpreter for a small dynamically-typed language, written in Go.
It is designed to be simple, extensible, and easy to integrate.

**Status:** Demo 0.3 • **License:** [MIT](./LICENSE) • **Requires:** Go ≥ 1.24

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

- **Variables**: `let`, `var`, `const`
- **Types**: Number, String, Boolean, Array, Map
- **Libraries**: Time
- **Operators**:
  - Arithmetic: `+`, `-`, `*`, `/`
  - Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
  - Logical: `&&`, `||`
  - String concatenation with `+`

- **Control flow**:
  - `if/else`
  - `while`
  - `for range(i, N)`

- **Comments**: Single-line with `//`
- **Built-ins**:
  - `println`, `systemout`, `logln`
  - `printf(format, ...)` (supports `\n`, `\t`, `\r\n`)
  - `pretty(v)` — single-line representation
  - `prettyml(v)` — multi-line representation
  - `printlnml(v)` — multi-line pretty print with line break

- **Error reporting**: Includes line and column numbers when available

---

## Quick Start

### Run a script

```powershell
go run . test.hg
```

### Build and run as binary

```powershell
go build -o hg.exe .
./hg.exe types_demo.hg
```

---

## Command-Line Usage

```text
hg <filename>
```

**Example**:

```powershell
go run . other.hg
```

---

## Language Overview

```text
let x = 10
var y = 20
const who = "DYMS"
let ok = true
let arr = [1, 2, 3]
let m = {"name": "DYMS", "stable": ok}

if (x > 5) { println("x > 5") } else { println("x <= 5") }

for range(i, 3) { println(i) }
```

```libraries
// time_demo.hg
// Demonstrates import aliasing and time.now()

import "time" as t

let t0 = t.now()
for range(i, 1000000) {
}
let t1 = t.now()

println("Elapsed sec = " + (t1 - t0))
```

---

## Architecture (High-Level)

- **Lexer → Tokens**: [lexer/lexer.go](./lexer/lexer.go)
- **Parser → AST**: [parser/parser.go](./parser/parser.go)
- **Runtime / Evaluator → Execution**:
  - Environments / scopes: [runtime/enviroment.go](./runtime/enviroment.go)
  - Values: [runtime/value.go](./runtime/value.go)
  - Interpreter & built-ins: [runtime/interpreter.go](./runtime/interpreter.go)
  - Pretty printers: [runtime/outputingpritier.go](./runtime/outputingpritier.go)

---

## Project Structure

- **Entry point**: [main.go](./main.go)
- **Core language**: [lexer/](./lexer), [parser/](./parser), [ast/ast.go](./ast/ast.go)
- **Runtime**: [runtime/](./runtime)
- **Examples**: [test.hg](./test.hg), [other.hg](./other.hg), [types_demo.hg](./types_demo.hg)

---

## Pretty Printing

- **`pretty(v)`** — inline, quoted strings; arrays and maps in single line
- **`prettyml(v)`** — multi-line, indented; maps with sorted keys for stability
- **`printlnml(v)`** — prints `prettyml` with trailing newline

---

## Demos

The recommended demo showcases types, control flow, and pretty printing:

```powershell
go run . types_demo.hg
```

---

## Roadmap

- Array and map indexing / field access
- Small standard library (strings, collections)
- Improved parser diagnostics and error recovery

---

## Further Reading & Inspiration

DYMS draws inspiration from both academic and practical works on programming language design, interpreters, and compilers. If you want to explore similar topics, here are some recommended resources:

- **Books**:
  - [Crafting Interpreters](https://craftinginterpreters.com/) by Robert Nystrom — a modern and practical guide to writing interpreters.
  - [Compilers: Principles, Techniques, and Tools (Dragon Book)](https://en.wikipedia.org/wiki/Compilers:_Principles,_Techniques,_and_Tools) by Aho, Lam, Sethi, and Ullman.
  - [Programming Languages: Application and Interpretation (PLAI)](http://cs.brown.edu/~sk/Publications/Books/ProgLangs/) by Shriram Krishnamurthi.

- **Projects / Languages**:
  - [Lua](https://www.lua.org/) — a lightweight embeddable scripting language.
  - [Wren](https://wren.io/) — a small, fast language by Robert Nystrom, focused on simplicity.
  - [Go](https://go.dev/) itself — whose minimalism and clarity influenced DYMS.

- **Other Resources**:
  - [Structure and Interpretation of Computer Programs (SICP)](https://mitpress.mit.edu/9780262510875/structure-and-interpretation-of-computer-programs/) — a foundational text on programming languages and abstractions.
  - [Awesome Compilers](https://github.com/aalhour/awesome-compilers) — a curated list of compilers, interpreters, and related resources.

---

© 2025 DYMS. Licensed under the [MIT License](./LICENSE).
