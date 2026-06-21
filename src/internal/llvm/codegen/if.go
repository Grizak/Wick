package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
)

func (g *Generator) generateIf(node *ast.NodeIf) error {
	// Generate the condition
	cond, err := g.generateCondition(node.Condition)
	if err != nil {
		return err
	}

	// Create unique labels
	thenLabel := g.newLabel("then")
	endLabel := g.newLabel("end")
	elseLabel := endLabel
	if node.Else != nil {
		elseLabel = g.newLabel("else")
	}

	// Conditional branch
	g.writeLine(fmt.Sprintf("    br i1 %s, label %%%s, label %%%s", cond, thenLabel, elseLabel))

	// Then block
	g.writeLine(fmt.Sprintf("%s:", thenLabel))
	g.enterScope()
	g.typeChecker.EnterScope()
	for _, stmt := range node.Then.Statements {
		if err := g.generateStatement(&stmt); err != nil {
			return err
		}
	}
	g.exitScope()
	g.typeChecker.ExitScope()
	g.writeLine(fmt.Sprintf("    br label %%%s", endLabel))

	// Else block
	if node.Else != nil {
		g.writeLine(fmt.Sprintf("%s:", elseLabel))
		g.enterScope()
		g.typeChecker.EnterScope()
		for _, stmt := range node.Else.Statements {
			if err := g.generateStatement(&stmt); err != nil {
				return err
			}
		}
		g.exitScope()
		g.typeChecker.ExitScope()
		g.writeLine(fmt.Sprintf("    br label %%%s", endLabel))
	}

	// End label
	g.writeLine(fmt.Sprintf("%s:", endLabel))
	return nil
}
