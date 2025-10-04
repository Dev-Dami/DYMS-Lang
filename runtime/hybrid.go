package runtime

import (
	"DYMS/ast"
)

// HybridEngine combines VM and interpreter for optimal performance
type HybridEngine struct {
	vm              *VM
	interpreter     *Environment
	compiler        *Compiler
	vmCallCount     int
	interpreterCallCount int
	performanceMode bool
}

// NewHybridEngine creates a new hybrid execution system
func NewHybridEngine(globalEnv *Environment) *HybridEngine {
	return &HybridEngine{
		vm:                   NewVM(globalEnv),
		interpreter:          globalEnv,
		compiler:             NewCompiler(),
		vmCallCount:          0,
		interpreterCallCount: 0,
		performanceMode:      true, // Prefer VM for performance
	}
}

// SetPerformanceMode toggles between performance and compatibility
func (h *HybridEngine) SetPerformanceMode(enabled bool) {
	h.performanceMode = enabled
}

// GetStats returns execution statistics
func (h *HybridEngine) GetStats() (vmCalls, interpreterCalls int) {
	return h.vmCallCount, h.interpreterCallCount
}

// Execute decides whether to use VM or interpreter based on the AST node
func (h *HybridEngine) Execute(node ast.Stmt) (RuntimeVal, *Error) {
	switch n := node.(type) {
	case *ast.Program:
		return h.executeProgram(n)
	case *ast.FunctionDeclaration:
		return h.executeFunctionDeclaration(n)
	case *ast.CallExpr:
		return h.executeCallExpr(n)
	case *ast.BinaryExpr:
		return h.executeBinaryExpr(n)
	case *ast.ForStatement, *ast.WhileStatement:
		return h.executeLoop(n)
	default:
		// Fallback to interpreter for complex constructs
		return Evaluate(node, h.interpreter)
	}
}

// Execute program with hybrid approach
func (h *HybridEngine) executeProgram(program *ast.Program) (RuntimeVal, *Error) {
	var lastResult RuntimeVal
	var err *Error
	
	for _, stmt := range program.Body {
		lastResult, err = h.Execute(stmt)
		if err != nil {
			return nil, err
		}
		if _, isRet := lastResult.(*ReturnVal); isRet {
			return lastResult.(*ReturnVal).Inner, nil
		}
	}
	return lastResult, nil
}

// Functions use interpreter for reliability, VM for pure math functions
func (h *HybridEngine) executeFunctionDeclaration(fd *ast.FunctionDeclaration) (RuntimeVal, *Error) {
	// For now, use interpreter for all functions (reliable and fast enough)
	uf := &UserFunction{Params: fd.Params, Body: fd.Body, Env: h.interpreter}
	h.interpreter.DeclareVar(fd.Name, uf, true)
	return uf, nil
}

// Function calls use interpreter for reliability
func (h *HybridEngine) executeCallExpr(call *ast.CallExpr) (RuntimeVal, *Error) {
	// Use interpreter for all function calls
	return Evaluate(call, h.interpreter)
}

// Binary expressions use fast interpreter evaluation
func (h *HybridEngine) executeBinaryExpr(expr *ast.BinaryExpr) (RuntimeVal, *Error) {
	// Use optimized interpreter for all binary expressions
	return evalBinaryExpr(expr, h.interpreter)
}

// Loops use optimized interpreter for reliability
func (h *HybridEngine) executeLoop(stmt ast.Stmt) (RuntimeVal, *Error) {
	switch loop := stmt.(type) {
	case *ast.ForStatement:
		return evalForStatement(loop, h.interpreter)
	case *ast.WhileStatement:
		return evalWhileStatement(loop, h.interpreter)
	}
	return nil, NewError("unknown loop type", 0, 0)
}


