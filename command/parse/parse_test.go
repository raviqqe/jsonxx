package parse

import (
	"testing"

	"github.com/ein-lang/ein/command/ast"
	"github.com/ein-lang/ein/command/debug"
	"github.com/ein-lang/ein/command/types"
	"github.com/raviqqe/parcom"
	"github.com/stretchr/testify/assert"
)

func TestParseWithEmptySource(t *testing.T) {
	x, err := Parse("", "")

	assert.Equal(t, ast.NewModule("", []ast.Bind{}), x)
	assert.Nil(t, err)
}

func TestParseError(t *testing.T) {
	_, err := Parse("foo.ein", "bar")

	assert.Error(t, err)
	assert.Equal(t, "foo.ein:1:4:\tbar", err.(debug.Error).DebugInformation().String())
}

func TestStateModule(t *testing.T) {
	for _, s := range []string{
		"x : Number\nx = 42",
		"x : Number\nx = 42\ny : Number\ny = 42",
		"f : Number -> Number\nf x = 42",
		"f : Number -> Number -> Number\nf x y = 42",
		"f : Number -> Number\nf x = let y = x in y",
	} {
		_, err := newState("", s).module("")()
		assert.Nil(t, err)
	}
}

func TestStateModuleError(t *testing.T) {
	for _, s := range []string{
		"x : Number",
		"x : Number\nx = 42\n  y : Number\n  y = 42",
		" x : Number\n x = 42",
	} {
		_, err := newState("", s).module("")()
		assert.Error(t, err)
	}
}

func TestStateModuleErrorWithErrorMessage(t *testing.T) {
	_, err := newState("", "x : Number\nx = 🗿").module("")()

	assert.Error(t, err)
	assert.Equal(t, "invalid character '🗿'", err.Error())
	assert.Equal(t, 2, err.(parcom.Error).Line())
	assert.Equal(t, 5, err.(parcom.Error).Column())
}

func TestStateBind(t *testing.T) {
	for _, s := range []string{
		"x : Number\nx = 42",
		"f : Number -> Number\nf x = 42",
		"f : Number -> Number -> Number\nf x y = 42",
		"x :\n Number\nx = 42",
		"x : Number\nx =\n 42",
	} {
		_, err := newState("", s).bind()()
		assert.Nil(t, err)
	}
}

func TestStateBindWithVariableBind(t *testing.T) {
	x, err := Parse("", "x : Number\nx = 42")

	assert.Equal(
		t,
		ast.NewModule(
			"",
			[]ast.Bind{
				ast.NewBind(
					"x",
					types.NewNumber(debug.NewInformation("", 1, 5, "x : Number")),
					ast.NewNumber(42),
				),
			},
		),
		x,
	)
	assert.Nil(t, err)
}

func TestStateBindErrorWithInvalidIndents(t *testing.T) {
	_, err := newState("", "x : Number\n x = 42").bind()()
	assert.Error(t, err)
}

func TestStateBindErrorWithInconsistentIdentifiers(t *testing.T) {
	_, err := newState("", "x : Number\ny = 42").bind()()
	assert.Error(t, err)
}

func TestStateIdentifier(t *testing.T) {
	for _, s := range []string{"x", "az", "a09"} {
		_, err := newState("", s).identifier()()
		assert.Nil(t, err)
	}
}

func TestStateIdentifierError(t *testing.T) {
	for _, s := range []string{"0", "1x", "let"} {
		_, err := newState("", s).identifier()()
		assert.Error(t, err)
	}
}

func TestStateIdentifierErrorWithKeywords(t *testing.T) {
	_, err := newState("foo.ein", "let").identifier()()

	assert.Equal(t, "SyntaxError: 'let' is a keyword", err.Error())
	assert.Equal(t, "foo.ein:1:1:\tlet", err.(debug.Error).DebugInformation().String())
}

func TestStateNumberLiteral(t *testing.T) {
	for _, s := range []string{
		"1",
		"42",
		"1.0",
		"-1",
		"0",
	} {
		_, err := newState("", s).numberLiteral()()
		assert.Nil(t, err)
	}
}

