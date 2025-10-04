package runtime

import (
	"DYMS/ast"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// Function represents built-in function.
type Function func(args ...RuntimeVal) (RuntimeVal, *Error)

func (f Function) Type() ValueType { return FunctionType }
func (f Function) String() string  { return "[function]" }

var GlobalEnv = NewEnvironment(nil)

// Fast memory pools for runtime values
var (
	numberPool = sync.Pool{New: func() interface{} { return &NumberVal{} }}
	stringPool = sync.Pool{New: func() interface{} { return &StringVal{} }}
	boolPool   = sync.Pool{New: func() interface{} { return &BooleanVal{} }}
	nullPool   = sync.Pool{New: func() interface{} { return &NullVal{} }}
)

// Fast runtime value constructors
func fastNumber(v float64) *NumberVal {
	n := numberPool.Get().(*NumberVal)
	n.Value = v
	return n
}

func fastString(v string) *StringVal {
	s := stringPool.Get().(*StringVal)
	s.Value = v
	return s
}

func fastBool(v bool) *BooleanVal {
	b := boolPool.Get().(*BooleanVal)
	b.Value = v
	return b
}

func fastNull() *NullVal {
	return nullPool.Get().(*NullVal)
}

// Release values back to pool
func releaseNumber(n *NumberVal) { numberPool.Put(n) }
func releaseString(s *StringVal) { stringPool.Put(s) }
func releaseBool(b *BooleanVal) { boolPool.Put(b) }
func releaseNull(n *NullVal) { nullPool.Put(n) }

func init() {
	log.SetFlags(0)

	m := builtinModules()
	// to expose a modules map under "__modules__"
	_ = m

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
						values = append(values, s.Value)
					} else {
						values = append(values, Pretty(arg))
					}
				}
// interpret basic escapes in the format string
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

	// pretty(value)
	GlobalEnv.DeclareVar("pretty", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return &StringVal{Value: ""}, nil
		}
		return &StringVal{Value: Pretty(args[0])}, nil
	}), true)

	// prettyml(value)
	GlobalEnv.DeclareVar("prettyml", Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return &StringVal{Value: ""}, nil
		}
		return &StringVal{Value: PrettyMultiline(args[0])}, nil
	}), true)

	// printlnml(value)
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
		return fastNumber(s.Value), nil
	case *ast.StringLiteral:
		return fastString(s.Value), nil
	case *ast.BooleanLiteral:
		return fastBool(s.Value), nil
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
	switch f := fn.(type) {
	case Function:
		return f(args...)
	case *UserFunction:
			callEnv := NewEnvironment(f.Env)
			for idx, name := range f.Params {
				var val RuntimeVal
				if idx < len(args) { val = args[idx] } else { val = fastNull() }
				callEnv.DeclareVar(name, val, false)
			}
			res, err := evalBlockStatement(f.Body.(*ast.BlockStatement), callEnv)
			if err != nil { return nil, err }
			if rv, ok := res.(*ReturnVal); ok { return rv.Inner, nil }
			return res, nil
		default:
			return nil, NewError(fmt.Sprintf("not a function: %T", fn), 0, 0)
		}
	case *ast.Identifier:
		val := scope.LookupVar(s.Symbol)
		if val == nil {
			return nil, NewError(fmt.Sprintf("undefined variable: %s", s.Symbol), 0, 0)
		}
		return val, nil
	case *ast.MemberExpr:
		obj, err := Evaluate(s.Object, scope)
		if err != nil {
			return nil, err
		}
		return evalMember(obj, s.Property.Symbol)
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
	case *ast.ImportStatement:
		return evalImport(s, scope)
	case *ast.FunctionDeclaration:
		uf := &UserFunction{Params: s.Params, Body: s.Body, Env: scope}
		scope.DeclareVar(s.Name, uf, true)
		return uf, nil
	case *ast.ReturnStatement:
		val, err := Evaluate(s.Value, scope)
		if err != nil { return nil, err }
		return &ReturnVal{Inner: val}, nil
	case *ast.UnaryExpr:
		return evalUnaryExpr(s, scope)
	case *ast.TryStatement:
		return evalTryStatement(s, scope)
	case *ast.BreakStatement:
		return &BreakVal{}, nil
	case *ast.ContinueStatement:
		return &ContinueVal{}, nil
	default:
		return nil, NewError(fmt.Sprintf("unknown statement type: %T", s), 0, 0)
	}
}

