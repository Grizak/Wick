package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseIf(pos types.Position) (*ast.NodeIf, error) {
	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBlock *ast.NodeBlock
	if token, err := p.peek(0); err != nil {
		return nil, err
	} else if token.Type == types.TokenElse {
		if _, err := p.consume(); err != nil {
			return nil, err
		}
		// else if
		if token, err := p.peek(0); err != nil {
			return nil, err
		} else if token.Type == types.TokenIf {
			ifTok, err := p.consume()
			if err != nil {
				return nil, err
			}
			elseIf, err := p.parseIf(ifTok.Pos)
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
