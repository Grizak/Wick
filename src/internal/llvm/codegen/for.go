package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
)

func (g *Generator) generateFor(node *ast.NodeFor) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	g.clearMutableStaticValues()

	loopLabel := g.newLabel("loop")
	bodyLabel := g.newLabel("body")
	postLabel := g.newLabel("post")
	endLabel := g.newLabel("loop_end")

	// Save outer loop labels for nested loops
	prevEndLabel := g.loopEndLabel
	prevNextLabel := g.loopNextLabel
	defer func() {
		g.loopEndLabel = prevEndLabel
		g.loopNextLabel = prevNextLabel
	}()

	g.loopEndLabel = endLabel
	g.loopNextLabel = postLabel

	// Init
	if node.Init != nil {
		if err := g.generateStatement(node.Init); err != nil {
			return err
		}
	}

	g.writeLine(fmt.Sprintf("    br label %%%s", loopLabel))
	g.writeLine(fmt.Sprintf("%s:", loopLabel))

	// Condition
	if node.Condition != nil {
		cond, err := g.generateCondition(*node.Condition)
		if err != nil {
			return err
		}
		g.writeLine(fmt.Sprintf("    br i1 %s, label %%%s, label %%%s", cond, bodyLabel, endLabel))
	} else {
		g.writeLine(fmt.Sprintf("    br label %%%s", bodyLabel))
	}

	// Body
	g.writeLine(fmt.Sprintf("%s:", bodyLabel))
	for _, stmt := range node.Body.Statements {
		if err := g.generateStatement(&stmt); err != nil {
			return err
		}
	}
	g.writeLine(fmt.Sprintf("    br label %%%s", postLabel))

	// Post
	g.writeLine(fmt.Sprintf("%s:", postLabel))
	if node.Post != nil {
		if err := g.generateStatement(node.Post); err != nil {
			return err
		}
	}
	g.writeLine(fmt.Sprintf("    br label %%%s", loopLabel))

	g.writeLine(fmt.Sprintf("%s:", endLabel))
	return nil
}

func (g *Generator) clearMutableStaticValues() {
	scope := g.scope
	for scope != nil {
		for name, sym := range scope.symbols {
			if !sym.isConst {
				sym.staticValue = nil
				scope.symbols[name] = sym
			}
		}
		scope = scope.parent
	}
}

func (g *Generator) generateBreak(node *ast.NodeBreak) error {
	if g.loopEndLabel == "" {
		return g.error("break outside of loop", node.Pos)
	}
	g.writeLine(fmt.Sprintf("    br label %%%s", g.loopEndLabel))
	return nil
}

func (g *Generator) generateContinue(node *ast.NodeContinue) error {
	if g.loopNextLabel == "" {
		return g.error("continue outside of loop", node.Pos)
	}
	g.writeLine(fmt.Sprintf("    br label %%%s", g.loopNextLabel))
	return nil
}
