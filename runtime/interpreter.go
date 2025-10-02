package runtime

import (
	"fmt"
	"DYMS/ast"
	"log"
)

// Function represents a function in the language.
type Function func(args ...RuntimeVal) (RuntimeVal, *Error)

func (f Function) Type() ValueType { return "Function" }
func (f Function) String() string  { return "[function]" }

var GlobalEnv = NewEnvironment(nil)

func init() {
	log.SetFlags(0)
	GlobalEnv.DeclareVar("systemout", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		for _, arg := range args {
			log.Println(Pretty(arg))
		}
		return nil, nil
	}), true)
GlobalEnv.DeclareVar("println", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		for _, arg := range args {
			if s, ok := arg.(*StringVal); ok {
				fmt.Printf("[println]: %s\n", Unescape(s.Value))
			} else {
				fmt.Printf("[println]: %s\n", formatValue(arg))
			}
		}
		return nil, nil
	}), true)
	GlobalEnv.DeclareVar("printf", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) > 0 {
			format, ok := args[0].(*StringVal)
			if ok {
				var values []interface{}
				for _, arg := range args[1:] {
					if f, isFloat := arg.(*NumberVal); isFloat {
						values = append(values, int(f.Value))
					} else if s, isStr := arg.(*StringVal); isStr {
						// Pass raw string content to respect formatting and newlines
						values = append(values, s.Value)
					} else {
						values = append(values, Pretty(arg))
					}
				}
// Interpret basic escapes in the format string
				fmtStr := Unescape(format.Value)
				fmt.Printf(fmtStr, values...)
			} else {
				fmt.Println("First argument to printf must be a string")
			}
		}
		return nil, nil
	}), true)
	GlobalEnv.DeclareVar("logln", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		for _, arg := range args {
			log.Print("[logln]: ", formatValue(arg), " \n")
		}
		return nil, nil
	}), true)

	// pretty(value): returns single-line pretty string
	GlobalEnv.DeclareVar("pretty", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return &StringVal{Value: ""}, nil
		}
		return &StringVal{Value: Pretty(args[0])}, nil
	}), true)

	// prettyml(value): returns multi-line pretty string
	GlobalEnv.DeclareVar("prettyml", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return &StringVal{Value: ""}, nil
		}
		return &StringVal{Value: PrettyMultiline(args[0])}, nil
	}), true)

	// printlnml(value): print multi-line pretty string with trailing newline
	GlobalEnv.DeclareVar("printlnml", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			fmt.Println()
			return nil, nil
		}
		fmt.Println(PrettyMultiline(args[0]))
		return nil, nil
	}), true)
}

// Evaluator
func Evaluate(stmt ast.Stmt, scope *Environment) (RuntimeVal, *Error) {
	switch s := stmt.(type) {
	case *ast.NumericLiteral:
		return &NumberVal{Value: s.Value}, nil
	case *ast.StringLiteral:
		return &StringVal{Value: s.Value}, nil
	case *ast.BooleanLiteral:
		return &BooleanVal{Value: s.Value}, nil
	case *ast.ArrayLiteral:
		return evalArrayLiteral(s, scope)
	case *ast.MapLiteral:
		return evalMapLiteral(s, scope)
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
			return f(args...)
		} else {
			// No positional info at runtime for this node; surface a clean error
			return nil, NewError(fmt.Sprintf("not a function: %T", fn), 0, 0)
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
		return nil, NewError(fmt.Sprintf("unknown statement type: %T", s), 0, 0)
	}
}

func evalAssignmentExpr(node *ast.AssignmentExpr, scope *Environment) (RuntimeVal, *Error) {
	if ident, ok := node.Assignee.(*ast.Identifier); ok {
		value, err := Evaluate(node.Value, scope)
		if err != nil {
			return nil, err
		}
		return scope.AssignVar(ident.Symbol, value), nil
	}
	return nil, NewError(fmt.Sprintf("invalid assignment target: %T", node.Assignee), 0, 0)
}

func evalProgram(program *ast.Program, scope *Environment) (RuntimeVal, *Error) {
	var lastResult RuntimeVal
	var err *Error
	for _, stmt := range program.Body {
		lastResult, err = Evaluate(stmt, scope)
		if err != nil {
			return nil, err
		}
	}
	return lastResult, nil
}

func evalBlockStatement(block *ast.BlockStatement, scope *Environment) (RuntimeVal, *Error) {
	blockScope := NewEnvironment(scope)
	var lastResult RuntimeVal
	var err *Error
	for _, stmt := range block.Statements {
		lastResult, err = Evaluate(stmt, blockScope)
		if err != nil {
			return nil, err
		}
	}
	return lastResult, nil
}