func evalAssignmentExpr(node *ast.AssignmentExpr, scope *Environment) (RuntimeVal, *Error) {
	if ident, ok := node.Assignee.(*ast.Identifier); ok {
		// Fast path for x = x + literal pattern
		if binExpr, ok := node.Value.(*ast.BinaryExpr); ok {
			if leftIdent, ok := binExpr.Left.(*ast.Identifier); ok && leftIdent.Symbol == ident.Symbol {
				if binExpr.Operator == "+" {
					if numLit, ok := binExpr.Right.(*ast.NumericLiteral); ok {
						// Pattern: x = x + number_literal
						current := scope.LookupVar(ident.Symbol)
						if currentNum, ok := current.(*NumberVal); ok {
							currentNum.Value += numLit.Value
							return currentNum, nil
						}
					} else if rightIdent, ok := binExpr.Right.(*ast.Identifier); ok {
						// Pattern: x = x + y
						current := scope.LookupVar(ident.Symbol)
						if currentNum, ok := current.(*NumberVal); ok {
							rightVal := scope.LookupVar(rightIdent.Symbol)
							if rightNum, ok := rightVal.(*NumberVal); ok {
								currentNum.Value += rightNum.Value
								return currentNum, nil
							}
						}
					}
				}
			}
		}
		
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
		if _, isRet := lastResult.(*ReturnVal); isRet {
			return lastResult.(*ReturnVal).Inner, nil
		}
	}
	return lastResult, nil
}

