package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseIdent(token types.Token) (*ast.NodeStatement, error) {
	next, err := p.peek(1)
	if err != nil {
		return nil, err
	}
	switch next.Type {
	case types.TokenEquals:
		assign, err := p.parseVarAssign()
		if err != nil {
			return nil, err
		}
		return &ast.NodeStatement{VarAssign: assign}, nil
	case types.TokenOpenParen:
		call, err := p.parseFuncCall()
		if err != nil {
			return nil, err
		}
		return &ast.NodeStatement{FuncCall: &call}, nil
	case types.TokenDot:
		if _, err := p.consume(); err != nil { // Consume the `.`
			return nil, err
		}
		return p.parseIdent(token)
	default:
		return nil, p.error("expected `=` or `(` after identifier", token)
	}
}
