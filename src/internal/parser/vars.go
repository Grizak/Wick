package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseVarDecl(isConst bool) (*ast.NodeVarDecl, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}

	var typeName *string
	if p.peek(0).Type == types.TokenColon {
		_, err := p.consume()
		if err != nil {
			return nil, err
		}
		typeStr, err := p.parseType()
		if err != nil {
			return nil, err
		}
		typeName = &typeStr
	}

	_, err = p.expect(types.TokenEquals)
	if err != nil {
		return nil, err
	}

	expr, err := p.parseComparison()
	if err != nil {
		return &ast.NodeVarDecl{}, err
	}

	return &ast.NodeVarDecl{
		Name:  *name.Value,
		Type:  typeName,
		Expr:  expr,
		Const: isConst,
		Pos:   expr.Pos,
	}, nil
}

func (p *Parser) parseVarAssign() (*ast.NodeVarAssign, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}
	_, err = p.expect(types.TokenEquals)
	if err != nil {
		return nil, err
	}
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	return &ast.NodeVarAssign{
		Name: *name.Value,
		Expr: expr,
		Pos:  expr.Pos,
	}, nil
}
