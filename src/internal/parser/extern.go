package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseExternStmt() (*ast.NodeExtern, error) {
	var externNode ast.NodeExtern
	_, err := p.expect(types.TokenOpenParen)
	if err != nil {
		return nil, err
	}
	token, err := p.expect(types.TokenStringLit)
	if err != nil {
		return nil, err
	}
	if *token.Value != "c" {
		return nil, p.error("only 'c' is supported for extern statements", token)
	}

	token, err = p.peek(0)
	if err != nil {
		return nil, err
	}
	if token.Type == types.TokenComma {
		_, err = p.consume()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(types.TokenLibs)
		if err != nil {
			return nil, err
		}
		_, err = p.expect(types.TokenEquals)
		if err != nil {
			return nil, err
		}
		token, err = p.expect(types.TokenStringLit)
		if err != nil {
			return nil, err
		}
		p.AddLibs(*token.Value)
	}

	return &externNode, nil
}
