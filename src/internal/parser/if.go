package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseIf() (*ast.NodeIf, error) {
	pos := p.peek(0).Pos

	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBlock *ast.NodeBlock
	if p.peek(0).Type == types.TokenElse {
		p.consume()
		// else if
		if p.peek(0).Type == types.TokenIf {
			p.consume()
			elseIf, err := p.parseIf()
			if err != nil {
				return nil, err
			}
			// wrap the else if in a block
			elseBlock = &ast.NodeBlock{
				Statements: []ast.NodeStatement{{If: elseIf}},
				Pos:        elseIf.Pos,
			}
		} else {
			elseBlock, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
		}
	}

	return &ast.NodeIf{
		Condition: condition,
		Then:      *then,
		Else:      elseBlock,
		Pos:       pos,
	}, nil
}
