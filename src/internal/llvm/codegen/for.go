package codegen

import "github.com/Grizak/Wick/src/internal/ast"

func (g *Generator) generateFor(node *ast.NodeFor) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	g.clearMutableStaticValues()

	conditionBlock := g.newBlock("for_cond")
	bodyBlock := g.newBlock("for_body")
	postBlock := g.newBlock("for_post")
	endBlock := g.newBlock("for_end")

	previousEnd, previousNext := g.loopEndBlock, g.loopNextBlock
	g.loopEndBlock = endBlock
	g.loopNextBlock = postBlock
	defer func() {
		g.loopEndBlock, g.loopNextBlock = previousEnd, previousNext
	}()

	if node.Init != nil {
		if err := g.generateStatement(node.Init); err != nil {
			return err
		}
	}
	if g.currentBlockTerminated() {
		return nil
	}
	g.builder.CreateBr(conditionBlock)

	g.builder.SetInsertPointAtEnd(conditionBlock)
	if node.Condition == nil {
		g.builder.CreateBr(bodyBlock)
	} else {
		condition, err := g.generateCondition(*node.Condition)
		if err != nil {
			return err
		}
		g.builder.CreateCondBr(condition, bodyBlock, endBlock)
	}

	g.builder.SetInsertPointAtEnd(bodyBlock)
	if err := g.generateBlock(&node.Body); err != nil {
		return err
	}
	if !g.currentBlockTerminated() {
		g.builder.CreateBr(postBlock)
	}

	g.builder.SetInsertPointAtEnd(postBlock)
	if node.Post != nil {
		if err := g.generateStatement(node.Post); err != nil {
			return err
		}
	}
	if !g.currentBlockTerminated() {
		g.builder.CreateBr(conditionBlock)
	}

	g.builder.SetInsertPointAtEnd(endBlock)
	return nil
}

func (g *Generator) generateBreak(node *ast.NodeBreak) error {
	if g.loopEndBlock.IsNil() {
		return g.error("break outside of loop", node.Pos)
	}
	g.builder.CreateBr(g.loopEndBlock)
	return nil
}

func (g *Generator) generateContinue(node *ast.NodeContinue) error {
	if g.loopNextBlock.IsNil() {
		return g.error("continue outside of loop", node.Pos)
	}
	g.builder.CreateBr(g.loopNextBlock)
	return nil
}
