package runtime

import (
	"DYMS/ast"
)

// HybridEngine combines VM and interpreter
type HybridEngine struct {
	vm                   *VM
	interpreter          *Environment
	compiler             *Compiler
	vmCallCount          int
	interpreterCallCount int
	performanceMode      bool
	functionStats        map[string]*FunctionStats
	loopComplexityThreshold int
}

// FunctionStats tracks performance metrics for functions
type FunctionStats struct {
	CallCount    int
	VMTime       int64 
	InterpreterTime int64
	PreferVM     bool
}

// NewHybridEngine creates a new hybrid execution system
func NewHybridEngine(globalEnv *Environment) *HybridEngine {
	return &HybridEngine{
		vm:                      NewVM(globalEnv),
		interpreter:             globalEnv,
		compiler:                NewCompiler(),
		vmCallCount:             0,
		interpreterCallCount:    0,
		performanceMode:         true,
		functionStats:           make(map[string]*FunctionStats),
		loopComplexityThreshold: 5, 
	}
}

func (h *HybridEngine) SetPerformanceMode(enabled bool) {
	h.performanceMode = enabled
}

func (h *HybridEngine) GetStats() (vmCalls, interpreterCalls int) {
	return h.vmCallCount, h.interpreterCallCount
}

func (h *HybridEngine) shouldUseVM(node ast.Stmt) bool {
	if !h.performanceMode {
		return false
	}
	
	// Simple expressions/math operations should prefer VM
	// for fast executions
	switch n := node.(type) {
	case *ast.BinaryExpr:
		return h.isMathExpression(n)
	case *ast.NumericLiteral, *ast.Identifier:
		return true
	case *ast.ForStatement:
		return h.isSimpleLoop(n)
	case *ast.FunctionDeclaration:
		return h.isPureFunction(n)
	default:
		return false
	}
}

func (h *HybridEngine) isMathExpression(expr *ast.BinaryExpr) bool {
	switch expr.Operator {
	case "+", "-", "*", "/", "%":
		return true
	case "==", "!=", "<", ">", "<=", ">=":
		return true 
	default:
		return false
	}
}

// isSimpleLoop determines if loop is simple enough for VM
func (h *HybridEngine) isSimpleLoop(stmt *ast.ForStatement) bool {
	if stmt.Body == nil {
		return true
	}
	statementCount := len(stmt.Body.Statements)
	return statementCount <= h.loopComplexityThreshold
}

func (h *HybridEngine) isPureFunction(fn *ast.FunctionDeclaration) bool {
	if fn.Body == nil {
		return false
	}
	// heuristic
	return len(fn.Body.Statements) <= 3
}

// Executes whether to use VM or interpreter based on heuristics
func (h *HybridEngine) Execute(node ast.Stmt) (RuntimeVal, *Error) {
	switch n := node.(type) {
	case *ast.Program:
		return h.executeProgram(n)
	case *ast.FunctionDeclaration:
		return h.executeFunctionDeclaration(n)
	case *ast.CallExpr:
		return h.executeCallExpr(n)
	case *ast.BinaryExpr:
		if h.shouldUseVM(n) {
			h.vmCallCount++
			return h.executeBinaryExprVM(n)
		}
		h.interpreterCallCount++
		return evalBinaryExpr(n, h.interpreter)
	case *ast.ForStatement, *ast.WhileStatement:
		return h.executeLoop(n)
	case *ast.BreakStatement, *ast.ContinueStatement:
		// control flow use interpreter
		h.interpreterCallCount++
		return Evaluate(node, h.interpreter)
	case *ast.TryStatement:
		// Exceptions handling use interpreter
		h.interpreterCallCount++
		return evalTryStatement(n, h.interpreter)
	default:
		// fallback to interpreter if complex constructs
		h.interpreterCallCount++
		return Evaluate(node, h.interpreter)
	}
}

// execution of program with a hybrid approach
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

// Functions to use interpreter for reliability and VM for pure math functions
func (h *HybridEngine) executeFunctionDeclaration(fd *ast.FunctionDeclaration) (RuntimeVal, *Error) {
	// For now, use interpreter for all functions for reliability
	// Would change to more hybrid solution
	uf := &UserFunction{Params: fd.Params, Body: fd.Body, Env: h.interpreter}
	h.interpreter.DeclareVar(fd.Name, uf, true)
	return uf, nil
}

// Function calls use interpreter for reliability
func (h *HybridEngine) executeCallExpr(call *ast.CallExpr) (RuntimeVal, *Error) {
	return Evaluate(call, h.interpreter)
}

// Binary expressions use interpreter evaluation
func (h *HybridEngine) executeBinaryExpr(expr *ast.BinaryExpr) (RuntimeVal, *Error) {
	// Use interpreter for all binary expressions
	return evalBinaryExpr(expr, h.interpreter)
}

// executeBinaryExprVM attempts to compile and run on VM
func (h *HybridEngine) executeBinaryExprVM(expr *ast.BinaryExpr) (RuntimeVal, *Error) {
	// For now, fallingback to interpreter bcuz VM compilation will be complex
	// For futute enhancement: i would compile simple math expressions to bytecode
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


