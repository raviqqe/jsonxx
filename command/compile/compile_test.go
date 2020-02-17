package compile

import (
	"testing"

	"github.com/raviqqe/lazy-ein/command/ast"
	"github.com/raviqqe/lazy-ein/command/compile/desugar"
	"github.com/raviqqe/lazy-ein/command/compile/metadata"
	"github.com/raviqqe/lazy-ein/command/compile/tinfer"
	coreast "github.com/raviqqe/lazy-ein/command/core/ast"
	coretypes "github.com/raviqqe/lazy-ein/command/core/types"
	"github.com/raviqqe/lazy-ein/command/types"
	"github.com/stretchr/testify/assert"
)

var numberAlgebraic = coretypes.NewAlgebraic(coretypes.NewConstructor(coretypes.NewFloat64()))
var numberConstructor = coreast.NewConstructor(numberAlgebraic, 0)
var listAlgebraic = coretypes.Unbox(
	types.NewList(types.NewNumber(nil), nil).ToCore(),
).(coretypes.Algebraic)
var consConstructor = coreast.NewConstructor(listAlgebraic, 0)
var nilConstructor = coreast.NewConstructor(listAlgebraic, 1)

func TestCompileWithEmptySource(t *testing.T) {
	_, err := Compile(ast.NewModule("", ast.NewExport(), nil, []ast.Bind{}), nil)
	assert.Nil(t, err)
}

func TestCompileWithFunctionApplications(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda([]string{"x"}, ast.NewVariable("x")),
				),
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewApplication(
						ast.NewVariable("f"),
						[]ast.Expression{ast.NewNumber(42)},
					),
				),
			},
		),
		nil,
	)

	assert.Nil(t, err)
}

func TestCompileWithNestedFunctionApplications(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda([]string{"x"}, ast.NewVariable("x")),
				),
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewApplication(
						ast.NewVariable("f"),
						[]ast.Expression{
							ast.NewApplication(
								ast.NewVariable("f"),
								[]ast.Expression{ast.NewVariable("x")},
							),
						},
					),
				),
			},
		),
		nil,
	)

	assert.Nil(t, err)
}

func TestCompileWithDeeplyNestedFunctionApplicationsInLambdaExpressions(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda([]string{"x"}, ast.NewVariable("x")),
				),
				ast.NewBind(
					"g",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda(
						[]string{"x"},
						ast.NewApplication(
							ast.NewVariable("f"),
							[]ast.Expression{
								ast.NewApplication(
									ast.NewVariable("f"),
									[]ast.Expression{
										ast.NewApplication(
											ast.NewVariable("f"),
											[]ast.Expression{ast.NewVariable("x")},
										),
									},
								),
							},
						),
					),
				),
			},
		),
		nil,
	)

	assert.Nil(t, err)
}

func TestCompileWithLists(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewList(types.NewNumber(nil), nil),
					ast.NewList(
						types.NewList(types.NewNumber(nil), nil),
						[]ast.ListArgument{ast.NewListArgument(ast.NewNumber(42), false)},
					),
				),
			},
		),
		nil,
	)

	assert.Nil(t, err)
}

func TestCompileErrorWithUnknownVariables(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{ast.NewBind("x", types.NewNumber(nil), ast.NewVariable("y"))},
		),
		nil,
	)
	assert.Error(t, err)
}

func TestCompilePanicWithUntypedGlobals(t *testing.T) {
	assert.Panics(t, func() {
		Compile(
			ast.NewModule(
				"",
				ast.NewExport(),
				nil,
				[]ast.Bind{ast.NewBind("x", types.NewUnknown(nil), ast.NewNumber(42))},
			),
			nil,
		)
	})
}

func TestCompileToCoreWithEmptySource(t *testing.T) {
	m, err := compileToCore(ast.NewModule("", ast.NewExport(), nil, []ast.Bind{}), nil)
	assert.Nil(t, err)

	assert.Equal(t, coreast.NewModule(nil, []coreast.Bind{}), m)
}

