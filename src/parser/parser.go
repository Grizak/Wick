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
func (p *Parser) Parse(input chan types.Token) types.NodeProgram {
	var program types.NodeProgram
	p.input = input

	for {
		token := p.consume()

		switch token.Type {
		case types.TokenExit:
			program.Statements = append(program.Statements, types.NodeStatement{Exit: p.parseExit()})
		case types.TokenOpenParen:
			panic("unexpected `(`")
		case types.TokenCloseParen:
			panic("unexpected `)`")
		case types.TokenIntLit:
			panic("unexpected int literal")
		case types.TokenEOF:
			return program
		default:
			panic("unexpected token")
		}
	}
}

func (p *Parser) parseExit() *types.NodeExit {
	p.expect(types.TokenOpenParen)
	expr := p.parseExpression()
	p.expect(types.TokenCloseParen)

	return &types.NodeExit{Expr: expr}
}

func (p *Parser) parseExpression() types.NodeExpression {
	term := p.parseTerm()

	if p.peek(0).Type == types.TokenPlus {
		p.consume()
		right := p.parseExpression()
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpAdd,
				Right: right,
			},
		}
	}

	return term
}

func (p *Parser) parseTerm() types.NodeExpression {
	factor := p.parseFactor()

	if p.peek(0).Type == types.TokenStar {
		p.consume()
		right := p.parseTerm()
		return types.NodeExpression{
			BinExpr: &types.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpMul,
				Right: right,
			},
		}
	}
	return factor
}

func (p *Parser) parseFactor() types.NodeExpression {
	token := p.expect(types.TokenIntLit)
	i, err := strconv.Atoi(*token.Value)
	if err != nil {
		p.panic("invalid int literal", token)
	}
	return types.NodeExpression{IntLit: &i}
}

func (p *Parser) panic(err string, token types.Token) {
	panic(fmt.Sprintf("%s:%d:%d: %s", p.filename, token.Pos.Line, token.Pos.Column, err))
}

func (p *Parser) expect(tokenType types.TokenType) types.Token {
	token := p.consume()
	if token.Type != tokenType {
		p.panic(fmt.Sprintf("expected `%s` but got `%s`", tokenType, token.Type), token)
	}
	return token
}
