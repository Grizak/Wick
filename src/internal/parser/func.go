package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseFuncDecl() (*ast.NodeFuncDecl, error) {
	if p.peek(0).Type == types.TokenFunctionDecl {
		p.consume()
	}
	p.expect(types.TokenOpenParen)

	var params []ast.NodeParam

	if p.peek(0).Type != types.TokenCloseParen {
		for {
			decl, err := p.parseParam()
			if err != nil {
				return nil, err
			}
			params = append(params, *decl)

			if ok, _ := p.tryConsume(types.TokenComma); ok {
				continue
			}
			break
		}
		if _, err := p.expect(types.TokenCloseParen); err != nil {
			return nil, err
		}
	} else {
		if _, err := p.expect(types.TokenCloseParen); err != nil {
			return nil, err
		}
	}

	// Optional return type: e.g. `: i64`
	var returnType *string
	if p.peek(0).Type == types.TokenColon {
		p.consume()
		typeTok, err := p.expect(types.TokenIdent)
		if err != nil {
			return nil, err
		}
		name := typeTok.Value
		returnType = name
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.NodeFuncDecl{
		Body:       *body,
		Params:     params,
		ReturnType: returnType,
		Pos:        body.Pos,
	}, nil
}

func (p *Parser) parseReturnStmt() (*ast.NodeReturn, error) {
	p.expect(types.TokenReturnStmt)

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.NodeReturn{Expr: &expr, Pos: expr.Pos}, nil
}

func (p *Parser) parseParam() (*ast.NodeParam, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}

	var typeName *string
	if p.peek(0).Type == types.TokenColon {
		if _, err := p.consume(); err != nil {
			return nil, err
		}
		typeStr, err := p.parseType()
		if err != nil {
			return nil, err
		}
		typeName = &typeStr
	}

	var defaultExpr *ast.NodeExpression
	if p.peek(0).Type == types.TokenEquals {
		if _, err := p.consume(); err != nil {
			return nil, err
		}
		expr, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		defaultExpr = &expr
	}

	return &ast.NodeParam{
		Name:    *name.Value,
		Type:    typeName,
		Default: defaultExpr,
		Pos:     name.Pos,
	}, nil
}