func TestStateNumberLiteralError(t *testing.T) {
	for _, s := range []string{
		"- 42",
		"01",
		"1.",
	} {
		s := newState("", s)
		_, err := s.Exhaust(s.numberLiteral())()
		assert.Error(t, err)
	}
}

func TestStateListLiteral(t *testing.T) {
	for _, s := range []string{
		"[]",
		"[42]",
		"[42, 42]",
		"[42, 42,]",
		"[[42]]",
	} {
		_, err := newState("", s).listLiteral()()
		assert.Nil(t, err)
	}
}

func TestStateListLiteralError(t *testing.T) {
	for _, s := range []string{
		"[,]",
		"[,42]",
		"[42,,]",
	} {
		_, err := newState("", s).listLiteral()()
		assert.Error(t, err)
	}
}

func TestStateVariable(t *testing.T) {
	_, err := newState("", "x").variable()()
	assert.Nil(t, err)
}

func TestStateApplication(t *testing.T) {
	for s, a := range map[string]ast.Application{
		"f x": ast.NewApplication(ast.NewVariable("f"), []ast.Expression{ast.NewVariable("x")}),
		"f x y": ast.NewApplication(
			ast.NewVariable("f"),
			[]ast.Expression{ast.NewVariable("x"), ast.NewVariable("y")},
		),
		"f (f x) y": ast.NewApplication(
			ast.NewVariable("f"),
			[]ast.Expression{
				ast.NewApplication(ast.NewVariable("f"), []ast.Expression{ast.NewVariable("x")}),
				ast.NewVariable("y"),
			},
		),
		"(f x) x": ast.NewApplication(
			ast.NewApplication(ast.NewVariable("f"), []ast.Expression{ast.NewVariable("x")}),
			[]ast.Expression{ast.NewVariable("x")},
		),
	} {
		aa, err := newState("", s).application()()

		assert.Nil(t, err)
		assert.Equal(t, a, aa)
	}
}

func TestStateLet(t *testing.T) {
	for _, s := range []string{
		"let x = 42 in 42",
		"let\n x = 42 in 42",
		"let x = 42\n    y = 42 in 42",
		"let x = 42\nin 42",
	} {
		s := newState("", s)
		_, err := s.Exhaust(s.let())()
		assert.Nil(t, err)
	}
}

func TestStateLetError(t *testing.T) {
	for _, s := range []string{
		"let\nx = 42 in 42",
		"let\n x =\n 42 in 42",
		"let x = 42\n y = 42 in 42",
	} {
		s := newState("", s)
		_, err := s.Exhaust(s.let())()
		assert.Error(t, err)
	}
}

func TestStateCaseOf(t *testing.T) {
	for _, c := range []struct {
		source string
		ast    ast.Case
	}{
		{
			"case 1 of 2 -> 3",
			ast.NewCaseWithoutDefault(
				ast.NewNumber(1),
				types.NewUnknown(debug.NewInformation("", 1, 1, "case 1 of 2 -> 3")),
				[]ast.Alternative{ast.NewAlternative(ast.NewNumber(2), ast.NewNumber(3))},
			),
		},
		{
			"case 1 of\n 2 -> 3\n 4 -> 5",
			ast.NewCaseWithoutDefault(
				ast.NewNumber(1),
				types.NewUnknown(debug.NewInformation("", 1, 1, "case 1 of")),
				[]ast.Alternative{
					ast.NewAlternative(ast.NewNumber(2), ast.NewNumber(3)),
					ast.NewAlternative(ast.NewNumber(4), ast.NewNumber(5)),
				},
			),
		},
		{
			"case 1 of x -> 2",
			ast.NewCase(
				ast.NewNumber(1),
				types.NewUnknown(debug.NewInformation("", 1, 1, "case 1 of x -> 2")),
				[]ast.Alternative{},
				ast.NewDefaultAlternative("x", ast.NewNumber(2)),
			),
		},
		{
			"case 1 of\n 2 -> 3\n x -> 4",
			ast.NewCase(
				ast.NewNumber(1),
				types.NewUnknown(debug.NewInformation("", 1, 1, "case 1 of")),
				[]ast.Alternative{ast.NewAlternative(ast.NewNumber(2), ast.NewNumber(3))},
				ast.NewDefaultAlternative("x", ast.NewNumber(4)),
			),
		},
	} {
		a, err := newState("", c.source).caseOf()()

		assert.Nil(t, err)
		assert.Equal(t, c.ast, a)
	}
}

