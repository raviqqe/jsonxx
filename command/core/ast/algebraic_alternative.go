package ast

import "github.com/ein-lang/ein/command/core/types"

// AlgebraicAlternative is an algebraic alternative.
type AlgebraicAlternative struct {
	constructor  Constructor
	elementNames []string
	expression   Expression
}

// NewAlgebraicAlternative creates an algebraic alternative.
func NewAlgebraicAlternative(c Constructor, es []string, e Expression) AlgebraicAlternative {
	return AlgebraicAlternative{c, es, e}
}

// Constructor returns a constructor.
func (a AlgebraicAlternative) Constructor() Constructor {
	return a.constructor
}

// ElementNames returns element names.
func (a AlgebraicAlternative) ElementNames() []string {
	return a.elementNames
}

// Expression is an expression.
func (a AlgebraicAlternative) Expression() Expression {
	return a.expression
}

// ConvertTypes converts types.
func (a AlgebraicAlternative) ConvertTypes(f func(types.Type) types.Type) AlgebraicAlternative {
	return AlgebraicAlternative{a.constructor, a.elementNames, a.expression.ConvertTypes(f)}
}
