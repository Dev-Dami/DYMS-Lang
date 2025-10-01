package runtime

import (
	"fmt"
	"holygo/ast"
)

// Evaluates a binary expression using numeric operands
func EvalBinaryExpr(lhs float64, rhs float64, operator string) float64 {
	switch operator {
	case "+":
		return lhs + rhs
	case "-":
		return lhs - rhs
	case "*":
		return lhs * rhs
	case "/":
		if rhs == 0 {
			panic("division by zero")
		}
		return lhs / rhs
	case "%":
		return float64(int(lhs) % int(rhs))
	default:
		panic(fmt.Sprintf("unknown operator: %s", operator))
	}
}

// Evaluates an identifier using the environment
func EvalIdentifier(ident *ast.Identifier, env *Environment) interface{} {
	return env.LookupVar(ident.Symbol)
}

// Evaluates a list of statements as a "program"
func EvalProgram(program *ast.Program, env *Environment) interface{} {
	var last interface{}
	for _, stmt := range program.Body {
		last, _ = Evaluate(stmt)
	}
	return last
}


