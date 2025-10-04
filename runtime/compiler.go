package runtime

import (
	"DYMS/ast"
)

type functionScope struct {
	locals     map[string]int // name -> slot
	localsMax  int
	isTopLevel bool
}

type Compiler struct {
	chunk   *Chunk
	scopes  []*functionScope
}

func NewCompiler() *Compiler {
	c := &Compiler{chunk: NewChunk()}
	c.pushScope(true)
	return c
}

func (c *Compiler) pushScope(isTop bool) {
	s := &functionScope{locals: map[string]int{}, isTopLevel: isTop}
	c.scopes = append(c.scopes, s)
}

func (c *Compiler) popScope() { c.scopes = c.scopes[:len(c.scopes)-1] }

func (c *Compiler) scope() *functionScope { return c.scopes[len(c.scopes)-1] }

func (c *Compiler) Compile(prog *ast.Program) *VMFunction {
	// Compile program body into a top-level function
	for _, stmt := range prog.Body {
		c.compileStmt(stmt)
	}
	// implicit return
	c.chunk.emit(OP_LOAD_NULL)
	c.chunk.emit(OP_RET)
	
	// Run optimization pass for better performance
	c.optimize()
	
	return &VMFunction{Name: "<main>", Arity: 0, Chunk: c.chunk, LocalsMax: c.scope().localsMax}
}

func (c *Compiler) compileStmt(s ast.Stmt) {
	switch n := s.(type) {
	case *ast.VarDeclaration:
		c.compileExpr(n.Value)
		if c.scope().isTopLevel {
			nameIdx := c.chunk.addConst(&StringVal{Value: n.Identifier})
			c.chunk.emit(OP_STORE_GLOBAL, nameIdx)
		} else {
			slot := c.ensureLocal(n.Identifier)
			c.chunk.emit(OP_STORE_LOCAL, slot)
		}
	case *ast.AssignmentExpr:
		// Only identifier targets supported here
		if ident, ok := n.Assignee.(*ast.Identifier); ok {
			c.compileExpr(n.Value)
			if slot, ok := c.scope().locals[ident.Symbol]; ok {
				c.chunk.emit(OP_STORE_LOCAL, slot)
			} else {
				nameIdx := c.chunk.addConst(&StringVal{Value: ident.Symbol})
				c.chunk.emit(OP_STORE_GLOBAL, nameIdx)
			}
		}
	case *ast.IfStatement:
		c.compileExpr(n.Condition)
		jfalse := c.chunk.emit(OP_JUMP_IF_FALSE, -1)
		c.compileBlock(n.Consequence)
		jend := c.chunk.emit(OP_JUMP, -1)
		c.patch(jfalse, len(c.chunk.Code))
		if n.Alternative != nil {
			c.compileBlock(n.Alternative)
		}
		c.patch(jend, len(c.chunk.Code))
	case *ast.WhileStatement:
		start := len(c.chunk.Code)
		c.compileExpr(n.Condition)
		jfalse := c.chunk.emit(OP_JUMP_IF_FALSE, -1)
		c.compileBlock(n.Body)
		c.chunk.emit(OP_JUMP, start)
		c.patch(jfalse, len(c.chunk.Code))
	case *ast.ForStatement:
		// Optimized for range(i, N) compilation
		slot := c.ensureLocal(n.Identifier.Symbol)
		
		// Initialize loop counter to 0 (fast opcode)
		c.chunk.emit(OP_LOAD_CONST_0)
		c.chunk.emit(OP_STORE_LOCAL, slot)
		
		// Compile range expression once
		c.compileExpr(n.Range)
		loopStart := len(c.chunk.Code)
		
		// Optimized loop condition and increment
		c.chunk.emit(OP_FOR_LOOP_NEXT, slot) // handles condition check and increment
		jfalse := c.chunk.emit(OP_JUMP_IF_FALSE, -1)
		
		// Compile body
		c.compileBlock(n.Body)
		
		// Jump back to start
		c.chunk.emit(OP_JUMP, loopStart)
		c.patch(jfalse, len(c.chunk.Code))
		
		// Clean up range value from stack
		c.chunk.emit(OP_POP)
	case *ast.FunctionDeclaration:
		fn := c.compileFunction(n)
		idx := c.chunk.addConst(fn)
		// bind to name (store global at top-level or local inside function)
		if c.scope().isTopLevel {
			nameIdx := c.chunk.addConst(&StringVal{Value: n.Name})
			c.chunk.emit(OP_CONST, idx)
			c.chunk.emit(OP_STORE_GLOBAL, nameIdx)
		} else {
			slot := c.ensureLocal(n.Name)
			c.chunk.emit(OP_CONST, idx)
			c.chunk.emit(OP_STORE_LOCAL, slot)
		}
	case *ast.ReturnStatement:
		c.compileExpr(n.Value)
		c.chunk.emit(OP_RET)
	case *ast.ImportStatement:
		aliasIdx := c.chunk.addConst(&StringVal{Value: n.Alias})
		pathIdx := c.chunk.addConst(&StringVal{Value: n.Path})
		c.chunk.emit(OP_IMPORT, aliasIdx, pathIdx)
	default:
		// expression statement
		c.compileExpr(n.(ast.Expr))
		c.chunk.emit(OP_POP)
	}
}

