package libraries

import (
	"time"
	"DYMS/runtime"
)

func RegisterTime(env *runtime.Environment) {
    env.Set("now", runtime.BuiltinFunction(func(args ...runtime.Value) runtime.Value {
        return runtime.Number(float64(time.Now().UnixNano()) / 1e9) // seconds
    }))
    env.Set("sleep", runtime.BuiltinFunction(func(args ...runtime.Value) runtime.Value {
        if len(args) < 1 {
            return runtime.Null
        }
        sec, ok := args[0].(runtime.NumberValue)
        if ok {
            time.Sleep(time.Duration(sec.Float() * float64(time.Second)))
        }
        return runtime.Null
    }))
}
}