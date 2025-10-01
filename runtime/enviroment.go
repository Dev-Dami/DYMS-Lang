package runtime

import "fmt"



type Environment struct {
	parent    *Environment
	variables map[string]RuntimeVal
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]RuntimeVal),
	}
}

func (env *Environment) DeclareVar(name string, value RuntimeVal) RuntimeVal {
	if _, exists := env.variables[name]; exists {
		panic(fmt.Sprintf("Cannot declare variable '%s'. It already exists.", name))
	}
	env.variables[name] = value
	return value
}

func (env *Environment) AssignVar(name string, value RuntimeVal) RuntimeVal {
	target := env.Resolve(name)
	target.variables[name] = value
	return value
}

func (env *Environment) LookupVar(name string) RuntimeVal {
	target := env.Resolve(name)
	return target.variables[name]
}

func (env *Environment) Resolve(name string) *Environment {
	if _, ok := env.variables[name]; ok {
		return env
	}
	if env.parent == nil {
		panic(fmt.Sprintf("Cannot resolve '%s'. Variable does not exist.", name))
	}
	return env.parent.Resolve(name)
}
