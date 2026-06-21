package parser

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseComparison() (ast.NodeExpression, error) {
	if p.peek(0).Type == types.TokenFunctionDecl || p.peek(0).Type == types.TokenOpenParen {
		funcDecl, err := p.parseFuncDecl()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			FuncDecl: funcDecl,
			Pos:      funcDecl.Pos,
		}, nil
	}

	left, err := p.parseExpression()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenEqEq:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpEq, Right: right}}, nil
	case types.TokenNotEq:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpNotEq, Right: right}}, nil
	case types.TokenLt:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpLt, Right: right}}, nil
	case types.TokenGt:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpGt, Right: right}}, nil
	case types.TokenLtEq:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpLtEq, Right: right}}, nil
	case types.TokenGtEq:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{BinExpr: &ast.NodeBinExpr{Left: left, Op: types.BinOpGtEq, Right: right}}, nil
	}

	return left, nil
}

func (p *Parser) parseExpression() (ast.NodeExpression, error) {
	term, err := p.parseTerm()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenPlus:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpAdd,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	case types.TokenMinus:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpSub,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	}

	return term, nil
}

func (p *Parser) parseTerm() (ast.NodeExpression, error) {
	factor, err := p.parseFactor()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenStar:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpMul,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	case types.TokenFSlash:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpDiv,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	}
	return factor, nil
}

func (p *Parser) parseFactor() (ast.NodeExpression, error) {
	token := p.peek(0)

	if token.Type == types.TokenOpenParen {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		expr, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		_, err = p.expect(types.TokenCloseParen)
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return expr, nil
	}

	if token.Type == types.TokenIdent {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		// Check if this is a function call
		if p.peek(0).Type == types.TokenOpenParen {
			_, err := p.consume() // consume (
			if err != nil {
				return ast.NodeExpression{}, err
			}
			var args []ast.NodeExpression
			if p.peek(0).Type != types.TokenCloseParen {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return ast.NodeExpression{}, err
					}
					args = append(args, arg)
					if p.peek(0).Type != types.TokenComma {
						break
					}
					_, err = p.consume() // consume comma
					if err != nil {
						return ast.NodeExpression{}, err
					}
				}
			}
			_, err = p.expect(types.TokenCloseParen)
			if err != nil {
				return ast.NodeExpression{}, err
			}
			return ast.NodeExpression{
				FuncCall: &ast.NodeFuncCall{
					Name: *token.Value,
					Args: args,
					Pos:  token.Pos,
				},
			}, nil
		}
		return ast.NodeExpression{Ident: token.Value, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenIntLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		i, err := strconv.Atoi(*token.Value)
		if err != nil {
			return ast.NodeExpression{}, p.error("invalid int literal", token)
		}
		return ast.NodeExpression{IntLit: &i, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenFloatLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		f, err := strconv.ParseFloat(*token.Value, 64)
		if err != nil {
			return ast.NodeExpression{}, p.error("invalid float literal", token)
		}
		return ast.NodeExpression{FloatLit: &f, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenStringLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{StringLit: token.Value, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenTrue {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		t := true
		return ast.NodeExpression{BoolLit: &t, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenFalse {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		f := false
		return ast.NodeExpression{BoolLit: &f, Pos: token.Pos}, nil
	}

	return ast.NodeExpression{}, p.error(fmt.Sprintf("expected literal or identifier, got %s", token.Type), token)
}
