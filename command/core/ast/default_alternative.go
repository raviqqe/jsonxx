package ast

import "github.com/raviqqe/lazy-ein/command/core/types"

// DefaultAlternative is a default alternative.
type DefaultAlternative struct {
	variable   string
	expression Expression
}

// NewDefaultAlternative creates a default alternative.
func NewDefaultAlternative(s string, e Expression) DefaultAlternative {
	return DefaultAlternative{s, e}
}

// Variable returns a bound variable.
func (a DefaultAlternative) Variable() string {
	return a.variable
}

// Expression is an expression.
func (a DefaultAlternative) Expression() Expression {
	return a.expression
}

// VisitExpressions visits expressions.
func (a DefaultAlternative) VisitExpressions(f func(Expression) error) error {
	return a.expression.VisitExpressions(f)
}

// ConvertTypes converts types.
func (a DefaultAlternative) ConvertTypes(f func(types.Type) types.Type) DefaultAlternative {
	return DefaultAlternative{a.variable, a.expression.ConvertTypes(f)}
}

// RenameVariables renames variables.
func (a DefaultAlternative) RenameVariables(vs map[string]string) DefaultAlternative {
	return DefaultAlternative{
		a.variable,
		a.expression.RenameVariables(removeVariables(vs, a.variable)),
	}
}
