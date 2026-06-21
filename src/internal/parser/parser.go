package parser

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

type Parser struct {
	input    chan types.LexerResult
	buffer   []types.LexerResult
	filename string
}

func NewParser(filename string) *Parser {
	return &Parser{
		filename: filename,
	}
}

// Read from input, block when input is empty
func (p *Parser) Parse(input chan types.LexerResult) (ast.NodeProgram, error) {
	p.input = input
	var program ast.NodeProgram
	stmts, err := p.parseStatements(types.TokenEOF)
	if err != nil {
		return program, err
	}
	program.Statements = stmts
	return program, nil
}

func (p *Parser) parseExit() (*ast.NodeExit, error) {
	_, err := p.expect(types.TokenOpenParen)
	if err != nil {
		return &ast.NodeExit{}, err
	}
	expr, err := p.parseComparison()
	if err != nil {
		return &ast.NodeExit{}, err
	}
	_, err = p.expect(types.TokenCloseParen)
	if err != nil {
		return &ast.NodeExit{}, err
	}

	return &ast.NodeExit{Expr: expr, Pos: expr.Pos}, nil
}

func (p *Parser) error(msg string, token types.Token) *types.CompileError {
	return &types.CompileError{
		File: p.filename,
		Pos:  &token.Pos,
		Msg:  msg,
	}
}

func (p *Parser) expect(tokenType types.TokenType) (types.Token, error) {
	token, err := p.consume()
	if err != nil {
		return types.Token{}, err
	}
	if token.Type != tokenType {
		return token, p.error(fmt.Sprintf("expected `%s` but got `%s`", tokenType, token.Type), token)
	}
	return token, nil
}

func (p *Parser) peek(offset int) types.Token {
	// Fill buffer up to offset+1
	for len(p.buffer) <= offset {
		p.buffer = append(p.buffer, <-p.input)
	}
	return p.buffer[offset].Token
}

func (p *Parser) consume() (types.Token, error) {
	// Make sure buffer has at least one token
	if len(p.buffer) == 0 {
		result := <-p.input
		if result.Err != nil {
			return types.Token{}, result.Err
		}
		return result.Token, nil
	}

	token := p.buffer[0]
	if token.Err != nil {
		return types.Token{}, token.Err
	}
	p.buffer = p.buffer[1:]
	return token.Token, nil
}

func (p *Parser) tryConsume(tokenType types.TokenType) (bool, error) {
	if len(p.buffer) == 0 {
		result := <-p.input
		if result.Err != nil {
			return false, result.Err
		}
		if result.Token.Type == tokenType {
			return true, nil
		}
		return false, nil
	}

	token := p.buffer[0]
	if token.Err != nil {
		return false, token.Err
	}

	if token.Token.Type == tokenType {
		p.buffer = p.buffer[1:]
		return true, nil
	}

	return false, nil
}