func TestStateExpressionWithOperators(t *testing.T) {
	for _, c := range []struct {
		source     string
		expression ast.Expression
	}{
		{
			"1 + 1",
			ast.NewBinaryOperation(ast.Add, ast.NewNumber(1), ast.NewNumber(1)),
		},
		{
			"(1 + 1) + 1",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewBinaryOperation(ast.Add, ast.NewNumber(1), ast.NewNumber(1)),
				ast.NewNumber(1),
			),
		},
		{
			"1 + (1 + 1)",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewNumber(1),
				ast.NewBinaryOperation(ast.Add, ast.NewNumber(1), ast.NewNumber(1)),
			),
		},
		{
			"1 + 1 + 1",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewBinaryOperation(ast.Add, ast.NewNumber(1), ast.NewNumber(1)),
				ast.NewNumber(1),
			),
		},
		{
			"1 + 1 * 1",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewNumber(1),
				ast.NewBinaryOperation(ast.Multiply, ast.NewNumber(1), ast.NewNumber(1)),
			),
		},
		{
			"1 + 1 * 1 + 1",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewBinaryOperation(
					ast.Add,
					ast.NewNumber(1),
					ast.NewBinaryOperation(ast.Multiply, ast.NewNumber(1), ast.NewNumber(1)),
				),
				ast.NewNumber(1),
			),
		},
		{
			"1 + 1 * 1 + 1 * 1 / 2 - 3",
			ast.NewBinaryOperation(
				ast.Subtract,
				ast.NewBinaryOperation(
					ast.Add,
					ast.NewBinaryOperation(
						ast.Add,
						ast.NewNumber(1),
						ast.NewBinaryOperation(ast.Multiply, ast.NewNumber(1), ast.NewNumber(1)),
					),
					ast.NewBinaryOperation(
						ast.Divide,
						ast.NewBinaryOperation(
							ast.Multiply,
							ast.NewNumber(1),
							ast.NewNumber(1),
						),
						ast.NewNumber(2),
					),
				),
				ast.NewNumber(3),
			),
		},
		{
			"f x + 1",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewApplication(ast.NewVariable("f"), []ast.Expression{ast.NewVariable("x")}),
				ast.NewNumber(1),
			),
		},
		{
			"1 + f x",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewNumber(1),
				ast.NewApplication(ast.NewVariable("f"), []ast.Expression{ast.NewVariable("x")}),
			),
		},
		{
			"1 + let x = y in 2 - 3",
			ast.NewBinaryOperation(
				ast.Add,
				ast.NewNumber(1),
				ast.NewLet(
					[]ast.Bind{
						ast.NewBind(
							"x",
							types.NewUnknown(debug.NewInformation("", 1, 9, "1 + let x = y in 2 - 3")),
							ast.NewVariable("y"),
						),
					},
					ast.NewBinaryOperation(
						ast.Subtract,
						ast.NewNumber(2),
						ast.NewNumber(3),
					),
				),
			),
		},
	} {
		s := newState("", c.source)
		e, err := s.Exhaust(s.expression())()

		assert.Nil(t, err)
		assert.Equal(t, c.expression, e)
	}
}

func TestStateType(t *testing.T) {
	for _, s := range []string{
		"Number",
		"Number -> Number",
		"Number -> Number -> Number",
		"(Number -> Number) -> Number",
		"[Number]",
		"[[Number]]",
	} {
		_, err := newState("", s).typ()()
		assert.Nil(t, err)
	}
}

func TestStateTypeWithMultipleArguments(t *testing.T) {
	s := "Number -> Number -> Number"
	x, err := newState("", s).typ()()

	assert.Equal(
		t,
		types.NewFunction(
			types.NewNumber(debug.NewInformation("", 1, 1, s)),
			types.NewFunction(
				types.NewNumber(debug.NewInformation("", 1, 11, s)),
				types.NewNumber(debug.NewInformation("", 1, 21, s)),
				debug.NewInformation("", 1, 11, s),
			),
			debug.NewInformation("", 1, 1, s),
		),
		x,
	)
	assert.Nil(t, err)
}
