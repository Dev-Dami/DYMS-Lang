package runtime

import "fmt"

// OpCode represents a VM instruction opcode.
type OpCode int

const (
	// Core opcodes
	OP_CONST OpCode = iota // push constant by index
	OP_LOAD_GLOBAL         // load global by name const index
	OP_STORE_GLOBAL        // store global by name const index
	OP_LOAD_LOCAL          // load local by slot
	OP_STORE_LOCAL         // store local by slot
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_CMP_EQ
	OP_CMP_NE
	OP_CMP_LT
	OP_CMP_LE
	OP_CMP_GT
	OP_CMP_GE
	OP_JUMP          // absolute ip
	OP_JUMP_IF_FALSE // absolute ip
	OP_CALL          // arg count
	OP_RET
	OP_POP
	OP_GET_PROP      // object on stack, prop name as const index
	OP_IMPORT        // alias const index, path const index
	
	// Fast opcodes for common patterns
	OP_INCREMENT_LOCAL    // increment local variable by 1 (slot)
	OP_DECREMENT_LOCAL    // decrement local variable by 1 (slot)
	OP_ADD_CONST         // add constant to top of stack (const_idx)
	OP_LOAD_CONST_0      // push constant 0 (no operands)
	OP_LOAD_CONST_1      // push constant 1 (no operands) 
	OP_LOAD_TRUE         // push true (no operands)
	OP_LOAD_FALSE        // push false (no operands)
	OP_LOAD_NULL         // push null (no operands)
	OP_DUP               // duplicate top of stack
	OP_SWAP              // swap top two stack values
	
	// String operations
	OP_CONCAT_2          // concat top 2 strings on stack
	OP_CONCAT_N          // concat N strings on stack (N as operand)
	
	// Array/Map operations  
	OP_MAKE_ARRAY        // create array from N stack values (N as operand)
	OP_MAKE_MAP          // create map from 2N stack values (N pairs, N as operand)
	OP_GET_INDEX         // array[index] - array and index on stack
	OP_SET_INDEX         // array[index] = value - array, index, value on stack
	
	// Loop optimization
	OP_FOR_LOOP_START    // optimized for loop initialization
	OP_FOR_LOOP_NEXT     // optimized for loop increment and check
	
	// Boolean operations
	OP_NOT               // logical not
	OP_AND               // logical and (short-circuit)
	OP_OR                // logical or (short-circuit)
)

// Chunk holds bytecode and a constant pool with optimizations.
type Chunk struct {
	Code      []int        // interleaved op and operands (ints for simplicity)
	Consts    []RuntimeVal // constants pool
	constMap  map[string]int // cache for constant deduplication
	lineInfo  []int        // line number information for debugging
}

// NewChunk creates a new optimized chunk
func NewChunk() *Chunk {
	return &Chunk{
		Code:     make([]int, 0, 256),
		Consts:   make([]RuntimeVal, 0, 64),
		constMap: make(map[string]int),
		lineInfo: make([]int, 0, 256),
	}
}

func (c *Chunk) emit(op OpCode, operands ...int) int {
	ip := len(c.Code)
	c.Code = append(c.Code, int(op))
	c.Code = append(c.Code, operands...)
	// Add line info placeholder (will be filled by compiler)
	c.lineInfo = append(c.lineInfo, 0)
	return ip
}

// Fast emit for common single-operand instructions
func (c *Chunk) emitFast(op OpCode, operand int) {
	c.Code = append(c.Code, int(op), operand)
	c.lineInfo = append(c.lineInfo, 0, 0)
}

// Add constant with deduplication for better memory usage
func (c *Chunk) addConst(v RuntimeVal) int {
	// Try to deduplicate simple constants
	key := c.getConstKey(v)
	if key != "" {
		if idx, exists := c.constMap[key]; exists {
			return idx
		}
	}
	
	idx := len(c.Consts)
	c.Consts = append(c.Consts, v)
	
	if key != "" {
		c.constMap[key] = idx
	}
	
	return idx
}

// Generate key for constant deduplication
func (c *Chunk) getConstKey(v RuntimeVal) string {
	switch val := v.(type) {
	case *NumberVal:
		if val.Value == 0 { return "num:0" }
		if val.Value == 1 { return "num:1" }
		if val.Value == -1 { return "num:-1" }
		return fmt.Sprintf("num:%f", val.Value)
	case *StringVal:
		if len(val.Value) < 64 { // Only cache short strings
			return "str:" + val.Value
		}
	case *BooleanVal:
		if val.Value { return "bool:true" }
		return "bool:false"
	case *NullVal:
		return "null"
	}
	return "" // Don't cache complex types
}
