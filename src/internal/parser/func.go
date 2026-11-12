package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseFuncDecl() (*ast.NodeFuncDecl, error) {
	start, err := p.peek(0)
	if err != nil {
		return nil, err
	}

	if start.Type == types.TokenFunctionDecl {
		if _, err := p.consume(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(types.TokenOpenParen); err != nil {
		return nil, err
	}

	var params []ast.NodeParam

	token, err := p.peek(0)
	if err != nil {
		return nil, err
	}
	if token.Type != types.TokenCloseParen {
		for {
			decl, err := p.parseParam()
			if err != nil {
				return nil, err
			}
			params = append(params, *decl)

			hasComma, err := p.tryConsume(types.TokenComma)
			if err != nil {
				return nil, err
			}
			if !hasComma {
				break
			}
		}
	}

	if _, err := p.expect(types.TokenCloseParen); err != nil {
		return nil, err
	}

	var returnType *string
	token, err = p.peek(0)
	if err != nil {
		return nil, err
	}
	if token.Type == types.TokenColon {
		if _, err := p.consume(); err != nil {
			return nil, err
		}

		typeName, err := p.parseType()
		if err != nil {
			return nil, err
		}
		returnType = &typeName
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.NodeFuncDecl{
		Body:       *body,
		Params:     params,
		ReturnType: returnType,
		Pos:        start.Pos,
	}, nil
}

func (p *Parser) parseReturnStmt() (*ast.NodeReturn, error) {
	token, err := p.expect(types.TokenReturnStmt)
	if err != nil {
		return nil, err
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.NodeReturn{Expr: &expr, Pos: token.Pos}, nil
}

func (p *Parser) parseParam() (*ast.NodeParam, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}

	var typeName *string
	if token, err := p.peek(0); err != nil {
		return nil, err
	} else if token.Type == types.TokenColon {
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
	if token, err := p.peek(0); err != nil {
		return nil, err
	} else if token.Type == types.TokenEquals {
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
