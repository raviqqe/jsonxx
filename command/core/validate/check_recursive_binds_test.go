package validate

import (
	"testing"

	"github.com/raviqqe/lazy-ein/command/core/ast"
	"github.com/raviqqe/lazy-ein/command/core/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateError(t *testing.T) {
	tt := types.NewAlgebraic(types.NewConstructor())

	assert.Error(
		t,
		checkRecursiveBinds(
			ast.NewModule(
				nil,
				[]ast.Bind{
					ast.NewBind(
						"x",
						ast.NewVariableLambda(
							nil,
							ast.NewFunctionApplication(ast.NewVariable("x"), nil),
							types.NewBoxed(tt),
						),
					),
				},
			),
		),
	)
}

func TestValidateErrorWithLetExpressions(t *testing.T) {
	tt := types.NewAlgebraic(types.NewConstructor())

	assert.Error(
		t,
		checkRecursiveBinds(
			ast.NewModule(
				nil,
				[]ast.Bind{
					ast.NewBind(
						"x",
						ast.NewVariableLambda(
							nil,
							ast.NewLet(
								[]ast.Bind{
									ast.NewBind(
										"y",
										ast.NewVariableLambda(
											[]ast.Argument{ast.NewArgument("y", types.NewBoxed(tt))},
											ast.NewFunctionApplication(ast.NewVariable("y"), nil),
											types.NewBoxed(tt),
										),
									),
								},
								ast.NewFunctionApplication(ast.NewVariable("y"), nil),
							),
							types.NewBoxed(tt),
						),
					),
				},
			),
		),
	)
}
