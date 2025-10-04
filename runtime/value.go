package runtime

import "fmt"

type ValueType string

const (
	NumberType  ValueType = "Number"
	StringType  ValueType = "String"
	BooleanType ValueType = "Boolean"
	ArrayType   ValueType = "Array"
	MapType     ValueType = "Map"
	NullType    ValueType = "Null"
	FunctionType ValueType = "Function"
	ReturnType   ValueType = "Return"
)

type RuntimeVal interface {
	Type() ValueType
	String() string
}

type NumberVal struct {
	Value float64
}

func (n *NumberVal) Type() ValueType { return NumberType }
func (n *NumberVal) String() string  { return fmt.Sprintf("%v", n.Value) }

type StringVal struct {
	Value string
}

func (s *StringVal) Type() ValueType { return StringType }
func (s *StringVal) String() string  { return s.Value }

type BooleanVal struct {
	Value bool
}

func (b *BooleanVal) Type() ValueType { return BooleanType }
func (b *BooleanVal) String() string  { return fmt.Sprintf("%v", b.Value) }


type ArrayVal struct {
	Elements []RuntimeVal
}

func (a *ArrayVal) Type() ValueType { return ArrayType }
func (a *ArrayVal) String() string  { return "[...Array]" }

type MapVal struct {
	Properties map[string]RuntimeVal
}

func (m *MapVal) Type() ValueType { return MapType }
func (m *MapVal) String() string  { return "{...Map}" }

type NullVal struct {
	Value interface{}
}

func (n *NullVal) Type() ValueType { return NullType }
func (n *NullVal) String() string  { return "null" }

// User-defined function value
type UserFunction struct {
	Params []string
	Body   interface{} // *ast.BlockStatement (kept untyped here to avoid import cycle)
	Env    *Environment
}

func (u *UserFunction) Type() ValueType { return FunctionType }
func (u *UserFunction) String() string  { return "[function]" }

// VM-compiled function value
type VMFunction struct {
	Name      string
	Arity     int
	Chunk     *Chunk
	LocalsMax int
}

func (v *VMFunction) Type() ValueType { return FunctionType }
func (v *VMFunction) String() string  { return "[function]" }

// Return value wrapper used to unwind up to call site
type ReturnVal struct {
	Inner RuntimeVal
}

func (r *ReturnVal) Type() ValueType { return ReturnType }
func (r *ReturnVal) String() string  { return r.Inner.String() }
