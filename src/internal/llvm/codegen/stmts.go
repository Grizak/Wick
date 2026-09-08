package codegen

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"tinygo.org/x/go-llvm"
)

func (g *Generator) generateStatement(mod llvm.Module, builder llvm.Builder, ctx llvm.Context, stmt *ast.NodeStatement) error {
	if stmt.Exit != nil {
		return g.generateExit(mod, builder, ctx, stmt.Exit)
	}
	if stmt.VarDecl != nil {
		return g.generateVarDecl(mod, builder, ctx, stmt.VarDecl)
	}
	if stmt.VarAssign != nil {
		return g.generateVarAssign(mod, builder, ctx, stmt.VarAssign)
	}
	if stmt.Block != nil {
		return g.generateBlock(mod, builder, ctx, stmt.Block)
	}
	if stmt.If != nil {
		return g.generateIf(mod, builder, ctx, stmt.If)
	}
	if stmt.For != nil {
		return g.generateFor(mod, builder, ctx, stmt.For)
	}
	if stmt.Break != nil {
		return g.generateBreak(mod, builder, ctx, stmt.Break)
	}
	if stmt.Continue != nil {
		return g.generateContinue(mod, builder, ctx, stmt.Continue)
	}
	if stmt.Return != nil {
		return g.generateReturn(mod, builder, ctx, stmt.Return)
	}
	return nil
}

func (g *Generator) generateBlock(mod llvm.Module, builder llvm.Builder, ctx llvm.Context, block *ast.NodeBlock) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	for _, stmt := range block.Statements {
		g.generateStatement(mod, builder, ctx, &stmt)
	}
	return nil
}
