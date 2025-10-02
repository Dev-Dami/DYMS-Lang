package runtime

import (
	"fmt"
	"DYMS/ast"
	"log"
)

// Function represents a function in the language.
type Function func(args ...RuntimeVal) RuntimeVal

var GlobalEnv = NewEnvironment(nil)

func init() {
	log.SetFlags(0)
	GlobalEnv.DeclareVar("systemout", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			log.Println(arg)
		}
		return nil
	}), true)
	GlobalEnv.DeclareVar("println", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			fmt.Printf("[println]: %s\n", formatValue(arg))
		}
		return nil
	}), true)
	GlobalEnv.DeclareVar("printf", Function(func(args ...RuntimeVal) RuntimeVal {
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
			} else {
				fmt.Println("First argument to printf must be a string")
			}
		}
		return nil
	}), true)
	GlobalEnv.DeclareVar("logln", Function(func(args ...RuntimeVal) RuntimeVal {
		for _, arg := range args {
			log.Print("[logln]: ", formatValue(arg), " \n")
		}
		return nil
	}), true)
}

func formatValue(v RuntimeVal) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		return val
	case Function:
		return "[function]"
	default:
		return fmt.Sprintf("%v", v)
	}
}
// Evaluator
func Evaluate(stmt ast.Stmt, scope *Environment) (RuntimeVal, error) {
	switch s := stmt.(type) {
	case *ast.NumericLiteral:
		return s.Value, nil
	case *ast.StringLiteral:
		return s.Value, nil
	case *ast.BinaryExpr:
		return evalBinaryExpr(s, scope)
	case *ast.Program:
		return evalProgram(s, scope)
	case *ast.VarDeclaration:
		value, err := Evaluate(s.Value, scope)
		if err != nil {
			return nil, err
		}
		return scope.DeclareVar(s.Identifier, value, s.Constant), nil
	case *ast.CallExpr:
		fn, err := Evaluate(s.Callee, scope)
		if err != nil {
			return nil, err
		}
		args := make([]RuntimeVal, len(s.Args))
		for i, arg := range s.Args {
			args[i], err = Evaluate(arg, scope)
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
		return scope.LookupVar(s.Symbol), nil
	case *ast.BlockStatement:
		return evalBlockStatement(s, scope)
	case *ast.IfStatement:
		return evalIfStatement(s, scope)
	case *ast.ForStatement:
		return evalForStatement(s, scope)
	case *ast.WhileStatement:
		return evalWhileStatement(s, scope)
	case *ast.AssignmentExpr:
		return evalAssignmentExpr(s, scope)
	default:
		return nil, fmt.Errorf("unknown statement type: %T", s)
	}
}

func evalAssignmentExpr(node *ast.AssignmentExpr, scope *Environment) (RuntimeVal, error) {
	if ident, ok := node.Assignee.(*ast.Identifier); ok {
		value, err := Evaluate(node.Value, scope)
		if err != nil {
			return nil, err
		}
		return scope.AssignVar(ident.Symbol, value), nil
	}
	return nil, fmt.Errorf("invalid assignment target: %T", node.Assignee)
}

func evalProgram(program *ast.Program, scope *Environment) (RuntimeVal, error) {
	var lastResult RuntimeVal
	var err error
	for _, stmt := range program.Body {
		lastResult, err = Evaluate(stmt, scope)
		if err != nil {
			return nil, err
		}
	}
	return lastResult, nil
}

func evalBlockStatement(block *ast.BlockStatement, scope *Environment) (RuntimeVal, error) {
	blockScope := NewEnvironment(scope)
	var lastResult RuntimeVal
	var err error
	for _, stmt := range block.Statements {
		lastResult, err = Evaluate(stmt, blockScope)
		if err != nil {
			return nil, err
		}
	}
	return lastResult, nil
}

func evalIfStatement(stmt *ast.IfStatement, scope *Environment) (RuntimeVal, error) {
	condition, err := Evaluate(stmt.Condition, scope)
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return evalBlockStatement(stmt.Consequence, scope)
	} else if stmt.Alternative != nil {
		return evalBlockStatement(stmt.Alternative, scope)
	}

	return nil, nil
}

func evalForStatement(stmt *ast.ForStatement, scope *Environment) (RuntimeVal, error) {
	rangeVal, err := Evaluate(stmt.Range, scope)
	if err != nil {
		return nil, err
	}

	if rng, ok := rangeVal.(float64); ok {
		forScope := NewEnvironment(scope)
		forScope.DeclareVar(stmt.Identifier.Symbol, float64(0), false)
		for i := 0; i < int(rng); i++ {
			forScope.AssignVar(stmt.Identifier.Symbol, float64(i))
			_, err := evalBlockStatement(stmt.Body, forScope)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, fmt.Errorf("for loop range must be a number")
	}

	return nil, nil
}

func evalWhileStatement(stmt *ast.WhileStatement, scope *Environment) (RuntimeVal, error) {
	for {
		condition, err := Evaluate(stmt.Condition, scope)
		if err != nil {
			return nil, err
		}

		if !isTruthy(condition) {
			break
		}

		_, err = evalBlockStatement(stmt.Body, scope)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func evalBinaryExpr(expr *ast.BinaryExpr, scope *Environment) (RuntimeVal, error) {
	leftVal, err := Evaluate(expr.Left, scope)
	if err != nil {
		return nil, err
	}
	rightVal, err := Evaluate(expr.Right, scope)
	if err != nil {
		return nil, err
	}

	// Handle numeric operations
	if leftFloat, okLeft := leftVal.(float64); okLeft {
		if rightFloat, okRight := rightVal.(float64); okRight {
			switch expr.Operator {
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
			case "==":
				return leftFloat == rightFloat, nil
			case "!=":
				return leftFloat != rightFloat, nil
			case "<":
				return leftFloat < rightFloat, nil
			case "<=":
				return leftFloat <= rightFloat, nil
			case ">":
				return leftFloat > rightFloat, nil
			case ">=":
				return leftFloat >= rightFloat, nil
			}
		}
	}

	// Handle logical operations
	if leftBool, okLeft := leftVal.(bool); okLeft {
		if rightBool, okRight := rightVal.(bool); okRight {
			switch expr.Operator {
			case "&&":
				return leftBool && rightBool, nil
			case "||":
				return leftBool || rightBool, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown operator %s for types %T and %T", expr.Operator, leftVal, rightVal)
}

func isTruthy(val RuntimeVal) bool {
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	if n, ok := val.(float64); ok {
		return n != 0
	}
	if s, ok := val.(string); ok {
		return s != ""
	}
	return true
}
