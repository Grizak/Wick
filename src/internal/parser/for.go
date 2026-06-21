package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseFor() (*ast.NodeFor, error) {
	pos := p.peek(0).Pos

	// Infinite loop: for {
	if p.peek(0).Type == types.TokenOpenBrace {
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &ast.NodeFor{
			Body: *body,
			Pos:  pos,
		}, nil
	}

	// Peek ahead to see if this is a C-style for
	// by checking if the first statement is followed by a semicolon
	// We need to try parsing an init statement and see if ; follows
	isCStyle := p.isCStyleFor()

	if isCStyle {
		return p.parseCStyleFor(pos)
	}

	// While-style: for <condition> { }
	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.NodeFor{
		Condition: &condition,
		Body:      *body,
		Pos:       pos,
	}, nil
}

func (p *Parser) isCStyleFor() bool {
	i := 0
	depth := 0
	for {
		token := p.peek(i)
		switch token.Type {
		case types.TokenOpenParen:
			depth++
		case types.TokenCloseParen:
			depth--
		case types.TokenSemicolon:
			if depth == 0 {
				return true
			}
		case types.TokenOpenBrace, types.TokenEOF:
			return false
		}
		i++
	}
}

func (p *Parser) parseCStyleFor(pos types.Position) (*ast.NodeFor, error) {
	// Init statement
	init, err := p.parseForStatement()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(types.TokenSemicolon); err != nil {
		return nil, err
	}

	// Condition
	condition, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(types.TokenSemicolon); err != nil {
		return nil, err
	}

	// Post statement
	post, err := p.parseForStatement()
	if err != nil {
		return nil, err
	}

	// Body
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.NodeFor{
		Init:      init,
		Condition: &condition,
		Post:      post,
		Body:      *body,
		Pos:       pos,
	}, nil
}

func (p *Parser) parseForStatement() (*ast.NodeStatement, error) {
	token := p.peek(0)

	switch token.Type {
	case types.TokenLet:
		p.consume()
		decl, err := p.parseVarDecl(false)
		if err != nil {
			return nil, err
		}
		return &ast.NodeStatement{VarDecl: decl}, nil
	case types.TokenIdent:
		if p.peek(1).Type == types.TokenEquals {
			assign, err := p.parseVarAssign()
			if err != nil {
				return nil, err
			}
			return &ast.NodeStatement{VarAssign: assign}, nil
		}
		return nil, p.error("expected `=` after identifier", token)
	default:
		return nil, p.error("expected variable declaration or assignment", token)
	}
}
