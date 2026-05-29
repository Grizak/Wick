package parser

import (
	"strconv"

	"github.com/Grizak/Wick/src/types"
)

type Parser struct {
	input  chan types.Token
	buffer []types.Token
}

func NewParser() *Parser {
	return &Parser{}
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
	var exit types.NodeExit

	token := p.consume()
	// Expect an `OpenParen`, then `int_lit`, then `CloseParen`
	if token.Type != types.TokenOpenParen {
		panic("expected `(`")
	}
	exit.Expr = p.parseExpression()

	token = p.consume()
	if token.Type != types.TokenCloseParen {
		panic("expected `)`")
	}

	return &exit
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
	token := p.consume()
	if token.Type != types.TokenIntLit {
		panic("expected int literal")
	}
	i, err := strconv.Atoi(*token.Value)
	if err != nil {
		panic("invalid int literal")
	}
	return types.NodeExpression{IntLit: &i}
}
