package types

import (
	coretypes "github.com/raviqqe/lazy-ein/command/core/types"
	"github.com/raviqqe/lazy-ein/command/debug"
)

// Variable is a type variable used exclusively on type inference.
type Variable struct {
	identifier       int
	debugInformation *debug.Information
}

// NewVariable creates a variable.
func NewVariable(id int, i *debug.Information) Variable {
	return Variable{id, i}
}

// Identifier returns an identifier.
func (v Variable) Identifier() int {
	return v.identifier
}

// Unify unifies itself with another type.
func (v Variable) Unify(t Type) ([]Equation, error) {
	return []Equation{NewEquation(v, t)}, nil
}

// SubstituteVariable substitutes type variables.
func (v Variable) SubstituteVariable(vv Variable, t Type) Type {
	if v.identifier == vv.identifier {
		return t
	}

	return v
}

// DebugInformation returns debug information.
func (v Variable) DebugInformation() *debug.Information {
	return v.debugInformation
}

// ToCore returns a type in the core language.
func (v Variable) ToCore() coretypes.Type {
	panic("unreachable")
}

// VisitTypes visits types.
func (v Variable) VisitTypes(f func(Type) error) error {
	return f(v)
}