func evalIfStatement(stmt *ast.IfStatement, scope *Environment) (RuntimeVal, *Error) {
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

func evalForStatement(stmt *ast.ForStatement, scope *Environment) (RuntimeVal, *Error) {
	rangeVal, err := Evaluate(stmt.Range, scope)
	if err != nil {
		return nil, err
	}

	if rng, ok := rangeVal.(*NumberVal); ok {
		forScope := NewEnvironment(scope)
		forScope.DeclareVar(stmt.Identifier.Symbol, &NumberVal{Value: 0}, false)
		for i := 0; i < int(rng.Value); i++ {
			forScope.AssignVar(stmt.Identifier.Symbol, &NumberVal{Value: float64(i)})
			_, err := evalBlockStatement(stmt.Body, forScope)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, NewError("for loop range must be a number", 0, 0)
	}

	return nil, nil
}

func evalWhileStatement(stmt *ast.WhileStatement, scope *Environment) (RuntimeVal, *Error) {
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

func evalArrayLiteral(lit *ast.ArrayLiteral, scope *Environment) (RuntimeVal, *Error) {
	elements := make([]RuntimeVal, len(lit.Elements))
	for i, el := range lit.Elements {
		evaled, err := Evaluate(el, scope)
		if err != nil {
			return nil, err
		}
		elements[i] = evaled
	}
	return &ArrayVal{Elements: elements}, nil
}

func evalMapLiteral(lit *ast.MapLiteral, scope *Environment) (RuntimeVal, *Error) {
	properties := make(map[string]RuntimeVal)
	for _, prop := range lit.Properties {
		key, err := Evaluate(prop.Key, scope)
		if err != nil {
			return nil, err
		}
		strKey, ok := key.(*StringVal)
		if !ok {
			return nil, NewError(fmt.Sprintf("map key must be a string, got %T", key), 0, 0)
		}

		value, err := Evaluate(prop.Value, scope)
		if err != nil {
			return nil, err
		}
		properties[strKey.Value] = value
	}
	return &MapVal{Properties: properties}, nil
}

func evalBinaryExpr(expr *ast.BinaryExpr, scope *Environment) (RuntimeVal, *Error) {
	leftVal, err := Evaluate(expr.Left, scope)
	if err != nil {
		return nil, err
	}
	rightVal, err := Evaluate(expr.Right, scope)
	if err != nil {
		return nil, err
	}

	// Handle numeric operations
	if leftNum, okLeft := leftVal.(*NumberVal); okLeft {
		if rightNum, okRight := rightVal.(*NumberVal); okRight {
			switch expr.Operator {
			case "+":
				return &NumberVal{Value: leftNum.Value + rightNum.Value}, nil
			case "-":
				return &NumberVal{Value: leftNum.Value - rightNum.Value}, nil
			case "*":
				return &NumberVal{Value: leftNum.Value * rightNum.Value}, nil
			case "/":
				if rightNum.Value == 0 {
					return nil, NewError("division by zero", 0, 0)
				}
				return &NumberVal{Value: leftNum.Value / rightNum.Value}, nil
			case "==":
				return &BooleanVal{Value: leftNum.Value == rightNum.Value}, nil
			case "!=":
				return &BooleanVal{Value: leftNum.Value != rightNum.Value}, nil
			case "<":
				return &BooleanVal{Value: leftNum.Value < rightNum.Value}, nil
			case "<=":
				return &BooleanVal{Value: leftNum.Value <= rightNum.Value}, nil
			case ">":
				return &BooleanVal{Value: leftNum.Value > rightNum.Value}, nil
			case ">=":
				return &BooleanVal{Value: leftNum.Value >= rightNum.Value}, nil
			}
		}
	}

	// Handle logical operations
	if leftBool, okLeft := leftVal.(*BooleanVal); okLeft {
		if rightBool, okRight := rightVal.(*BooleanVal); okRight {
			switch expr.Operator {
			case "&&":
				return &BooleanVal{Value: leftBool.Value && rightBool.Value}, nil
			case "||":
				return &BooleanVal{Value: leftBool.Value || rightBool.Value}, nil
			}
		}
	}

	// Handle string operations
	if leftStr, okLeft := leftVal.(*StringVal); okLeft {
		// If right is string and operator is + or comparison
		switch r := rightVal.(type) {
		case *StringVal:
			switch expr.Operator {
			case "+":
				return &StringVal{Value: leftStr.Value + r.Value}, nil
			case "==":
				return &BooleanVal{Value: leftStr.Value == r.Value}, nil
			case "!=":
				return &BooleanVal{Value: leftStr.Value != r.Value}, nil
			default:
				return nil, NewError(fmt.Sprintf("unknown operator %s for string operands", expr.Operator), 0, 0)
			}
		default:
			if expr.Operator == "+" {
				return &StringVal{Value: leftStr.Value + rightVal.String()}, nil
			}
		}
	}

	// Handle boolean equality
	if leftBool, okLeft := leftVal.(*BooleanVal); okLeft {
		if rightBool, okRight := rightVal.(*BooleanVal); okRight {
			switch expr.Operator {
			case "==":
				return &BooleanVal{Value: leftBool.Value == rightBool.Value}, nil
			case "!=":
				return &BooleanVal{Value: leftBool.Value != rightBool.Value}, nil
			}
		}
	}

	return nil, NewError(fmt.Sprintf("unknown operator %s for types %s and %s", expr.Operator, leftVal.Type(), rightVal.Type()), 0, 0)
}

func isTruthy(val RuntimeVal) bool {
	if val == nil {
		return false
	}
	if b, ok := val.(*BooleanVal); ok {
		return b.Value
	}
	if n, ok := val.(*NumberVal); ok {
		return n.Value != 0
	}
	if s, ok := val.(*StringVal); ok {
		return s.Value != ""
	}
	return true
}