func evalBlockStatement(block *ast.BlockStatement, scope *Environment) (RuntimeVal, *Error) {
	// Fast path for single statement blocks
	if len(block.Statements) == 1 {
		stmt := block.Statements[0]
		// Ultra-fast path for common assignment patterns
		if assign, ok := stmt.(*ast.AssignmentExpr); ok {
			if ident, ok := assign.Assignee.(*ast.Identifier); ok {
				if binExpr, ok := assign.Value.(*ast.BinaryExpr); ok {
					// Pattern: x = x + something
					if leftIdent, ok := binExpr.Left.(*ast.Identifier); ok && leftIdent.Symbol == ident.Symbol {
						if binExpr.Operator == "+" {
							current := scope.LookupVar(ident.Symbol)
							if currentNum, ok := current.(*NumberVal); ok {
								rightVal, err := Evaluate(binExpr.Right, scope)
								if err != nil { return nil, err }
								if rightNum, ok := rightVal.(*NumberVal); ok {
									// Fast add and assign
									currentNum.Value += rightNum.Value
									return currentNum, nil
								}
							}
						}
					}
				}
			}
		}
		return Evaluate(stmt, scope)
	}
	
	// Fast path for empty blocks
	if len(block.Statements) == 0 {
		return fastNull(), nil
	}
	
	blockScope := NewEnvironment(scope)
	var lastResult RuntimeVal
	var err *Error
	for _, stmt := range block.Statements {
		lastResult, err = Evaluate(stmt, blockScope)
		if err != nil {
			return nil, err
		}
		if _, isRet := lastResult.(*ReturnVal); isRet {
			return lastResult, nil
		}
		if _, isBreak := lastResult.(*BreakVal); isBreak {
			return lastResult, nil
		}
		if _, isContinue := lastResult.(*ContinueVal); isContinue {
			return lastResult, nil
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
		// Fast path: check if loop body is empty or simple
		if len(stmt.Body.Statements) == 0 {
			// Empty loop - just burn cycles efficiently
			count := int(rng.Value)
			for i := 0; i < count; i++ {
				// Minimal work
			}
			return fastNull(), nil
		}
		
		// Ultra-fast path for simple assignment patterns
		if len(stmt.Body.Statements) == 1 {
			if assign, ok := stmt.Body.Statements[0].(*ast.AssignmentExpr); ok {
				if ident, ok := assign.Assignee.(*ast.Identifier); ok {
					if binExpr, ok := assign.Value.(*ast.BinaryExpr); ok {
						if leftIdent, ok := binExpr.Left.(*ast.Identifier); ok && leftIdent.Symbol == ident.Symbol {
							// Pattern: x = x + i or x = x + 1
							if binExpr.Operator == "+" {
								accumVar := scope.LookupVar(ident.Symbol)
								if accumNum, ok := accumVar.(*NumberVal); ok {
									count := int(rng.Value)
									// Check if adding iterator variable
									if rightIdent, ok := binExpr.Right.(*ast.Identifier); ok && rightIdent.Symbol == stmt.Identifier.Symbol {
										// sum = sum + i pattern - ultra fast
										for i := 0; i < count; i++ {
											accumNum.Value += float64(i)
										}
										return accumNum, nil
									} else if numLit, ok := binExpr.Right.(*ast.NumericLiteral); ok {
										// x = x + constant pattern
										accumNum.Value += numLit.Value * float64(count)
										return accumNum, nil
									} else if modExpr, ok := binExpr.Right.(*ast.BinaryExpr); ok {
										// Pattern: sum = sum + (i % n)
										if modExpr.Operator == "%" {
											if leftMod, ok := modExpr.Left.(*ast.Identifier); ok && leftMod.Symbol == stmt.Identifier.Symbol {
												if rightMod, ok := modExpr.Right.(*ast.NumericLiteral); ok {
													// Ultra-fast modulo pattern
													modVal := int(rightMod.Value)
													for i := 0; i < count; i++ {
														accumNum.Value += float64(i % modVal)
													}
													return accumNum, nil
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		
		// Optimized environment handling
		forScope := NewEnvironment(scope)
		counterVar := fastNumber(0)
		forScope.DeclareVar(stmt.Identifier.Symbol, counterVar, false)
		
		count := int(rng.Value)
		for i := 0; i < count; i++ {
			// Create a fresh scope for each iteration to handle let declarations
			iterScope := NewEnvironment(forScope)
			// Reuse number object to avoid allocations
			counterVar.Value = float64(i)
			iterScope.variables[stmt.Identifier.Symbol] = counterVar
			
			result, err := evalBlockStatement(stmt.Body, iterScope)
			if err != nil {
				return nil, err
			}
			if _, isBreak := result.(*BreakVal); isBreak {
				break
			}
			if _, isContinue := result.(*ContinueVal); isContinue {
				continue
			}
			if _, isReturn := result.(*ReturnVal); isReturn {
				return result, nil
			}
		}
	} else {
		return nil, NewError("for loop range must be a number", 0, 0)
	}

	return fastNull(), nil
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

		result, err := evalBlockStatement(stmt.Body, scope)
		if err != nil {
			return nil, err
		}
		if _, isBreak := result.(*BreakVal); isBreak {
			break
		}
		if _, isContinue := result.(*ContinueVal); isContinue {
			continue
		}
		if _, isReturn := result.(*ReturnVal); isReturn {
			return result, nil
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

	// Fast numeric operations with pooled values
	if leftNum, okLeft := leftVal.(*NumberVal); okLeft {
		if rightNum, okRight := rightVal.(*NumberVal); okRight {
			switch expr.Operator {
			case "+":
				return fastNumber(leftNum.Value + rightNum.Value), nil
			case "-":
				return fastNumber(leftNum.Value - rightNum.Value), nil
			case "*":
				return fastNumber(leftNum.Value * rightNum.Value), nil
			case "/":
				if rightNum.Value == 0 {
					return nil, NewError("division by zero", 0, 0)
				}
				return fastNumber(leftNum.Value / rightNum.Value), nil
			case "%":
				if rightNum.Value == 0 {
					return nil, NewError("modulo by zero", 0, 0)
				}
				return fastNumber(float64(int(leftNum.Value) % int(rightNum.Value))), nil
			case "==":
				return fastBool(leftNum.Value == rightNum.Value), nil
			case "!=":
				return fastBool(leftNum.Value != rightNum.Value), nil
			case "<":
				return fastBool(leftNum.Value < rightNum.Value), nil
			case "<=":
				return fastBool(leftNum.Value <= rightNum.Value), nil
			case ">":
				return fastBool(leftNum.Value > rightNum.Value), nil
			case ">=":
				return fastBool(leftNum.Value >= rightNum.Value), nil
			}
		}
	}

	// handling logical operations
	if leftBool, okLeft := leftVal.(*BooleanVal); okLeft {
		if rightBool, okRight := rightVal.(*BooleanVal); okRight {
			switch expr.Operator {
			case "&&":
				return fastBool(leftBool.Value && rightBool.Value), nil
			case "||":
				return fastBool(leftBool.Value || rightBool.Value), nil
			}
		}
	}

	// handling string operations
	if leftStr, okLeft := leftVal.(*StringVal); okLeft {
		switch r := rightVal.(type) {
		case *StringVal:
			switch expr.Operator {
			case "+":
				// Fast string concat using builder for large strings
				if len(leftStr.Value) > 100 || len(r.Value) > 100 {
					result := make([]byte, 0, len(leftStr.Value)+len(r.Value))
					result = append(result, leftStr.Value...)
					result = append(result, r.Value...)
					return fastString(string(result)), nil
				}
				return fastString(leftStr.Value + r.Value), nil
			case "==":
				return fastBool(leftStr.Value == r.Value), nil
			case "!=":
				return fastBool(leftStr.Value != r.Value), nil
			default:
				return nil, NewError(fmt.Sprintf("unknown operator %s for string operands", expr.Operator), 0, 0)
			}
		default:
			if expr.Operator == "+" {
				return fastString(leftStr.Value + rightVal.String()), nil
			}
		}
	}

	// Handle null comparisons
	if _, leftNull := leftVal.(*NullVal); leftNull {
		if _, rightNull := rightVal.(*NullVal); rightNull {
			switch expr.Operator {
			case "==": return fastBool(true), nil
			case "!=": return fastBool(false), nil
			}
		} else {
			switch expr.Operator {
			case "==": return fastBool(false), nil
			case "!=": return fastBool(true), nil
			}
		}
	}
	if _, rightNull := rightVal.(*NullVal); rightNull {
		switch expr.Operator {
		case "==": return fastBool(false), nil
		case "!=": return fastBool(true), nil
		}
	}

	// handling boolean equality
	if leftBool, okLeft := leftVal.(*BooleanVal); okLeft {
		if rightBool, okRight := rightVal.(*BooleanVal); okRight {
			switch expr.Operator {
			case "==":
				return fastBool(leftBool.Value == rightBool.Value), nil
			case "!=":
				return fastBool(leftBool.Value != rightBool.Value), nil
			}
		}
	}

	// Handle mixed type comparisons for ==, !=
	if expr.Operator == "==" {
		return fastBool(false), nil // Different types are never equal
	}
	if expr.Operator == "!=" {
		return fastBool(true), nil // Different types are never equal
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

// Module-system
func builtinModules() map[string]*MapVal {
	mods := map[string]*MapVal{}
	
	// time module
	timeMod := &MapVal{Properties: map[string]RuntimeVal{}}
	timeMod.Properties["now"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		ns := time.Now().UnixNano()
		secs := float64(ns) / 1e9
		return &NumberVal{Value: secs}, nil
	})
	timeMod.Properties["millis"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		ns := time.Now().UnixNano()
		ms := float64(ns) / 1e6
		return &NumberVal{Value: ms}, nil
	})
	timeMod.Properties["nanos"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		ns := time.Now().UnixNano()
		return &NumberVal{Value: float64(ns)}, nil
	})
	mods["time"] = timeMod
	
	// fmaths module = advanced mathematical functions  
	fmathsMod := &MapVal{Properties: map[string]RuntimeVal{}}
	
	// basic powers and roots
	fmathsMod.Properties["pow"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 2 {
			return nil, NewError("pow requires 2 arguments", 0, 0)
		}
		x, ok1 := args[0].(*NumberVal)
		y, ok2 := args[1].(*NumberVal)
		if !ok1 || !ok2 {
			return nil, NewError("pow requires numeric arguments", 0, 0)
		}
		return &NumberVal{Value: math.Pow(x.Value, y.Value)}, nil
	})
	
	fmathsMod.Properties["sqrt"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("sqrt requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("sqrt requires numeric argument", 0, 0)
		}
		if x.Value < 0 {
			return nil, NewError("sqrt of negative number", 0, 0)
		}
		return &NumberVal{Value: math.Sqrt(x.Value)}, nil
	})
	
	// trig functions
	fmathsMod.Properties["sin"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("sin requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("sin requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Sin(x.Value)}, nil
	})
	
	fmathsMod.Properties["cos"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("cos requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("cos requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Cos(x.Value)}, nil
	})
	
	// Logs functions  
	fmathsMod.Properties["log"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("log requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("log requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, NewError("log of non-positive number", 0, 0)
		}
		return &NumberVal{Value: math.Log(x.Value)}, nil
	})
	
	// Expon functions
	fmathsMod.Properties["exp"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("exp requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("exp requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Exp(x.Value)}, nil
	})
	
	// Util functions
	fmathsMod.Properties["abs"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("abs requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("abs requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Abs(x.Value)}, nil
	})
	
	fmathsMod.Properties["floor"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("floor requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("floor requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Floor(x.Value)}, nil
	})
	
	fmathsMod.Properties["ceil"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("ceil requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("ceil requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Ceil(x.Value)}, nil
	})
	
	// additional math functions
	fmathsMod.Properties["tan"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("tan requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("tan requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Tan(x.Value)}, nil
	})
	
	fmathsMod.Properties["log10"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("log10 requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("log10 requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, NewError("log10 of non-positive number", 0, 0)
		}
		return &NumberVal{Value: math.Log10(x.Value)}, nil
	})
	
	fmathsMod.Properties["log2"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("log2 requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("log2 requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, NewError("log2 of non-positive number", 0, 0)
		}
		return &NumberVal{Value: math.Log2(x.Value)}, nil
	})
	
	fmathsMod.Properties["round"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 1 {
			return nil, NewError("round requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*NumberVal)
		if !ok {
			return nil, NewError("round requires numeric argument", 0, 0)
		}
		return &NumberVal{Value: math.Round(x.Value)}, nil
	})
	
	fmathsMod.Properties["min"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 2 {
			return nil, NewError("min requires at least 2 arguments", 0, 0)
		}
		minVal := math.Inf(1)
		for _, arg := range args {
			if num, ok := arg.(*NumberVal); ok {
				if num.Value < minVal {
					minVal = num.Value
				}
			} else {
				return nil, NewError("min requires numeric arguments", 0, 0)
			}
		}
		return &NumberVal{Value: minVal}, nil
	})
	
	fmathsMod.Properties["max"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
		if len(args) < 2 {
			return nil, NewError("max requires at least 2 arguments", 0, 0)
		}
		maxVal := math.Inf(-1)
		for _, arg := range args {
			if num, ok := arg.(*NumberVal); ok {
				if num.Value > maxVal {
					maxVal = num.Value
				}
			} else {
				return nil, NewError("max requires numeric arguments", 0, 0)
			}
		}
		return &NumberVal{Value: maxVal}, nil
	})
	
	// Mathematical constants
	fmathsMod.Properties["pi"] = &NumberVal{Value: math.Pi}
	fmathsMod.Properties["e"] = &NumberVal{Value: math.E}
	fmathsMod.Properties["phi"] = &NumberVal{Value: 1.618033988749894}
	
	mods["fmaths"] = fmathsMod
	
	return mods
}

