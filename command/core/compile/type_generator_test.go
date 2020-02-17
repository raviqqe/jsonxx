package compile

import (
	"testing"

	"github.com/raviqqe/lazy-ein/command/core/ast"
	"github.com/raviqqe/lazy-ein/command/core/types"
	"github.com/llvm-mirror/llvm/bindings/go/llvm"
	"github.com/stretchr/testify/assert"
)

func TestTypeGeneratorGenerateSizedPayload(t *testing.T) {
	a := types.NewAlgebraic(types.NewConstructor(types.NewFloat64()))

	for _, c := range []struct {
		lambda ast.Lambda
		size   int
	}{
		{
			lambda: ast.NewVariableLambda(
				nil,
				ast.NewConstructorApplication(ast.NewConstructor(a, 0), []ast.Atom{ast.NewFloat64(42)}),
				a,
			),
			size: 8,
		},
		{
			lambda: ast.NewVariableLambda(
				[]ast.Argument{
					ast.NewArgument("x", types.NewFloat64()),
					ast.NewArgument("y", types.NewFloat64()),
				},
				ast.NewConstructorApplication(ast.NewConstructor(a, 0), []ast.Atom{ast.NewFloat64(42)}),
				a,
			),
			size: 16,
		},
		{
			lambda: ast.NewFunctionLambda(
				nil,
				[]ast.Argument{ast.NewArgument("y", types.NewFloat64())},
				ast.NewConstructorApplication(ast.NewConstructor(a, 0), []ast.Atom{ast.NewFloat64(42)}),
				a,
			),
			size: 0,
		},
	} {
		assert.Equal(
			t,
			c.size,
			newTypeGenerator(llvm.NewModule("foo")).generateSizedPayload(c.lambda.ToDeclaration()).ArrayLength(),
		)
	}
}

func TestTypeGeneratorGenerateWithRecursiveTypes(t *testing.T) {
	for _, t := range []types.Type{
		types.NewAlgebraic(types.NewConstructor(types.NewBoxed(types.NewIndex(0)))),
		types.NewAlgebraic(
			types.NewConstructor(
				types.NewBoxed(
					types.NewAlgebraic(types.NewConstructor(types.NewBoxed(types.NewIndex(1)))),
				),
			),
		),
		types.NewFunction([]types.Type{types.NewIndex(0)}, types.NewFloat64()),
		types.NewFunction(
			[]types.Type{types.NewFunction([]types.Type{types.NewIndex(1)}, types.NewFloat64())},
			types.NewFloat64(),
		),
	} {
		newTypeGenerator(llvm.NewModule("")).Generate(t)
	}
}

func TestTypeGeneratorBytesToWords(t *testing.T) {
	for k, v := range map[int]int{
		0:  0,
		1:  1,
		2:  1,
		7:  1,
		8:  1,
		9:  2,
		16: 2,
		17: 3,
	} {
		assert.Equal(t, v, newTypeGenerator(llvm.NewModule("foo")).bytesToWords(k))
	}
}