func (c *Compiler) compileBlock(b *ast.BlockStatement) {
	for _, stmt := range b.Statements {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) compileFunction(fd *ast.FunctionDeclaration) *VMFunction {
	// New compiler for function body with optimized chunk
	inner := &Compiler{chunk: NewChunk()}
	inner.pushScope(false)
	// Reserve locals for params
	for _, p := range fd.Params {
		slot := inner.ensureLocal(p)
		_ = slot
	}
	inner.compileBlock(fd.Body)
	// Ensure function returns null if no explicit return
	inner.chunk.emit(OP_LOAD_NULL)
	inner.chunk.emit(OP_RET)
	return &VMFunction{Name: fd.Name, Arity: len(fd.Params), Chunk: inner.chunk, LocalsMax: inner.scope().localsMax}
}

func (c *Compiler) compileExpr(e ast.Expr) {
	switch n := e.(type) {
	case *ast.NumericLiteral:
		// Use fast opcodes for common constants
		if n.Value == 0 {
			c.chunk.emit(OP_LOAD_CONST_0)
		} else if n.Value == 1 {
			c.chunk.emit(OP_LOAD_CONST_1)
		} else {
			c.chunk.emit(OP_CONST, c.chunk.addConst(&NumberVal{Value: n.Value}))
		}
	case *ast.StringLiteral:
		c.chunk.emit(OP_CONST, c.chunk.addConst(&StringVal{Value: n.Value}))
	case *ast.BooleanLiteral:
		// Use fast opcodes for booleans
		if n.Value {
			c.chunk.emit(OP_LOAD_TRUE)
		} else {
			c.chunk.emit(OP_LOAD_FALSE)
		}
	case *ast.Identifier:
		if slot, ok := c.scope().locals[n.Symbol]; ok {
			c.chunk.emit(OP_LOAD_LOCAL, slot)
		} else {
			nameIdx := c.chunk.addConst(&StringVal{Value: n.Symbol})
			c.chunk.emit(OP_LOAD_GLOBAL, nameIdx)
		}
	case *ast.BinaryExpr:
		c.compileExpr(n.Left)
		c.compileExpr(n.Right)
		switch n.Operator {
		case "+": c.chunk.emit(OP_ADD)
		case "-": c.chunk.emit(OP_SUB)
		case "*": c.chunk.emit(OP_MUL)
		case "/": c.chunk.emit(OP_DIV)
		case "==": c.chunk.emit(OP_CMP_EQ)
		case "!=": c.chunk.emit(OP_CMP_NE)
		case "<":  c.chunk.emit(OP_CMP_LT)
		case "<=": c.chunk.emit(OP_CMP_LE)
		case ">":  c.chunk.emit(OP_CMP_GT)
		case ">=": c.chunk.emit(OP_CMP_GE)
		}
	case *ast.CallExpr:
		// Check for optimizable math function calls
		if c.tryOptimizeMathCall(n) {
			return
		}
		// Default function call
		c.compileExpr(n.Callee)
		for _, a := range n.Args { c.compileExpr(a) }
		c.chunk.emit(OP_CALL, len(n.Args))
	case *ast.MemberExpr:
		c.compileExpr(n.Object)
		nameIdx := c.chunk.addConst(&StringVal{Value: n.Property.Symbol})
		c.chunk.emit(OP_GET_PROP, nameIdx)
	case *ast.ArrayLiteral:
		// Not compiled in minimal VM; leave to interpreter if needed later
		// As a placeholder, push null
		c.chunk.emit(OP_CONST, c.chunk.addConst(&NullVal{}))
	case *ast.MapLiteral:
		c.chunk.emit(OP_CONST, c.chunk.addConst(&NullVal{}))
	default:
		c.chunk.emit(OP_CONST, c.chunk.addConst(&NullVal{}))
	}
}

func (c *Compiler) ensureLocal(name string) int {
	if slot, ok := c.scope().locals[name]; ok { return slot }
	s := c.scope()
	slot := s.localsMax
	s.locals[name] = slot
	s.localsMax++
	return slot
}

func (c *Compiler) patch(jumpPos int, target int) {
	// jumpPos points to the opcode; operand is at jumpPos+1
	c.chunk.Code[jumpPos+1] = target
}

// Try to optimize math function calls to fast opcodes
func (c *Compiler) tryOptimizeMathCall(call *ast.CallExpr) bool {
	// Check if this is a member expression like math.pow, math.sqrt, etc.
	if memberExpr, ok := call.Callee.(*ast.MemberExpr); ok {
		if ident, ok := memberExpr.Object.(*ast.Identifier); ok {
			// Check if calling functions on a math module
			if ident.Symbol == "math" || ident.Symbol == "m" { // common aliases
				switch memberExpr.Property.Symbol {
				case "pow":
					if len(call.Args) == 2 {
						c.compileExpr(call.Args[0]) // base
						c.compileExpr(call.Args[1]) // exponent
						c.chunk.emit(OP_POW)
						return true
					}
				case "sqrt":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_SQRT)
						return true
					}
				case "sin":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_SIN)
						return true
					}
				case "cos":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_COS)
						return true
					}
				case "log":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_LOG)
						return true
					}
				case "exp":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_EXP)
						return true
					}
				case "abs":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_ABS)
						return true
					}
				case "floor":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_FLOOR)
						return true
					}
				case "ceil":
					if len(call.Args) == 1 {
						c.compileExpr(call.Args[0])
						c.chunk.emit(OP_CEIL)
						return true
					}
				}
			}
		}
	}
	return false
}

