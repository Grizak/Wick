package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/utils"
)

func (g *Generator) generateStatement(stmt *ast.NodeStatement) error {
	utils.Debug(fmt.Sprintf("GEN STMT: Exit=%v VarDecl=%v VarAssign=%v Block=%v If=%v For=%v Return=%v Import=%v Extern=%v",
		stmt.Exit != nil, stmt.VarDecl != nil, stmt.VarAssign != nil,
		stmt.Block != nil, stmt.If != nil, stmt.For != nil,
		stmt.Return != nil, stmt.Import != nil, stmt.Extern != nil),
	)
	if stmt.VarDecl != nil {
		utils.Debug(fmt.Sprintf("  VarDecl.Name=%q FuncDecl=%v", stmt.VarDecl.Name, stmt.VarDecl.Expr.FuncDecl != nil))
	}
	if stmt.Extern != nil {
		utils.Debug(fmt.Sprintf("  Extern.Name=%q", stmt.Extern.Name))
	}
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
	case stmt.Import != nil:
		return g.generateImport(stmt.Import)
	case stmt.Extern != nil:
		return g.generateExtern(stmt.Extern)
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
