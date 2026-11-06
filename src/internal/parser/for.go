package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseFor(pos types.Position) (*ast.NodeFor, error) {
	// Infinite loop: for {
	if token, err := p.peek(0); err != nil {
		return nil, err
	} else if token.Type == types.TokenOpenBrace {
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
	isCStyle, err := p.isCStyleFor()
	if err != nil {
		return nil, err
	}

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

func (p *Parser) isCStyleFor() (bool, error) {
	i := 0
	depth := 0
	for {
		token, err := p.peek(i)
		if err != nil {
			return false, err
		}
		switch token.Type {
		case types.TokenOpenParen:
			depth++
		case types.TokenCloseParen:
			depth--
		case types.TokenSemicolon:
			if depth == 0 {
				return true, nil
			}
		case types.TokenOpenBrace, types.TokenEOF:
			return false, nil
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
	token, err := p.peek(0)
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case types.TokenLet:
		if _, err := p.consume(); err != nil {
			return nil, err
		}
		decl, err := p.parseVarDecl(false)
		if err != nil {
			return nil, err
		}
		return &ast.NodeStatement{VarDecl: decl}, nil
	case types.TokenIdent:
		if token, err := p.peek(1); err != nil {
			return nil, err
		} else if token.Type == types.TokenEquals {
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
