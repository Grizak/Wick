package codegen

import "github.com/Grizak/Wick/src/internal/ast"

func (g *Generator) generateIf(node *ast.NodeIf) error {
	condition, err := g.generateCondition(node.Condition)
	if err != nil {
		return err
	}

	thenBlock := g.newBlock("if_then")
	mergeBlock := g.newBlock("if_end")
	elseBlock := mergeBlock
	if node.Else != nil {
		elseBlock = g.newBlock("if_else")
	}
	g.builder.CreateCondBr(condition, thenBlock, elseBlock)

	g.builder.SetInsertPointAtEnd(thenBlock)
	if err := g.generateBlock(&node.Then); err != nil {
		return err
	}
	if !g.currentBlockTerminated() {
		g.builder.CreateBr(mergeBlock)
	}

	if node.Else != nil {
		g.builder.SetInsertPointAtEnd(elseBlock)
		if err := g.generateBlock(node.Else); err != nil {
			return err
		}
		if !g.currentBlockTerminated() {
			g.builder.CreateBr(mergeBlock)
		}
	}

	g.builder.SetInsertPointAtEnd(mergeBlock)
	return nil
}
