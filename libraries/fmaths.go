package libraries

import (
	"math"
)

// Define interfaces to avoid import cycles
type RuntimeVal interface {
	Type() string
	String() string
}

type NumberVal struct {
	Value float64
}

func (n *NumberVal) Type() string  { return "Number" }
func (n *NumberVal) String() string { return "" }

type MapVal struct {
	Properties map[string]RuntimeVal
}

func (m *MapVal) Type() string  { return "Map" }
func (m *MapVal) String() string { return "" }

type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Error() string { return e.Message }

func NewError(message string, line int, column int) *Error {
	return &Error{Message: message, Line: line, Column: column}
}

type Function func(args ...RuntimeVal) (RuntimeVal, *Error)

func (f Function) Type() string  { return "Function" }
func (f Function) String() string { return "[function]" }

// RegisterFMaths registers advanced mathematical functions
func RegisterFMaths() *MapVal {
	mathFuncs := make(map[string]RuntimeVal)
	
	// Basic powers and roots
	mathFuncs["pow"] = Function(func(args ...RuntimeVal) (RuntimeVal, *Error) {
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
	
	mathFuncs["sqrt"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("sqrt requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("sqrt requires numeric argument", 0, 0)
		}
		if x.Value < 0 {
			return nil, runtime.NewError("sqrt of negative number", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Sqrt(x.Value)}, nil
	})
	
	mathFuncs["cbrt"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("cbrt requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("cbrt requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Cbrt(x.Value)}, nil
	})
	
	// Logarithmic functions
	mathFuncs["log"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("log requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("log requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, runtime.NewError("log of non-positive number", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Log(x.Value)}, nil
	})
	
	mathFuncs["log10"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("log10 requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("log10 requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, runtime.NewError("log10 of non-positive number", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Log10(x.Value)}, nil
	})
	
	mathFuncs["log2"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("log2 requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("log2 requires numeric argument", 0, 0)
		}
		if x.Value <= 0 {
			return nil, runtime.NewError("log2 of non-positive number", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Log2(x.Value)}, nil
	})
	
	// Exponential functions
	mathFuncs["exp"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("exp requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("exp requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Exp(x.Value)}, nil
	})
	
	mathFuncs["exp2"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("exp2 requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("exp2 requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Exp2(x.Value)}, nil
	})
	
	// Trigonometric functions
	mathFuncs["sin"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("sin requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("sin requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Sin(x.Value)}, nil
	})
	
	mathFuncs["cos"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("cos requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("cos requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Cos(x.Value)}, nil
	})
	
	mathFuncs["tan"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("tan requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("tan requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Tan(x.Value)}, nil
	})
	
	// Inverse trigonometric functions
	mathFuncs["asin"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("asin requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("asin requires numeric argument", 0, 0)
		}
		if x.Value < -1 || x.Value > 1 {
			return nil, runtime.NewError("asin domain error: argument must be in [-1, 1]", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Asin(x.Value)}, nil
	})
	
	mathFuncs["acos"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("acos requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("acos requires numeric argument", 0, 0)
		}
		if x.Value < -1 || x.Value > 1 {
			return nil, runtime.NewError("acos domain error: argument must be in [-1, 1]", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Acos(x.Value)}, nil
	})
	
	mathFuncs["atan"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("atan requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("atan requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Atan(x.Value)}, nil
	})
	
	mathFuncs["atan2"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 2 {
			return nil, runtime.NewError("atan2 requires 2 arguments", 0, 0)
		}
		y, ok1 := args[0].(*runtime.NumberVal)
		x, ok2 := args[1].(*runtime.NumberVal)
		if !ok1 || !ok2 {
			return nil, runtime.NewError("atan2 requires numeric arguments", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Atan2(y.Value, x.Value)}, nil
	})
	
	// Hyperbolic functions
	mathFuncs["sinh"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("sinh requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("sinh requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Sinh(x.Value)}, nil
	})
	
	mathFuncs["cosh"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("cosh requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("cosh requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Cosh(x.Value)}, nil
	})
	
	mathFuncs["tanh"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("tanh requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("tanh requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Tanh(x.Value)}, nil
	})
	
	// Utility functions
	mathFuncs["abs"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("abs requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("abs requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Abs(x.Value)}, nil
	})
	
	mathFuncs["ceil"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("ceil requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("ceil requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Ceil(x.Value)}, nil
	})
	
	mathFuncs["floor"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("floor requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("floor requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Floor(x.Value)}, nil
	})
	
	mathFuncs["round"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("round requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("round requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Round(x.Value)}, nil
	})
	
	mathFuncs["min"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 2 {
			return nil, runtime.NewError("min requires at least 2 arguments", 0, 0)
		}
		minVal := math.Inf(1) // positive infinity
		for _, arg := range args {
			if num, ok := arg.(*runtime.NumberVal); ok {
				if num.Value < minVal {
					minVal = num.Value
				}
			} else {
				return nil, runtime.NewError("min requires numeric arguments", 0, 0)
			}
		}
		return &runtime.NumberVal{Value: minVal}, nil
	})
	
	mathFuncs["max"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 2 {
			return nil, runtime.NewError("max requires at least 2 arguments", 0, 0)
		}
		maxVal := math.Inf(-1) // negative infinity
		for _, arg := range args {
			if num, ok := arg.(*runtime.NumberVal); ok {
				if num.Value > maxVal {
					maxVal = num.Value
				}
			} else {
				return nil, runtime.NewError("max requires numeric arguments", 0, 0)
			}
		}
		return &runtime.NumberVal{Value: maxVal}, nil
	})
	
	// Mathematical constants
	mathFuncs["pi"] = &runtime.NumberVal{Value: math.Pi}
	mathFuncs["e"] = &runtime.NumberVal{Value: math.E}
	mathFuncs["phi"] = &runtime.NumberVal{Value: 1.618033988749894} // Golden ratio
	mathFuncs["sqrt2"] = &runtime.NumberVal{Value: math.Sqrt2}
	mathFuncs["ln2"] = &runtime.NumberVal{Value: math.Ln2}
	mathFuncs["ln10"] = &runtime.NumberVal{Value: math.Ln10}
	
	// Advanced functions
	mathFuncs["gamma"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("gamma requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("gamma requires numeric argument", 0, 0)
		}
		return &runtime.NumberVal{Value: math.Gamma(x.Value)}, nil
	})
	
	mathFuncs["factorial"] = runtime.Function(func(args ...runtime.RuntimeVal) (runtime.RuntimeVal, *runtime.Error) {
		if len(args) < 1 {
			return nil, runtime.NewError("factorial requires 1 argument", 0, 0)
		}
		x, ok := args[0].(*runtime.NumberVal)
		if !ok {
			return nil, runtime.NewError("factorial requires numeric argument", 0, 0)
		}
		n := int(x.Value)
		if n < 0 || float64(n) != x.Value {
			return nil, runtime.NewError("factorial requires non-negative integer", 0, 0)
		}
		result := 1.0
		for i := 2; i <= n; i++ {
			result *= float64(i)
		}
		return &runtime.NumberVal{Value: result}, nil
	})
	
	return &runtime.MapVal{Properties: mathFuncs}
}