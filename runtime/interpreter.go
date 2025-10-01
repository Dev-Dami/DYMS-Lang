package runtime

import (
	"fmt"
	"holygo/ast"
)

var env = NewEnvironment(nil)

// Evaluator
func Evaluate(stmt ast.Stmt) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.NumericLiteral:
		return s.Value, nil
	case *ast.Identifier:
		return env.LookupVar(s.Symbol), nil
	case *ast.BinaryExpr:
		leftVal, err := Evaluate(s.Left)
		if err != nil {
			return nil, err
		}
		rightVal, err := Evaluate(s.Right)
		if err != nil {
			return nil, err
		}

		leftFloat, okLeft := leftVal.(float64)
		rightFloat, okRight := rightVal.(float64)

		if !okLeft || !okRight {
			return nil, fmt.Errorf("binary operations can only be performed on numbers")
		}

		switch s.Operator {
		case "+":
			return leftFloat + rightFloat, nil
		case "-":
			return leftFloat - rightFloat, nil
		case "*":
			return leftFloat * rightFloat, nil
		case "/":
			if rightFloat == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return leftFloat / rightFloat, nil
		default:
			return nil, fmt.Errorf("unknown operator: %s", s.Operator)
		}
	case *ast.Program:
		var lastResult interface{}
		var err error
		for _, stmt := range s.Body {
			lastResult, err = Evaluate(stmt)
			if err != nil {
				return nil, err
			}
		}
		return lastResult, nil
	case *ast.VarDeclaration:
		value, err := Evaluate(s.Value)
		if err != nil {
			return nil, err
		}
		return env.DeclareVar(s.Identifier, value), nil
	default:
		return nil, fmt.Errorf("unknown statement type: %T", s)
	}
}