// Peephole optimization pass
func (c *Compiler) optimize() {
	code := c.chunk.Code
	for i := 0; i < len(code)-2; i++ {
		// Optimize: CONST 0, STORE_LOCAL -> LOAD_CONST_0, STORE_LOCAL
		if OpCode(code[i]) == OP_CONST && code[i+1] < len(c.chunk.Consts) {
			if num, ok := c.chunk.Consts[code[i+1]].(*NumberVal); ok && num.Value == 0 {
				code[i] = int(OP_LOAD_CONST_0)
				// Remove operand by shifting left
				copy(code[i+1:], code[i+2:])
				c.chunk.Code = code[:len(code)-1]
				continue
			}
			if num, ok := c.chunk.Consts[code[i+1]].(*NumberVal); ok && num.Value == 1 {
				code[i] = int(OP_LOAD_CONST_1)
				copy(code[i+1:], code[i+2:])
				c.chunk.Code = code[:len(code)-1]
				continue
			}
			if b, ok := c.chunk.Consts[code[i+1]].(*BooleanVal); ok {
				if b.Value {
					code[i] = int(OP_LOAD_TRUE)
				} else {
					code[i] = int(OP_LOAD_FALSE)
				}
				copy(code[i+1:], code[i+2:])
				c.chunk.Code = code[:len(code)-1]
				continue
			}
		}
		
		// Optimize: LOAD_LOCAL, CONST 1, ADD, STORE_LOCAL (same slot) -> INCREMENT_LOCAL
		if i+5 < len(code) && 
			OpCode(code[i]) == OP_LOAD_LOCAL &&
			OpCode(code[i+2]) == OP_CONST &&
			OpCode(code[i+4]) == OP_ADD &&
			OpCode(code[i+5]) == OP_STORE_LOCAL &&
			code[i+1] == code[i+6] { // same slot
			
			if constIdx := code[i+3]; constIdx < len(c.chunk.Consts) {
				if num, ok := c.chunk.Consts[constIdx].(*NumberVal); ok && num.Value == 1 {
					// Replace with INCREMENT_LOCAL
					code[i] = int(OP_INCREMENT_LOCAL)
					// Remove the 5 following instructions
					copy(code[i+2:], code[i+7:])
					c.chunk.Code = code[:len(code)-5]
				}
			}
		}
	}
}