func TestCompileToCoreWithVariableBinds(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{ast.NewBind("x", types.NewNumber(nil), ast.NewNumber(42))},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"x",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithFunctionBinds(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(
						types.NewNumber(nil),
						types.NewFunction(
							types.NewNumber(nil),
							types.NewNumber(nil),
							nil,
						),
						nil,
					),
					ast.NewLambda([]string{"x", "y"}, ast.NewNumber(42)),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"$literal-0",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"f",
					coreast.NewFunctionLambda(
						nil,
						[]coreast.Argument{
							coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
							coreast.NewArgument("y", coretypes.NewBoxed(numberAlgebraic)),
						},
						coreast.NewFunctionApplication(coreast.NewVariable("$literal-0"), nil),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithLetExpressions(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewLet(
						[]ast.Bind{ast.NewBind("y", types.NewNumber(nil), ast.NewNumber(42))},
						ast.NewVariable("y"),
					),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"$literal-0",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"x",
					coreast.NewVariableLambda(
						nil,
						coreast.NewLet(
							[]coreast.Bind{
								coreast.NewBind(
									"y",
									coreast.NewVariableLambda(
										nil,
										coreast.NewFunctionApplication(coreast.NewVariable("$literal-0"), nil),
										coretypes.NewBoxed(numberAlgebraic),
									),
								),
							},
							coreast.NewFunctionApplication(coreast.NewVariable("y"), nil),
						),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithLetExpressionsAndFreeVariables(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda(
						[]string{"x"},
						ast.NewLet(
							[]ast.Bind{ast.NewBind("y", types.NewNumber(nil), ast.NewVariable("x"))},
							ast.NewVariable("y"),
						),
					),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"f",
					coreast.NewFunctionLambda(
						nil,
						[]coreast.Argument{
							coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
						},
						coreast.NewLet(
							[]coreast.Bind{
								coreast.NewBind(
									"y",
									coreast.NewVariableLambda(
										[]coreast.Argument{
											coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
										},
										coreast.NewFunctionApplication(coreast.NewVariable("x"), nil),
										coretypes.NewBoxed(numberAlgebraic),
									),
								),
							},
							coreast.NewFunctionApplication(coreast.NewVariable("y"), nil),
						),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithNestedLetExpressionsInLambdaExpressions(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda(
						[]string{"x"},
						ast.NewLet(
							[]ast.Bind{
								ast.NewBind(
									"y", types.NewUnknown(nil), ast.NewLet(
										[]ast.Bind{ast.NewBind("z", types.NewUnknown(nil), ast.NewVariable("x"))},
										ast.NewVariable("z"),
									),
								),
							},
							ast.NewVariable("y"),
						),
					),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"f",
					coreast.NewFunctionLambda(
						nil,
						[]coreast.Argument{
							coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
						},
						coreast.NewLet(
							[]coreast.Bind{
								coreast.NewBind(
									"y",
									coreast.NewVariableLambda(
										[]coreast.Argument{
											coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
										},
										coreast.NewLet(
											[]coreast.Bind{
												coreast.NewBind(
													"z",
													coreast.NewVariableLambda(
														[]coreast.Argument{
															coreast.NewArgument("x", coretypes.NewBoxed(numberAlgebraic)),
														},
														coreast.NewFunctionApplication(coreast.NewVariable("x"), nil),
														coretypes.NewBoxed(numberAlgebraic),
													),
												),
											},
											coreast.NewFunctionApplication(coreast.NewVariable("z"), nil),
										),
										coretypes.NewBoxed(numberAlgebraic),
									),
								),
							},
							coreast.NewFunctionApplication(coreast.NewVariable("y"), nil),
						),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithLists(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewList(types.NewNumber(nil), nil),
					ast.NewList(
						types.NewUnknown(nil),
						[]ast.ListArgument{ast.NewListArgument(ast.NewNumber(42), false)},
					),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"$literal-0",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"x",
					coreast.NewVariableLambda(
						nil,
						coreast.NewLet(
							[]coreast.Bind{
								coreast.NewBind(
									"$nil",
									coreast.NewVariableLambda(
										nil,
										coreast.NewConstructorApplication(
											coreast.NewConstructor(listAlgebraic, 1),
											nil,
										),
										listAlgebraic,
									),
								),
								coreast.NewBind(
									"$list-0",
									coreast.NewVariableLambda(
										[]coreast.Argument{
											coreast.NewArgument("$nil", coretypes.NewBoxed(listAlgebraic)),
										},
										coreast.NewConstructorApplication(
											coreast.NewConstructor(listAlgebraic, 0),
											[]coreast.Atom{
												coreast.NewVariable("$literal-0"),
												coreast.NewVariable("$nil"),
											},
										),
										listAlgebraic,
									),
								),
							},
							coreast.NewFunctionApplication(coreast.NewVariable("$list-0"), nil),
						),
						coretypes.NewBoxed(listAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithListCaseExpressionsWithoutDefaultAlternatives(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewCaseWithoutDefault(
						ast.NewList(
							types.NewUnknown(nil),
							[]ast.ListArgument{
								ast.NewListArgument(ast.NewNumber(42), false),
							},
						),
						types.NewUnknown(nil),
						[]ast.Alternative{
							ast.NewAlternative(
								ast.NewList(
									types.NewUnknown(nil),
									[]ast.ListArgument{
										ast.NewListArgument(ast.NewNumber(42), false),
									},
								),
								ast.NewNumber(42),
							),
							ast.NewAlternative(ast.NewList(types.NewUnknown(nil), nil), ast.NewNumber(42)),
						},
					),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"$literal-0",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"$literal-1",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"$literal-2",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(42)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"x",
					coreast.NewVariableLambda(
						nil,
						coreast.NewAlgebraicCaseWithoutDefault(
							coreast.NewLet(
								[]coreast.Bind{
									coreast.NewBind(
										"$nil",
										coreast.NewVariableLambda(
											nil,
											coreast.NewConstructorApplication(nilConstructor, nil),
											listAlgebraic,
										),
									),
									coreast.NewBind(
										"$list-0",
										coreast.NewVariableLambda(
											[]coreast.Argument{
												coreast.NewArgument("$nil", coretypes.NewBoxed(listAlgebraic)),
											},
											listConstructorApplication(
												coreast.NewVariable("$literal-2"),
												coreast.NewVariable("$nil"),
											),
											listAlgebraic,
										),
									),
								},
								coreast.NewFunctionApplication(coreast.NewVariable("$list-0"), nil),
							),
							[]coreast.AlgebraicAlternative{
								coreast.NewAlgebraicAlternative(
									consConstructor,
									[]string{"$list-case.head-0", "$list-case.tail-0"},
									coreast.NewPrimitiveCaseWithoutDefault(
										coreast.NewAlgebraicCaseWithoutDefault(
											coreast.NewFunctionApplication(
												coreast.NewVariable("$list-case.head-0"),
												nil,
											),
											[]coreast.AlgebraicAlternative{
												coreast.NewAlgebraicAlternative(
													numberConstructor,
													[]string{"$primitive"},
													coreast.NewFunctionApplication(coreast.NewVariable("$primitive"), nil),
												),
											},
										),
										coretypes.NewFloat64(),
										[]coreast.PrimitiveAlternative{
											coreast.NewPrimitiveAlternative(
												coreast.NewFloat64(42),
												coreast.NewAlgebraicCaseWithoutDefault(
													coreast.NewFunctionApplication(
														coreast.NewVariable("$list-case.tail-0"),
														nil,
													),
													[]coreast.AlgebraicAlternative{
														coreast.NewAlgebraicAlternative(
															nilConstructor,
															nil,
															coreast.NewFunctionApplication(
																coreast.NewVariable("$literal-0"),
																nil,
															),
														),
													},
												),
											),
										},
									),
								),
								coreast.NewAlgebraicAlternative(
									nilConstructor,
									nil,
									coreast.NewFunctionApplication(coreast.NewVariable("$literal-1"), nil),
								),
							},
						),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileToCoreWithBinaryOperations(t *testing.T) {
	m, err := compileToCore(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewBinaryOperation(ast.Add, ast.NewNumber(1), ast.NewNumber(1)),
				),
			},
		),
		nil,
	)
	assert.Nil(t, err)

	assert.Equal(
		t,
		coreast.NewModule(
			nil,
			[]coreast.Bind{
				coreast.NewBind(
					"$literal-0",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(1)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"$literal-1",
					coreast.NewVariableLambda(
						nil,
						numberConstructorApplication(coreast.NewFloat64(1)),
						numberAlgebraic,
					),
				),
				coreast.NewBind(
					"x",
					coreast.NewVariableLambda(
						nil,
						coreast.NewLet(
							[]coreast.Bind{
								coreast.NewBind(
									"$boxedResult",
									coreast.NewVariableLambda(
										nil,
										coreast.NewAlgebraicCaseWithoutDefault(
											coreast.NewFunctionApplication(
												coreast.NewVariable("$literal-0"),
												nil,
											),
											[]coreast.AlgebraicAlternative{
												coreast.NewAlgebraicAlternative(
													numberConstructor,
													[]string{"$lhs"},
													coreast.NewAlgebraicCaseWithoutDefault(
														coreast.NewFunctionApplication(
															coreast.NewVariable("$literal-1"),
															nil,
														),
														[]coreast.AlgebraicAlternative{
															coreast.NewAlgebraicAlternative(
																numberConstructor,
																[]string{"$rhs"},
																coreast.NewPrimitiveCase(
																	coreast.NewPrimitiveOperation(
																		coreast.AddFloat64,
																		[]coreast.Atom{
																			coreast.NewVariable("$lhs"),
																			coreast.NewVariable("$rhs"),
																		},
																	),
																	coretypes.NewFloat64(),
																	nil,
																	coreast.NewDefaultAlternative(
																		"$result",
																		numberConstructorApplication(coreast.NewVariable("$result")),
																	),
																),
															),
														},
													),
												),
											},
										),
										numberAlgebraic,
									),
								),
							},
							coreast.NewFunctionApplication(coreast.NewVariable("$boxedResult"), nil),
						),
						coretypes.NewBoxed(numberAlgebraic),
					),
				),
			},
		),
		m,
	)
}

func TestCompileWithComplexBinaryOperations(t *testing.T) {
	_, err := Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewBinaryOperation(
						ast.Add,
						ast.NewNumber(1),
						ast.NewBinaryOperation(
							ast.Multiply,
							ast.NewNumber(2),
							ast.NewNumber(3),
						),
					),
				),
			},
		),
		nil,
	)

	assert.Nil(t, err)
}

func TestCompileWithCaseExpressions(t *testing.T) {
	for _, c := range []ast.Case{
		ast.NewCase(
			ast.NewNumber(1),
			types.NewUnknown(nil),
			[]ast.Alternative{
				ast.NewAlternative(ast.NewNumber(2), ast.NewNumber(3)),
			},
			ast.NewDefaultAlternative("y", ast.NewVariable("y")),
		),
		ast.NewCaseWithoutDefault(
			ast.NewNumber(1),
			types.NewUnknown(nil),
			[]ast.Alternative{
				ast.NewAlternative(ast.NewNumber(2), ast.NewNumber(3)),
			},
		),
		ast.NewCase(
			ast.NewNumber(1),
			types.NewUnknown(nil),
			nil,
			ast.NewDefaultAlternative("y", ast.NewVariable("y")),
		),
		ast.NewCase(
			ast.NewList(
				types.NewUnknown(nil),
				[]ast.ListArgument{ast.NewListArgument(ast.NewNumber(42), false)},
			),
			types.NewUnknown(nil),
			[]ast.Alternative{
				ast.NewAlternative(
					ast.NewList(
						types.NewUnknown(nil),
						[]ast.ListArgument{ast.NewListArgument(ast.NewNumber(42), false)},
					),
					ast.NewNumber(42),
				),
			},
			ast.NewDefaultAlternative("y", ast.NewNumber(42)),
		),
	} {
		_, err := Compile(
			ast.NewModule(
				"",
				ast.NewExport(),
				nil,
				[]ast.Bind{
					ast.NewBind(
						"x",
						types.NewNumber(nil),
						c,
					),
				},
			),
			nil,
		)

		assert.Nil(t, err)
	}
}

func TestCompileWithVariablesInImportedModules(t *testing.T) {
	m, err := desugarModule(
		ast.NewModule(
			"foo/bar",
			ast.NewExport("x"),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(nil),
					ast.NewNumber(42),
				),
			},
		),
	)
	assert.Nil(t, err)

	_, err = Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			[]ast.Import{ast.NewImport("foo/bar")},
			[]ast.Bind{ast.NewBind("y", types.NewNumber(nil), ast.NewVariable("bar.x"))},
		),
		[]metadata.Module{metadata.NewModule(m)},
	)

	assert.Nil(t, err)
}

func TestCompileWithFunctionsInImportedModules(t *testing.T) {
	m, err := desugarModule(
		ast.NewModule(
			"foo/bar",
			ast.NewExport("f"),
			nil,
			[]ast.Bind{
				ast.NewBind(
					"f",
					types.NewFunction(types.NewNumber(nil), types.NewNumber(nil), nil),
					ast.NewLambda([]string{"x"}, ast.NewVariable("x")),
				),
			},
		),
	)
	assert.Nil(t, err)

	_, err = Compile(
		ast.NewModule(
			"",
			ast.NewExport(),
			[]ast.Import{ast.NewImport("foo/bar")},
			[]ast.Bind{
				ast.NewBind(
					"y",
					types.NewNumber(nil),
					ast.NewApplication(ast.NewVariable("bar.f"), []ast.Expression{ast.NewNumber(42)}),
				),
			},
		),
		[]metadata.Module{metadata.NewModule(m)},
	)

	assert.Nil(t, err)
}

func desugarModule(m ast.Module) (ast.Module, error) {
	m, err := tinfer.InferTypes(desugar.WithoutTypes(m), nil)

	if err != nil {
		return ast.Module{}, err
	}

	return desugar.WithTypes(m), nil
}

func numberConstructorApplication(a coreast.Atom) coreast.Expression {
	return coreast.NewConstructorApplication(numberConstructor, []coreast.Atom{a})
}

func listConstructorApplication(a, aa coreast.Atom) coreast.Expression {
	return coreast.NewConstructorApplication(consConstructor, []coreast.Atom{a, aa})
}
