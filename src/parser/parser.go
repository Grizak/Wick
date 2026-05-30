package parser

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/types"
)

type Parser struct {
	input    chan types.Token
	buffer   []types.Token
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{
		filename: filename,
	}
}

func (p *Parser) peek(offset int) types.Token {
	// Fill buffer up to offset+1
	for len(p.buffer) <= offset {
		p.buffer = append(p.buffer, <-p.input)
	}
	return p.buffer[offset]
}

func (p *Parser) consume() types.Token {
	// Make sure buffer has at least one token
	if len(p.buffer) == 0 {
		return <-p.input
	}

	token := p.buffer[0]
	p.buffer = p.buffer[1:]
	return token
}

// Read from input, block when input is empty
func (p *Parser) Parse(input chan types.Token) (types.NodeProgram, error) {
	var program types.NodeProgram
	p.input = input

	for {
		token := p.peek(0)

		switch token.Type {
		case types.TokenExit:
			p.consume()
			exit, err := p.parseExit()
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, types.NodeStatement{Exit: exit})
		case types.TokenConst:
			p.consume()
			varDecl, err := p.parseVarDecl(true)
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, types.NodeStatement{VarDecl: varDecl})
		case types.TokenIdent:
			// Peek ahead to disambiguate
			next := p.peek(1)
			if next.Type == types.TokenEquals {
				varAssign, err := p.parseVarAssign()
				if err != nil {
					return program, err
				}
				program.Statements = append(program.Statements, types.NodeStatement{VarAssign: varAssign})
			} else {
				p.error("expected `=` after identifier", token)
			}
		case types.TokenLet:
			p.consume()
			varDecl, err := p.parseVarDecl(false)
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, types.NodeStatement{VarDecl: varDecl})
		case types.TokenOpenParen:
			p.error("unexpected `(`", token)
		case types.TokenCloseParen:
			p.error("unexpected `)`", token)
		case types.TokenIntLit:
			p.error("unexpected int literal", token)
		case types.TokenEOF:
			return program, nil
		default:
			p.error("unexpected token", token)
		}
	}
}

func (p *Parser) parseExit() (*types.NodeExit, error) {
	p.expect(types.TokenOpenParen)
	expr, err := p.parseExpression()
	if err != nil {
		return &types.NodeExit{}, err
	}
	p.expect(types.TokenCloseParen)

	return &types.NodeExit{Expr: expr, Pos: expr.Pos}, nil
}

func (p *Parser) parseExpression() (types.NodeExpression, error) {
	term, err := p.parseTerm()
	if err != nil {
		return types.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenPlus:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return types.NodeExpression{}, err
		}
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpAdd,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	case types.TokenMinus:
		p.consume()
		right, err := p.parseExpression()
		if err != nil {
			return types.NodeExpression{}, err
		}
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpSub,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	}

	return term, nil
}

func (p *Parser) parseTerm() (types.NodeExpression, error) {
	factor, err := p.parseFactor()
	if err != nil {
		return types.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenStar:
		p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return types.NodeExpression{}, err
		}
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpMul,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	case types.TokenFSlash:
		p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return types.NodeExpression{}, err
		}
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpDiv,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	}
	return factor, nil
}

func (p *Parser) parseFactor() (types.NodeExpression, error) {
	token := p.peek(0)

	if token.Type == types.TokenOpenParen {
		p.consume()
		expr, err := p.parseExpression()
		if err != nil {
			return types.NodeExpression{}, err
		}
		p.expect(types.TokenCloseParen)
		return expr, nil
	}

	if token.Type == types.TokenIdent {
		p.consume()
		return types.NodeExpression{Ident: token.Value, Pos: token.Pos}, nil
	}

	token, err := p.expect(types.TokenIntLit)
	if err != nil {
		return types.NodeExpression{}, err
	}
	i, err := strconv.Atoi(*token.Value)
	if err != nil {
		p.error("invalid int literal", token)
	}
	return types.NodeExpression{IntLit: &i, Pos: token.Pos}, nil
}

func (p *Parser) error(msg string, token types.Token) *types.CompileError {
	return &types.CompileError{
		File: p.filename,
		Pos:  token.Pos,
		Msg:  msg,
	}
}

func (p *Parser) expect(tokenType types.TokenType) (types.Token, error) {
	token := p.consume()
	if token.Type != tokenType {
		return token, p.error(fmt.Sprintf("expected `%s` but got `%s`", tokenType, token.Type), token)
	}
	return token, nil
}

func (p *Parser) parseVarDecl(isConst bool) (*types.NodeVarDecl, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}

	var typeName *string
	if p.peek(0).Type == types.TokenColon {
		p.consume()
		typeToken, err := p.expect(types.TokenIdent)
		if err != nil {
			return nil, err
		}
		typeName = typeToken.Value
	}

	p.expect(types.TokenEquals)
	expr, err := p.parseExpression()
	if err != nil {
		return &types.NodeVarDecl{}, err
	}

	return &types.NodeVarDecl{
		Name:  *name.Value,
		Type:  typeName,
		Expr:  expr,
		Const: isConst,
		Pos:   expr.Pos,
	}, nil
}

func (p *Parser) parseVarAssign() (*types.NodeVarAssign, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return &types.NodeVarAssign{}, err
	}
	p.expect(types.TokenEquals)
	expr, err := p.parseExpression()
	if err != nil {
		return &types.NodeVarAssign{}, err
	}

	return &types.NodeVarAssign{
		Name: *name.Value,
		Expr: expr,
		Pos:  expr.Pos,
	}, nil
}
