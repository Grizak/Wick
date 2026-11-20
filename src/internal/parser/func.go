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

	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}

	var returnType *string
	token, err := p.peek(0)
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

func (p *Parser) parseParams() ([]ast.NodeParam, error) {
	var params []ast.NodeParam

	if _, err := p.expect(types.TokenOpenParen); err != nil {
		return nil, err
	}

	token, err := p.peek(0)
	if err != nil {
		return nil, err
	}
	if token.Type != types.TokenCloseParen {
		for {
			param, err := p.parseParam()
			if err != nil {
				return nil, err
			}
			params = append(params, *param)

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

	return params, nil
}

func (p *Parser) parseFuncCall() (ast.NodeFuncCall, error) {
	token, err := p.consume()
	if err != nil {
		return ast.NodeFuncCall{}, err
	}
	module := ""
	name := *token.Value
	if token, err := p.peek(0); err != nil {
		return ast.NodeFuncCall{}, err
	} else if token.Type == types.TokenDot {
		_, err := p.consume()
		if err != nil {
			return ast.NodeFuncCall{}, err
		}
		member, err := p.expect(types.TokenIdent)
		if err != nil {
			return ast.NodeFuncCall{}, err
		}
		module = name
		name = *member.Value
	}
	_, err = p.consume() // consume (
	if err != nil {
		return ast.NodeFuncCall{}, err
	}
	var args []ast.NodeExpression
	if token, err := p.peek(0); err != nil {
		return ast.NodeFuncCall{}, err
	} else if token.Type != types.TokenCloseParen {
		for {
			arg, err := p.parseExpression()
			if err != nil {
				return ast.NodeFuncCall{}, err
			}
			args = append(args, arg)
			if token, err := p.peek(0); err != nil {
				return ast.NodeFuncCall{}, err
			} else if token.Type != types.TokenComma {
				break
			}
			_, err = p.consume() // consume comma
			if err != nil {
				return ast.NodeFuncCall{}, err
			}
		}
	}
	_, err = p.expect(types.TokenCloseParen)
	if err != nil {
		return ast.NodeFuncCall{}, err
	}
	return ast.NodeFuncCall{
		Module: module,
		Name:   name,
		Args:   args,
		Pos:    token.Pos,
	}, nil
}
