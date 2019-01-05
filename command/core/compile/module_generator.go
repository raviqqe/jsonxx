package compile

import (
	"github.com/ein-lang/ein/command/core/ast"
	"github.com/ein-lang/ein/command/core/compile/llir"
	"github.com/ein-lang/ein/command/core/compile/names"
	"github.com/ein-lang/ein/command/core/types"
	"llvm.org/llvm/bindings/go/llvm"
)

const (
	environmentArgumentName  = "environment"
	maxCodeOptimizationLevel = 3
	maxSizeOptimizationLevel = 2
)

type moduleGenerator struct {
	module          llvm.Module
	globalVariables map[string]llvm.Value
	typeGenerator   typeGenerator
}

func newModuleGenerator(m llvm.Module, ds []ast.TypeDefinition) (*moduleGenerator, error) {
	tg, err := newTypeGenerator(m, ds)

	if err != nil {
		return nil, err
	}

	cg := newConstructorGenerator(m, tg)

	for _, d := range ds {
		if t, ok := d.Type().(types.Algebraic); ok {
			if err := cg.Generate(t); err != nil {
				return nil, err
			}
		}
	}

	return &moduleGenerator{m, map[string]llvm.Value{}, tg}, nil
}

func (g *moduleGenerator) Generate(bs []ast.Bind) error {
	for _, b := range bs {
		t, err := g.typeGenerator.GenerateSizedClosure(b.Lambda())

		if err != nil {
			return err
		}

		g.globalVariables[b.Name()] = llvm.AddGlobal(g.module, t, b.Name())
	}

	for _, b := range bs {
		v := g.globalVariables[b.Name()]
		f, err := g.createLambda(b.Name(), b.Lambda())

		if err != nil {
			return err
		} else if err := llvm.VerifyFunction(f, llvm.AbortProcessAction); err != nil {
			return err
		}

		v.SetInitializer(
			llvm.ConstStruct(
				[]llvm.Value{f, llvm.ConstNull(v.Type().ElementType().StructElementTypes()[1])},
				false,
			),
		)
	}

	// nolint: gotype
	if err := llvm.VerifyModule(g.module, llvm.AbortProcessAction); err != nil {
		return err
	}

	g.optimize()

	return llvm.VerifyModule(g.module, llvm.AbortProcessAction)
}

func (g *moduleGenerator) createLambda(n string, l ast.Lambda) (llvm.Value, error) {
	t, err := g.typeGenerator.GenerateLambdaEntryFunction(l)

	if err != nil {
		return llvm.Value{}, err
	}

	f := llir.AddFunction(g.module, names.ToEntry(n), t)
	b := llvm.NewBuilder()
	b.SetInsertPointAtEnd(llvm.AddBasicBlock(f, ""))

	vs, err := g.createLogicalEnvironment(f, b, l)

	if err != nil {
		return llvm.Value{}, err
	}

	v, err := newFunctionBodyGenerator(b, vs, g.createLambda, g.typeGenerator).Generate(l.Body())

	if err != nil {
		return llvm.Value{}, err
	} else if _, ok := l.ResultType().(types.Boxed); ok && l.IsThunk() {
		// TODO: Steal child thunks in a thread-safe way.
		// TODO: Use loop to unbox children recursively.
		v = forceThunk(b, v, g.typeGenerator)
	}

	if l.IsUpdatable() {
		b.CreateStore(v, b.CreateBitCast(f.FirstParam(), llir.PointerType(v.Type()), ""))
		b.CreateStore(
			g.createUpdatedEntryFunction(n, f.Type().ElementType()),
			b.CreateGEP(
				b.CreateBitCast(f.FirstParam(), llir.PointerType(f.Type()), ""),
				[]llvm.Value{llvm.ConstIntFromString(g.typeGenerator.GenerateConstructorTag(), "-1", 10)},
				"",
			),
		)
	}

	b.CreateRet(v)

	return f, nil
}

func (g *moduleGenerator) createUpdatedEntryFunction(n string, t llvm.Type) llvm.Value {
	f := llir.AddFunction(g.module, names.ToNormalFormEntry(n), t)
	f.FirstParam().SetName(environmentArgumentName)

	b := llvm.NewBuilder()
	b.SetInsertPointAtEnd(llvm.AddBasicBlock(f, ""))
	b.CreateRet(
		b.CreateLoad(
			b.CreateBitCast(
				f.FirstParam(), llir.PointerType(f.Type().ElementType().ReturnType()),
				""),
			"",
		),
	)

	return f
}

func (g moduleGenerator) createLogicalEnvironment(f llvm.Value, b llvm.Builder, l ast.Lambda) (map[string]llvm.Value, error) {
	vs := make(map[string]llvm.Value, len(g.globalVariables))

	for k, v := range g.globalVariables {
		if v.Type().ElementType().StructElementTypes()[1].ArrayLength() != 0 {
			v = b.CreateBitCast(
				v,
				llir.PointerType(g.typeGenerator.GenerateUnsizedClosure(v.Type().ElementType())),
				"",
			)
		}

		vs[k] = v
	}

	t, err := g.typeGenerator.GenerateEnvironment(l)

	if err != nil {
		return nil, err
	}

	e := b.CreateBitCast(f.FirstParam(), llir.PointerType(t), "")

	for i, n := range l.FreeVariableNames() {
		vs[n] = b.CreateLoad(b.CreateStructGEP(e, i, ""), "")
	}

	for i, n := range append([]string{environmentArgumentName}, l.ArgumentNames()...) {
		v := f.Param(i)
		v.SetName(n)
		vs[n] = v
	}

	return vs, nil
}

func (g moduleGenerator) optimize() {
	b := llvm.NewPassManagerBuilder()
	b.SetOptLevel(maxCodeOptimizationLevel)
	b.SetSizeLevel(maxSizeOptimizationLevel)

	p := llvm.NewPassManager()
	b.PopulateFunc(p)
	b.Populate(p)

	p.Run(g.module)
}
