package ast

import "github.com/raviqqe/lazy-ein/command/core/types"

// Let is a let expression.
type Let struct {
	binds      []Bind
	expression Expression
}

// NewLet creates a let expression.
func NewLet(bs []Bind, e Expression) Let {
	return Let{bs, e}
}

// Binds returns binds.
func (l Let) Binds() []Bind {
	return l.binds
}

// Expression returns an expression.
func (l Let) Expression() Expression {
	return l.expression
}

// VisitExpressions visits expressions.
func (l Let) VisitExpressions(f func(Expression) error) error {
	for _, b := range l.binds {
		if err := b.VisitExpressions(f); err != nil {
			return err
		}
	}

	if err := l.expression.VisitExpressions(f); err != nil {
		return err
	}

	return f(l)
}

// ConvertTypes converts types.
func (l Let) ConvertTypes(f func(types.Type) types.Type) Expression {
	bs := make([]Bind, 0, len(l.binds))

	for _, b := range l.binds {
		bs = append(bs, b.ConvertTypes(f))
	}

	return Let{bs, l.expression.ConvertTypes(f)}
}

// RenameVariables renames variables.
func (l Let) RenameVariables(vs map[string]string) Expression {
	ss := make([]string, 0, len(l.binds))

	for _, b := range l.binds {
		ss = append(ss, b.Name())
	}

	vs = removeVariables(vs, ss...)
	bs := make([]Bind, 0, len(l.binds))

	for _, b := range l.binds {
		bs = append(bs, b.RenameVariables(vs))
	}

	return Let{bs, l.expression.RenameVariables(vs)}
}

func (Let) isExpression() {}
