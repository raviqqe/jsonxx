package ast

import "github.com/ein-lang/ein/command/core/types"

// Lambda is a lambda form.
type Lambda struct {
	freeVariables []Argument
	updatable     bool
	arguments     []Argument
	body          Expression
	resultType    types.Type
}

// NewLambda creates a lambda form.
func NewLambda(vs []Argument, u bool, as []Argument, e Expression, t types.Type) Lambda {
	return Lambda{vs, u, as, e, t}
}

// Arguments returns arguments.
func (l Lambda) Arguments() []Argument {
	return l.arguments
}

// ArgumentNames returns argument names.
func (l Lambda) ArgumentNames() []string {
	return argumentsToNames(l.arguments)
}

// ArgumentTypes returns argument types.
func (l Lambda) ArgumentTypes() []types.Type {
	return argumentsToTypes(l.arguments)
}

// Body returns a body expression.
func (l Lambda) Body() Expression {
	return l.body
}

// ResultType returns a result type.
func (l Lambda) ResultType() types.Type {
	return l.resultType
}

// FreeVariableNames returns free varriable names.
func (l Lambda) FreeVariableNames() []string {
	return argumentsToNames(l.freeVariables)
}

// FreeVariableTypes returns free varriable types.
func (l Lambda) FreeVariableTypes() []types.Type {
	return argumentsToTypes(l.freeVariables)
}

// IsUpdatable returns true if the lambda form is updatable, or false otherwise.
func (l Lambda) IsUpdatable() bool {
	return l.updatable
}

// IsThunk returns true if the lambda form is a thunk, or false otherwise.
func (l Lambda) IsThunk() bool {
	return len(l.arguments) == 0
}

// ConvertTypes converts types.
func (l Lambda) ConvertTypes(f func(types.Type) types.Type) Lambda {
	vs := make([]Argument, 0, len(l.freeVariables))

	for _, v := range l.freeVariables {
		vs = append(vs, v.ConvertTypes(f))
	}

	as := make([]Argument, 0, len(l.arguments))

	for _, a := range l.arguments {
		as = append(as, a.ConvertTypes(f))
	}

	return Lambda{vs, l.updatable, as, l.body.ConvertTypes(f), l.resultType.ConvertTypes(f)}
}

func argumentsToNames(as []Argument) []string {
	ss := make([]string, 0, len(as))

	for _, a := range as {
		ss = append(ss, a.Name())
	}

	return ss
}

func argumentsToTypes(as []Argument) []types.Type {
	ts := make([]types.Type, 0, len(as))

	for _, a := range as {
		ts = append(ts, a.Type())
	}

	return ts
}
