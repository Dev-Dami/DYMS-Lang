package runtime

import (
	"fmt"
	"math"
)

type frame struct {
	fn   *VMFunction
	ip   int
	base int // base -> stack index for my locals
}

type VM struct {
	stack   []RuntimeVal
	sp      int
	frames  []frame
	globals *Environment
}

// newvm -> pre-allocated stack
func NewVM(globals *Environment) *VM {
	return &VM{
		stack:   make([]RuntimeVal, 1024),
		sp:      0,
		frames:  make([]frame, 0, 64),
		globals: globals,
	}
}

// stack ops ->
func (vm *VM) push(v RuntimeVal) {
	if vm.sp >= len(vm.stack) {
		// push -> grow stack if needed
		newStack := make([]RuntimeVal, len(vm.stack)*2)
		copy(newStack, vm.stack)
		vm.stack = newStack
	}
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() RuntimeVal {
	vm.sp--
	v := vm.stack[vm.sp]
	vm.stack[vm.sp] = nil // pop -> clear to help gc
	return v
}

func (vm *VM) peek() RuntimeVal {
	return vm.stack[vm.sp-1]
}

// fast ops -> common constants
func (vm *VM) pushConst0() { vm.stack[vm.sp] = &NumberVal{Value: 0}; vm.sp++ }
func (vm *VM) pushConst1() { vm.stack[vm.sp] = &NumberVal{Value: 1}; vm.sp++ }
func (vm *VM) pushTrue()   { vm.stack[vm.sp] = &BooleanVal{Value: true}; vm.sp++ }
func (vm *VM) pushFalse()  { vm.stack[vm.sp] = &BooleanVal{Value: false}; vm.sp++ }
func (vm *VM) pushNull()   { vm.stack[vm.sp] = &NullVal{}; vm.sp++ }

// callfunction -> expect args on stack
func (vm *VM) callFunction(fn *VMFunction, argc int) {
	vm.frames = append(vm.frames, frame{fn: fn, ip: 0, base: vm.sp - argc})
}

func (vm *VM) Run(entry *VMFunction) (RuntimeVal, *Error) {
	vm.callFunction(entry, 0)

	for len(vm.frames) > 0 {
		fr := &vm.frames[len(vm.frames)-1]
		code := fr.fn.Chunk.Code
		consts := fr.fn.Chunk.Consts
		if fr.ip >= len(code) {
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				break
			}
			continue
		}
		op := OpCode(code[fr.ip])
		fr.ip++
		switch op {
		case OP_CONST:
			idx := code[fr.ip]
			fr.ip++
			vm.push(consts[idx])
		case OP_LOAD_GLOBAL:
			nameIdx := code[fr.ip]
			fr.ip++
			name := consts[nameIdx].(*StringVal).Value
			vm.push(vm.globals.LookupVar(name))
		case OP_STORE_GLOBAL:
			nameIdx := code[fr.ip]
			fr.ip++
			name := consts[nameIdx].(*StringVal).Value
			val := vm.pop()
			if _, ok := vm.globals.variables[name]; ok {
				vm.globals.variables[name] = val
			} else {
				vm.globals.DeclareVar(name, val, false)
			}
		case OP_LOAD_LOCAL:
			slot := code[fr.ip]
			fr.ip++
			vm.push(vm.stack[fr.base+int(slot)])
		case OP_STORE_LOCAL:
			slot := code[fr.ip]
			fr.ip++
			vm.stack[fr.base+int(slot)] = vm.peek()
		case OP_ADD, OP_SUB, OP_MUL, OP_DIV:
			r := vm.pop()
			l := vm.pop()
			ln, lok := l.(*NumberVal)
			rn, rok := r.(*NumberVal)
			if lok && rok {
				switch op {
				case OP_ADD:
					vm.push(&NumberVal{Value: ln.Value + rn.Value})
				case OP_SUB:
					vm.push(&NumberVal{Value: ln.Value - rn.Value})
				case OP_MUL:
					vm.push(&NumberVal{Value: ln.Value * rn.Value})
				case OP_DIV:
					if rn.Value == 0 {
						return nil, NewError("division by zero", 0, 0)
					}
					vm.push(&NumberVal{Value: ln.Value / rn.Value})
				}
				break
			}
			if op == OP_ADD {
				if ls, ok := l.(*StringVal); ok {
					vm.push(&StringVal{Value: ls.Value + r.String()})
					break
				}
				if rs, ok := r.(*StringVal); ok {
					vm.push(&StringVal{Value: l.String() + rs.Value})
					break
				}
			}
			return nil, NewError(fmt.Sprintf("unsupported operands for op: %s", op.String()), 0, 0)
		case OP_CMP_EQ, OP_CMP_NE, OP_CMP_LT, OP_CMP_LE, OP_CMP_GT, OP_CMP_GE:
			r := vm.pop()
			l := vm.pop()
			ln, lok := l.(*NumberVal)
			rn, rok := r.(*NumberVal)
			if lok && rok {
				switch op {
				case OP_CMP_EQ:
					vm.push(&BooleanVal{Value: ln.Value == rn.Value})
				case OP_CMP_NE:
					vm.push(&BooleanVal{Value: ln.Value != rn.Value})
				case OP_CMP_LT:
					vm.push(&BooleanVal{Value: ln.Value < rn.Value})
				case OP_CMP_LE:
					vm.push(&BooleanVal{Value: ln.Value <= rn.Value})
				case OP_CMP_GT:
					vm.push(&BooleanVal{Value: ln.Value > rn.Value})
				case OP_CMP_GE:
					vm.push(&BooleanVal{Value: ln.Value >= rn.Value})
				}
				break
			}
			if ls, ok := l.(*StringVal); ok {
				if rs, ok := r.(*StringVal); ok {
					if op == OP_CMP_EQ {
						vm.push(&BooleanVal{Value: ls.Value == rs.Value})
					} else if op == OP_CMP_NE {
						vm.push(&BooleanVal{Value: ls.Value != rs.Value})
					} else {
						return nil, NewError("unsupported string comparison", 0, 0)
					}
					break
				}
			}
			return nil, NewError("unsupported comparison", 0, 0)
		case OP_JUMP:
			fr.ip = int(code[fr.ip])
		case OP_JUMP_IF_FALSE:
			addr := int(code[fr.ip])
			fr.ip++
			cond := vm.pop()
			if b, ok := cond.(*BooleanVal); ok && !b.Value {
				fr.ip = addr
			}
		case OP_CALL:
			argc := int(code[fr.ip])
			fr.ip++
			callee := vm.stack[vm.sp-argc-1]
			switch f := callee.(type) {
			case Function:
				args := make([]RuntimeVal, argc)
				for i := argc - 1; i >= 0; i-- {
					args[i] = vm.pop()
				}
				vm.pop()
				res, err := f(args...)
				if err != nil {
					return nil, err
				}
				vm.push(res)
			case *VMFunction:
				vm.callFunction(f, argc)
			default:
				return nil, NewError("not a function", 0, 0)
			}
		case OP_RET:
			retVal := vm.pop()
			frame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return retVal, nil
			}
			vm.sp = frame.base - 1
			vm.stack = vm.stack[:vm.sp]
			vm.push(retVal)
		case OP_POP:
			_ = vm.pop()
		case OP_GET_PROP:
			nameIdx := code[fr.ip]
			fr.ip++
			name := consts[nameIdx].(*StringVal).Value
			obj := vm.pop()
			if m, ok := obj.(*MapVal); ok {
				if val, ok := m.Properties[name]; ok {
					vm.push(val)
					break
				}
				return nil, NewError("unknown property: "+name, 0, 0)
			}
			return nil, NewError(fmt.Sprintf("cannot get property '%s' on %s", name, obj.Type()), 0, 0)
		case OP_IMPORT:
			aliasIdx := code[fr.ip]
			pathIdx := code[fr.ip+1]
			fr.ip += 2
			alias := consts[aliasIdx].(*StringVal).Value
			path := consts[pathIdx].(*StringVal).Value
			mods := builtinModules()
			mod, ok := mods[path]
			if !ok {
				return nil, NewError("unknown module: "+path, 0, 0)
			}
			vm.globals.DeclareVar(alias, mod, true)

		// fast opcodes ->
		case OP_LOAD_CONST_0:
			vm.pushConst0()
		case OP_LOAD_CONST_1:
			vm.pushConst1()
		case OP_LOAD_TRUE:
			vm.pushTrue()
		case OP_LOAD_FALSE:
			vm.pushFalse()
		case OP_LOAD_NULL:
			vm.pushNull()

		// stack opcodes ->
		case OP_DUP:
			val := vm.peek()
			vm.push(val)
		case OP_SWAP:
			a := vm.pop()
			b := vm.pop()
			vm.push(a)
			vm.push(b)

		// local var opcodes ->
		case OP_INCREMENT_LOCAL:
			slot := code[fr.ip]
			fr.ip++
			if num, ok := vm.stack[fr.base+int(slot)].(*NumberVal); ok {
				vm.stack[fr.base+int(slot)] = &NumberVal{Value: num.Value + 1}
			} else {
				return nil, NewError("cannot increment non-number", 0, 0)
			}
		case OP_DECREMENT_LOCAL:
			slot := code[fr.ip]
			fr.ip++
			if num, ok := vm.stack[fr.base+int(slot)].(*NumberVal); ok {
				vm.stack[fr.base+int(slot)] = &NumberVal{Value: num.Value - 1}
			} else {
				return nil, NewError("cannot decrement non-number", 0, 0)
			}

		// string opcodes ->
		case OP_CONCAT_2:
			r := vm.pop()
			l := vm.pop()
			vm.push(&StringVal{Value: l.String() + r.String()})
		case OP_CONCAT_N:
			n := code[fr.ip]
			fr.ip++
			result := ""
			for i := 0; i < int(n); i++ {
				result = vm.pop().String() + result
			}
			vm.push(&StringVal{Value: result})

		// collection opcodes ->
		case OP_MAKE_ARRAY:
			n := code[fr.ip]
			fr.ip++
			elements := make([]RuntimeVal, n)
			for i := int(n) - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(&ArrayVal{Elements: elements})
		case OP_MAKE_MAP:
			n := code[fr.ip] // n -> key-value pairs
			fr.ip++
			props := make(map[string]RuntimeVal)
			for i := 0; i < int(n); i++ {
				value := vm.pop()
				key := vm.pop()
				if keyStr, ok := key.(*StringVal); ok {
					props[keyStr.Value] = value
				} else {
					return nil, NewError("map keys must be strings", 0, 0)
				}
			}
			vm.push(&MapVal{Properties: props})

		// for loop ->
		case OP_FOR_LOOP_NEXT:
			slot := code[fr.ip]
			fr.ip++
			// for_loop_next -> get counter & limit
			counter, ok1 := vm.stack[fr.base+int(slot)].(*NumberVal)
			limit, ok2 := vm.peek().(*NumberVal)
			if !ok1 || !ok2 {
				return nil, NewError("for loop requires numeric values", 0, 0)
			}
			// -> check counter < limit
			vm.push(&BooleanVal{Value: counter.Value < limit.Value})
			// -> increment counter
			vm.stack[fr.base+int(slot)] = &NumberVal{Value: counter.Value + 1}

		// math opcodes ->
		case OP_POW:
			y := vm.pop()
			x := vm.pop()
			xn, ok1 := x.(*NumberVal)
			yn, ok2 := y.(*NumberVal)
			if !ok1 || !ok2 {
				return nil, NewError("pow requires numeric arguments", 0, 0)
			}
			result := math.Pow(xn.Value, yn.Value)
			vm.push(&NumberVal{Value: result})

		case OP_SQRT:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("sqrt requires numeric argument", 0, 0)
			}
			if xn.Value < 0 {
				return nil, NewError("sqrt of negative number", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Sqrt(xn.Value)})

		case OP_SIN:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("sin requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Sin(xn.Value)})

		case OP_COS:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("cos requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Cos(xn.Value)})

		case OP_LOG:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("log requires numeric argument", 0, 0)
			}
			if xn.Value <= 0 {
				return nil, NewError("log of non-positive number", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Log(xn.Value)})

		case OP_EXP:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("exp requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Exp(xn.Value)})

		case OP_ABS:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("abs requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Abs(xn.Value)})

		case OP_FLOOR:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("floor requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Floor(xn.Value)})

		case OP_CEIL:
			x := vm.pop()
			xn, ok := x.(*NumberVal)
			if !ok {
				return nil, NewError("ceil requires numeric argument", 0, 0)
			}
			vm.push(&NumberVal{Value: math.Ceil(xn.Value)})

		default:
			return nil, NewError("unknown opcode", 0, 0)
		}
	}
	if vm.sp > 0 {
		return vm.pop(), nil
	}
	return &NullVal{}, nil
}

func (op OpCode) String() string {
	switch op {
	case OP_CONST:
		return "CONST"
	case OP_LOAD_GLOBAL:
		return "LOAD_GLOBAL"
	case OP_STORE_GLOBAL:
		return "STORE_GLOBAL"
	case OP_LOAD_LOCAL:
		return "LOAD_LOCAL"
	case OP_STORE_LOCAL:
		return "STORE_LOCAL"
	case OP_ADD:
		return "ADD"
	case OP_SUB:
		return "SUB"
	case OP_MUL:
		return "MUL"
	case OP_DIV:
		return "DIV"
	case OP_CMP_EQ:
		return "CMP_EQ"
	case OP_CMP_NE:
		return "CMP_NE"
	case OP_CMP_LT:
		return "CMP_LT"
	case OP_CMP_LE:
		return "CMP_LE"
	case OP_CMP_GT:
		return "CMP_GT"
	case OP_CMP_GE:
		return "CMP_GE"
	case OP_JUMP:
		return "JUMP"
	case OP_JUMP_IF_FALSE:
		return "JUMP_IF_FALSE"
	case OP_CALL:
		return "CALL"
	case OP_RET:
		return "RET"
	case OP_POP:
		return "POP"
	case OP_GET_PROP:
		return "GET_PROP"
	case OP_IMPORT:
		return "IMPORT"
	// fast opcodes ->
	case OP_LOAD_CONST_0:
		return "LOAD_CONST_0"
	case OP_LOAD_CONST_1:
		return "LOAD_CONST_1"
	case OP_LOAD_TRUE:
		return "LOAD_TRUE"
	case OP_LOAD_FALSE:
		return "LOAD_FALSE"
	case OP_LOAD_NULL:
		return "LOAD_NULL"
	case OP_DUP:
		return "DUP"
	case OP_SWAP:
		return "SWAP"
	case OP_INCREMENT_LOCAL:
		return "INCREMENT_LOCAL"
	case OP_DECREMENT_LOCAL:
		return "DECREMENT_LOCAL"
	case OP_CONCAT_2:
		return "CONCAT_2"
	case OP_CONCAT_N:
		return "CONCAT_N"
	case OP_MAKE_ARRAY:
		return "MAKE_ARRAY"
	case OP_MAKE_MAP:
		return "MAKE_MAP"
	case OP_FOR_LOOP_NEXT:
		return "FOR_LOOP_NEXT"
	// math opcodes ->
	case OP_POW:
		return "POW"
	case OP_SQRT:
		return "SQRT"
	case OP_SIN:
		return "SIN"
	case OP_COS:
		return "COS"
	case OP_LOG:
		return "LOG"
	case OP_EXP:
		return "EXP"
	case OP_ABS:
		return "ABS"
	case OP_FLOOR:
		return "FLOOR"
	case OP_CEIL:
		return "CEIL"
	default:
		return "?"
	}
}