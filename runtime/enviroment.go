package runtime

import "fmt"



type Environment struct {
	parent    *Environment
	variables map[string]RuntimeVal
	constants map[string]bool
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]RuntimeVal),
		constants: make(map[string]bool),
	}
}

func (env *Environment) DeclareVar(name string, value RuntimeVal, isConstant bool) RuntimeVal {
	if _, exists := env.variables[name]; exists {
		panic(fmt.Sprintf("Cannot declare variable '%s'. It already exists.", name))
	}
	env.variables[name] = value
	if isConstant {
		env.constants[name] = true
	}
	return value
}

func (env *Environment) AssignVar(name string, value RuntimeVal) RuntimeVal {
	if env.constants[name] {
		panic(fmt.Sprintf("Cannot assign to constant variable '%s'.", name))
	}
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
