package eval

import "fmt"

// Binding holds a named value with const/forgiven metadata.
type Binding struct {
	Value    *Value
	IsConst  bool
	Forgiven bool
}

// Env is a lexical scope with an optional parent.
type Env struct {
	bindings map[string]*Binding
	parent   *Env
}

// NewEnv creates a new environment with an optional parent scope.
func NewEnv(parent *Env) *Env {
	return &Env{
		bindings: make(map[string]*Binding),
		parent:   parent,
	}
}

// Define creates a new binding in the current scope.
func (e *Env) Define(name string, val *Value, isConst bool) {
	e.bindings[name] = &Binding{Value: val, IsConst: isConst}
}

// Get looks up a binding by name, walking the scope chain.
func (e *Env) Get(name string) (*Value, error) {
	if b, ok := e.bindings[name]; ok {
		return b.Value, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("undefined variable: %s", name)
}

// Set updates an existing binding. Returns error if const and not forgiven.
func (e *Env) Set(name string, val *Value) error {
	if b, ok := e.bindings[name]; ok {
		if b.IsConst && !b.Forgiven {
			return fmt.Errorf("cannot reassign const: %s", name)
		}
		b.Value = val
		return nil
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return fmt.Errorf("undefined variable: %s", name)
}

// Forgive marks a const binding as forgiven so it can be reassigned.
// Only searches the current scope — sorry() must be called in the same scope as the const.
func (e *Env) Forgive(name string) error {
	if b, ok := e.bindings[name]; ok {
		b.Forgiven = true
		return nil
	}
	return fmt.Errorf("sorry: %s not found in current scope", name)
}
