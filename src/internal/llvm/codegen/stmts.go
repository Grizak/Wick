package codegen

import "github.com/Grizak/Wick/src/internal/ast"

func (g *Generator) generateStatement(stmt *ast.NodeStatement) error {
	if stmt.Exit != nil {
		return g.generateExit(stmt.Exit)
	}
	if stmt.VarDecl != nil {
		return g.generateVarDecl(stmt.VarDecl)
	}
	if stmt.VarAssign != nil {
		return g.generateVarAssign(stmt.VarAssign)
	}
	if stmt.Block != nil {
		return g.generateBlock(stmt.Block)
	}
	if stmt.If != nil {
		return g.generateIf(stmt.If)
	}
	if stmt.For != nil {
		return g.generateFor(stmt.For)
	}
	if stmt.Break != nil {
		return g.generateBreak(stmt.Break)
	}
	if stmt.Continue != nil {
		return g.generateContinue(stmt.Continue)
	}
	if stmt.Return != nil {
		return g.generateReturn(stmt.Return)
	}
	return nil
}

func (g *Generator) generateBlock(block *ast.NodeBlock) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	for _, stmt := range block.Statements {
		g.generateStatement(&stmt)
	}
	return nil
}
