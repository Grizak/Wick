package parser

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseComparison() (ast.NodeExpression, error) {
	token, err := p.peek(0)
	if err != nil {
		return ast.NodeExpression{}, err
	}
	if token.Type == types.TokenFunctionDecl {
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

	token, err = p.peek(0)
	if err != nil {
		return ast.NodeExpression{}, err
	}

	var op types.BinOp
	switch token.Type {
	case types.TokenEqEq:
		op = types.BinOpEq
	case types.TokenNotEq:
		op = types.BinOpNotEq
	case types.TokenLt:
		op = types.BinOpLt
	case types.TokenGt:
		op = types.BinOpGt
	case types.TokenLtEq:
		op = types.BinOpLtEq
	case types.TokenGtEq:
		op = types.BinOpGtEq
	default:
		return left, nil
	}

	if _, err := p.consume(); err != nil {
		return ast.NodeExpression{}, err
	}

	right, err := p.parseExpression()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	return ast.NodeExpression{
		BinExpr: &ast.NodeBinExpr{
			Left:  left,
			Op:    op,
			Right: right,
			Pos:   left.Pos,
		},
	}, nil
}

func (p *Parser) parseExpression() (ast.NodeExpression, error) {
	left, err := p.parseTerm()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	for {
		token, err := p.peek(0)
		if err != nil {
			return ast.NodeExpression{}, err
		}

		var op types.BinOp
		switch token.Type {
		case types.TokenPlus:
			op = types.BinOpAdd
		case types.TokenMinus:
			op = types.BinOpSub
		default:
			return left, nil
		}

		if _, err := p.consume(); err != nil {
			return ast.NodeExpression{}, err
		}

		right, err := p.parseTerm()
		if err != nil {
			return ast.NodeExpression{}, err
		}

		left = ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  left,
				Op:    op,
				Right: right,
				Pos:   left.Pos,
			},
		}
	}
}

func (p *Parser) parseTerm() (ast.NodeExpression, error) {
	left, err := p.parseFactor()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	for {
		token, err := p.peek(0)
		if err != nil {
			return ast.NodeExpression{}, err
		}

		var op types.BinOp
		switch token.Type {
		case types.TokenStar:
			op = types.BinOpMul
		case types.TokenFSlash:
			op = types.BinOpDiv
		default:
			return left, nil
		}

		if _, err := p.consume(); err != nil {
			return ast.NodeExpression{}, err
		}

		right, err := p.parseFactor()
		if err != nil {
			return ast.NodeExpression{}, err
		}

		left = ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  left,
				Op:    op,
				Right: right,
				Pos:   left.Pos,
			},
		}
	}
}

func (p *Parser) parseFactor() (ast.NodeExpression, error) {
	token, err := p.peek(0)
	if err != nil {
		return ast.NodeExpression{}, err
	}

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
		module := ""
		name := *token.Value
		if token, err := p.peek(0); err != nil {
			return ast.NodeExpression{}, err
		} else if token.Type == types.TokenDot {
			_, err := p.consume()
			if err != nil {
				return ast.NodeExpression{}, err
			}
			member, err := p.expect(types.TokenIdent)
			if err != nil {
				return ast.NodeExpression{}, err
			}
			module = name
			name = *member.Value
		}
		// Check if this is a function call
		if token, err := p.peek(0); err != nil {
			return ast.NodeExpression{}, err
		} else if token.Type == types.TokenOpenParen {
			_, err := p.consume() // consume (
			if err != nil {
				return ast.NodeExpression{}, err
			}
			var args []ast.NodeExpression
			if token, err := p.peek(0); err != nil {
				return ast.NodeExpression{}, err
			} else if token.Type != types.TokenCloseParen {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return ast.NodeExpression{}, err
					}
					args = append(args, arg)
					if token, err := p.peek(0); err != nil {
						return ast.NodeExpression{}, err
					} else if token.Type != types.TokenComma {
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
					Module: module,
					Name:   name,
					Args:   args,
					Pos:    token.Pos,
				},
			}, nil
		}
		if module != "" {
			return ast.NodeExpression{}, p.error("expected function call after module namespace", token)
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