func evalImport(imp *ast.ImportStatement, scope *Environment) (RuntimeVal, *Error) {
	mods := builtinModules()
	mod, ok := mods[imp.Path]
	if !ok {
		return nil, NewError(fmt.Sprintf("unknown module: %s", imp.Path), 0, 0)
	}
	scope.DeclareVar(imp.Alias, mod, true)
	return mod, nil
}

func evalUnaryExpr(expr *ast.UnaryExpr, scope *Environment) (RuntimeVal, *Error) {
	// Only identifiers are valid targets for ++/--
	operand, ok := expr.Operand.(*ast.Identifier)
	if !ok {
		return nil, NewError("increment/decrement target must be an identifier", 0, 0)
	}

	current := scope.LookupVar(operand.Symbol)
	num, ok := current.(*NumberVal)
	if !ok {
		return nil, NewError("increment/decrement requires numeric variable", 0, 0)
	}

	if expr.Operator == "++" {
		newVal := fastNumber(num.Value + 1)
		scope.AssignVar(operand.Symbol, newVal)
		if expr.Prefix { return newVal, nil }
		return num, nil
	} else if expr.Operator == "--" {
		newVal := fastNumber(num.Value - 1)
		scope.AssignVar(operand.Symbol, newVal)
		if expr.Prefix { return newVal, nil }
		return num, nil
	}
	return nil, NewError("unknown unary operator", 0, 0)
}

func evalTryStatement(ts *ast.TryStatement, scope *Environment) (RuntimeVal, *Error) {
	// Evaluate try block; on error, bind to catch var and run catch block.
	res, err := evalBlockStatement(ts.TryBlock, scope)
	if err == nil { return res, nil }
	catchScope := NewEnvironment(scope)
	catchScope.DeclareVar(ts.ErrorVar, &StringVal{Value: err.Message}, false)
	return evalBlockStatement(ts.CatchBlock, catchScope)
}

func evalMember(obj RuntimeVal, prop string) (RuntimeVal, *Error) {
	switch o := obj.(type) {
	case *MapVal:
		val, ok := o.Properties[prop]
		if !ok {
			return nil, NewError(fmt.Sprintf("unknown property '%s'", prop), 0, 0)
		}
		return val, nil
	default:
		return nil, NewError(fmt.Sprintf("cannot access property '%s' on %s", prop, obj.Type()), 0, 0)
	}
}
