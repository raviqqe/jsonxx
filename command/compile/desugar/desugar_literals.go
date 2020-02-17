package desugar

import (
	"github.com/raviqqe/lazy-ein/command/ast"
	"github.com/raviqqe/lazy-ein/command/compile/desugar/names"
	"github.com/raviqqe/lazy-ein/command/types"
)

func desugarLiterals(m ast.Module) ast.Module {
	g := names.NewNameGenerator("")
	bs := []ast.Bind{}

	for _, b := range m.Binds() {
		if l, ok := b.Expression().(ast.Literal); ok {
			bs = append(
				bs,
				ast.NewBind(
					b.Name(),
					types.NewUnboxed(b.Type(), b.Type().DebugInformation()),
					ast.NewUnboxed(l),
				),
			)

			continue
		}

		bs = append(bs, b.ConvertExpressions(func(e ast.Expression) ast.Expression {
			l, ok := e.(ast.Literal)

			if !ok {
				return e
			}

			s := g.Generate("literal")

			// TODO: Handle other literals.
			switch l := l.(type) {
			case ast.Number:
				bs = append(
					bs,
					ast.NewBind(s, types.NewUnboxed(types.NewNumber(nil), nil), ast.NewUnboxed(l)),
				)
				return ast.NewVariable(s)
			}

			panic("unreachable")
		}))
	}

	return ast.NewModule(m.Name(), m.Export(), m.Imports(), bs)
}
