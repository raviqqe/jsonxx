package ast

import "github.com/ein-lang/ein/command/core/types"

// FunctionApplication is a function application.
type FunctionApplication struct {
	function  Variable
	arguments []Atom
}

// NewFunctionApplication creates an application.
func NewFunctionApplication(v Variable, as []Atom) FunctionApplication {
	return FunctionApplication{v, as}
}

// Function returns a function.
func (a FunctionApplication) Function() Variable {
	return a.function
}

// Arguments returns arguments.
func (a FunctionApplication) Arguments() []Atom {
	return a.arguments
}

// ConvertTypes converts types.
func (a FunctionApplication) ConvertTypes(func(types.Type) types.Type) Expression {
	return a
}

func (a FunctionApplication) isExpression() {}
