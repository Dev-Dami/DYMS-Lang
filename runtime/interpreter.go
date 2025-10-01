package runtime

import (
	"fmt"
	"holygo/ast"
	"log"
)

// Function represents a function in the language.
type Function func(args ...RuntimeVal) RuntimeVal

var env = NewEnvironment(nil)

func init() {
	log.SetFlags(0)
	env.DeclareVar("systemout", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			log.Println(arg)
		}
		log.Println()
		return nil
	}), true)
	env.DeclareVar("println", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			fmt.Println(arg)
		}
		fmt.Println()
		return nil
	}), true)
	env.DeclareVar("printf", Function(func(args ...RuntimeVal) RuntimeVal {
		if len(args) > 0 {
			format, ok := args[0].(string)
			if ok {
				var values []interface{}
				for _, arg := range args[1:] {
					if f, isFloat := arg.(float64); isFloat {
						values = append(values, int(f))
					} else {
						values = append(values, arg)
					}
				}
				fmt.Printf(format, values...)
				fmt.Println()
			} else {
				fmt.Println("First argument to printf must be a string")
			}
		}
		return nil
	}), true)
	env.DeclareVar("logln", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			log.Println(arg)
		}
		log.Println()
		return nil
	}), true)
}

// Evaluator
func Evaluate(stmt ast.Stmt) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.NumericLiteral:
		return s.Value, nil
	case *ast.StringLiteral:
		return s.Value, nil
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
		return env.DeclareVar(s.Identifier, value, s.Constant), nil
	case *ast.CallExpr:
		fn, err := Evaluate(s.Callee)
		if err != nil {
			return nil, err
		}
		args := make([]RuntimeVal, len(s.Args))
		for i, arg := range s.Args {
			args[i], err = Evaluate(arg)
			if err != nil {
				return nil, err
			}
		}
		if f, ok := fn.(Function); ok {
			return f(args...), nil
		} else {
			return nil, fmt.Errorf("not a function: %T", fn)
		}
	case *ast.Identifier:
		return env.LookupVar(s.Symbol), nil
	default:
		return nil, fmt.Errorf("unknown statement type: %T", s)
	}
}
