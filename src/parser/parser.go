package parser

import (
	"strconv"

	"github.com/Grizak/Wick/src/types"
)

type Parser struct{}

func NewParser() *Parser {
	p := Parser{}

	return &p
}

// Read from input, block when input is empty
func (p *Parser) Parse(input chan types.Token) types.NodeProgram {
	var program types.NodeProgram

	for {
		token := <-input

		switch token.Type {
		case types.TokenExit:
			program.Statements = append(program.Statements, types.NodeStatement{Exit: p.parseExit(input)})
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

func (p *Parser) parseExit(input chan types.Token) *types.NodeExit {
	var exit types.NodeExit

	token := <-input
	// Expect an `OpenParen`, then `int_lit`, then `CloseParen`
	if token.Type != types.TokenOpenParen {
		panic("expected `(`")
	}

	token = <-input
	if token.Type != types.TokenIntLit {
		panic("expected int literal")
	}

	i, err := strconv.Atoi(*token.Value)
	if err != nil {
		panic("invalid int literal")
	}
	exit.Expr.IntLit = i

	token = <-input
	if token.Type != types.TokenCloseParen {
		panic("expected `)`")
	}

	return &exit
}
