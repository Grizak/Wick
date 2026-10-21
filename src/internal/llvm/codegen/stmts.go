package codegen

import "github.com/Grizak/Wick/src/internal/ast"

func (g *Generator) generateStatement(stmt *ast.NodeStatement) error {
	switch {
	case stmt.Exit != nil:
		return g.generateExit(stmt.Exit)
	case stmt.VarDecl != nil:
		return g.generateVarDecl(stmt.VarDecl)
	case stmt.VarAssign != nil:
		return g.generateVarAssign(stmt.VarAssign)
	case stmt.Block != nil:
		return g.generateBlock(stmt.Block)
	case stmt.If != nil:
		return g.generateIf(stmt.If)
	case stmt.For != nil:
		return g.generateFor(stmt.For)
	case stmt.Break != nil:
		return g.generateBreak(stmt.Break)
	case stmt.Continue != nil:
		return g.generateContinue(stmt.Continue)
	case stmt.Return != nil:
		return g.generateReturn(stmt.Return)
	default:
		return nil
	}
}

func (g *Generator) generateStatements(statements []ast.NodeStatement) error {
	for i := range statements {
		if g.currentBlockTerminated() {
			break
		}
		if err := g.generateStatement(&statements[i]); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateBlock(block *ast.NodeBlock) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()
	return g.generateStatements(block.Statements)
}
