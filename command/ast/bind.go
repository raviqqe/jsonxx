package ast

import "github.com/raviqqe/jsonxx/command/types"

// Bind is a bind.
type Bind struct {
	name       string
	typ        types.Type
	expression Expression
}

// NewBind creates a new bind.
func NewBind(n string, t types.Type, e Expression) Bind {
	return Bind{n, t, e}
}

// Name returns a name.
func (b Bind) Name() string {
	return b.name
}

// Type returns a type.
func (b Bind) Type() types.Type {
	return b.typ
}

// Expression returns a expression.
func (b Bind) Expression() Expression {
	return b.expression
}
