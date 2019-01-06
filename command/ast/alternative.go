package ast

import "github.com/ein-lang/ein/command/types"

// Alternative is an alternative.
type Alternative struct {
	literal    Literal
	expression Expression
}

// NewAlternative creates an alternative.
func NewAlternative(l Literal, e Expression) Alternative {
	return Alternative{l, e}
}

// Literal returns a literal pattern.
func (a Alternative) Literal() Literal {
	return a.literal
}

// Expression is an expression.
func (a Alternative) Expression() Expression {
	return a.expression
}

// ConvertExpressions visits expressions.
func (a Alternative) ConvertExpressions(f func(Expression) Expression) Node {
	return NewAlternative(a.literal, a.expression.ConvertExpressions(f).(Expression))
}

// VisitTypes visits types.
func (a Alternative) VisitTypes(f func(types.Type) error) error {
	if err := a.literal.VisitTypes(f); err != nil {
		return err
	}

	return a.expression.VisitTypes(f)
}
