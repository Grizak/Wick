package main

import "strconv"

type NodeProgram struct {
	statements []NodeStatement
}

type NodeStatement struct {
	Exit *NodeExit
}

type NodeExit struct {
	Expr NodeExpression
}

type NodeExpression struct {
	IntLit int
}

type Parser struct{}

func NewParser() *Parser {
	p := Parser{}

	return &p
}

// Read from input, block when input is empty
func (p *Parser) Parse(input chan Token) NodeProgram {
	var program NodeProgram

	for {
		token := <-input

		switch token._type {
		case TokenExit:
			program.statements = append(program.statements, NodeStatement{Exit: p.parseExit(input)})
		case TokenOpenParen:
			panic("unexpected `(`")
		case TokenCloseParen:
			panic("unexpected `)`")
		case TokenIntLit:
			panic("unexpected int literal")
		case TokenEOF:
			return program
		default:
			panic("unexpected token")
		}
	}
}

func (p *Parser) parseExit(input chan Token) *NodeExit {
	var exit NodeExit

	token := <-input
	// Expect an `OpenParen`, then `int_lit`, then `CloseParen`
	if token._type != TokenOpenParen {
		panic("expected `(`")
	}

	token = <-input
	if token._type != TokenIntLit {
		panic("expected int literal")
	}

	i, err := strconv.Atoi(*token.value)
	if err != nil {
		panic("invalid int literal")
	}
	exit.Expr.IntLit = i

	token = <-input
	if token._type != TokenCloseParen {
		panic("expected `)`")
	}

	return &exit
}
