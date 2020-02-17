package desugar

import (
	"testing"

	"github.com/raviqqe/lazy-ein/command/ast"
	"github.com/raviqqe/lazy-ein/command/types"
	"github.com/stretchr/testify/assert"
)

func TestDesugarBinaryOperations(t *testing.T) {
	for _, ms := range [][2]ast.Module{
		// Empty modules
		{
			ast.NewModule("", ast.NewExport(), nil, []ast.Bind{}),
			ast.NewModule("", ast.NewExport(), nil, []ast.Bind{}),
		},
		// Arguments
		{
			ast.NewModule(
				"",
				ast.NewExport(),
				nil,
				[]ast.Bind{
					ast.NewBind(
						"a",
						types.NewUnboxed(types.NewNumber(nil), nil),
						ast.NewUnboxed(ast.NewNumber(42)),
					),
					ast.NewBind(
						"x",
						types.NewNumber(nil),
						ast.NewBinaryOperation(
							ast.Add,
							ast.NewVariable("a"),
							ast.NewLet(
								[]ast.Bind{ast.NewBind("y", types.NewUnknown(nil), ast.NewNumber(42))},
								ast.NewVariable("y"),
							),
						),
					),
				},
			),
			ast.NewModule(
				"",
				ast.NewExport(),
				nil,
				[]ast.Bind{
					ast.NewBind(
						"a",
						types.NewUnboxed(types.NewNumber(nil), nil),
						ast.NewUnboxed(ast.NewNumber(42)),
					),
					ast.NewBind(
						"x",
						types.NewNumber(nil),
						ast.NewLet(
							[]ast.Bind{
								ast.NewBind(
									"$binary-operation.argument-0",
									types.NewUnknown(nil),
									ast.NewLet(
										[]ast.Bind{ast.NewBind("y", types.NewUnknown(nil), ast.NewNumber(42))},
										ast.NewVariable("y"),
									),
								),
							},
							ast.NewBinaryOperation(
								ast.Add,
								ast.NewVariable("a"),
								ast.NewVariable("$binary-operation.argument-0"),
							),
						),
					),
				},
			),
		},
	} {
		assert.Equal(t, ms[1], desugarBinaryOperations(ms[0]))
	}
}
