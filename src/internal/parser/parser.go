package parser

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
	"github.com/Grizak/Wick/src/internal/utils"
)

type Parser struct {
	input    chan types.LexerResult
	buffer   []types.LexerResult
	Libs     []string
	libsSeen map[string]bool
	sawEOF   bool
}

func NewParser() *Parser {
	return &Parser{
		libsSeen: make(map[string]bool),
	}
}

// Read from input, block when input is empty
func (p *Parser) Parse(input chan types.LexerResult, dir string) (ast.NodeProgram, error) {
	p.input = input
	var program ast.NodeProgram
	stmts, err := p.parseStatements(types.TokenEOF)
	if err != nil {
		return program, err
	}
	program.Statements = stmts
	program.Dir = dir
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
		File: token.Pos.Filename,
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

func (p *Parser) peek(offset int) (types.Token, error) {
	utils.Debug(fmt.Sprintf("[peek offset=%d buffer=%d sawEOF=%v]", offset, len(p.buffer), p.sawEOF))
	if p.sawEOF {
		return types.Token{Type: types.TokenEOF}, nil
	}

	if offset < 0 {
		return types.Token{}, fmt.Errorf("negative lookahead offset: %d", offset)
	}

	for len(p.buffer) <= offset {
		result, ok := <-p.input
		if !ok {
			// The lexer should emit EOF before closing the channel.
			return types.Token{}, fmt.Errorf("lexer input closed before EOF (peek)")
		}
		if result.Err != nil {
			return types.Token{}, result.Err
		}

		p.buffer = append(p.buffer, result)

		utils.Debug(fmt.Sprintf("[peek got %s ok=%v]", result.Token.Type, ok))

		if result.Token.Type == types.TokenEOF {
			p.sawEOF = true
			break
		}
	}

	return p.buffer[offset].Token, nil
}

func (p *Parser) consume() (types.Token, error) {
	utils.Debug(fmt.Sprintf("[consume buffer=%d sawEOF=%v]", len(p.buffer), p.sawEOF))
	if p.sawEOF {
		return types.Token{Type: types.TokenEOF}, nil
	}

	if len(p.buffer) == 0 {
		result, ok := <-p.input
		if !ok {
			return types.Token{}, fmt.Errorf("lexer input closed before EOF (consume)")
		}
		if result.Err != nil {
			return types.Token{}, result.Err
		}
		return result.Token, nil
	}

	result := p.buffer[0]
	p.buffer = p.buffer[1:]

	if result.Err != nil {
		return types.Token{}, result.Err
	}

	if result.Token.Type == types.TokenEOF {
		p.sawEOF = true
	}
	return result.Token, nil
}

func (p *Parser) tryConsume(tokenType types.TokenType) (bool, error) {
	utils.Debug(fmt.Sprintf("[tryConsume want=%s buffer=%d sawEOF=%v]", tokenType, len(p.buffer), p.sawEOF))
	if p.sawEOF {
		return false, nil
	}

	if len(p.buffer) == 0 {
		result, ok := <-p.input
		if !ok {
			return false, fmt.Errorf("lexer input closed before EOF (tryConsume)")
		}
		if result.Err != nil {
			return false, result.Err
		}
		p.buffer = append(p.buffer, result)
	}

	result := p.buffer[0]
	if result.Err != nil {
		return false, result.Err
	}

	if result.Token.Type == types.TokenEOF {
		p.sawEOF = true
	}

	if result.Token.Type != tokenType {
		return false, nil
	}

	p.buffer = p.buffer[1:]
	return true, nil
}

func (p *Parser) AddLibs(libs string) {
	libsSlice := strings.SplitSeq(libs, ",")
	for lib := range libsSlice {
		lib = strings.TrimSpace(lib)
		if _, ok := p.libsSeen[lib]; ok {
			continue
		}
		p.Libs = append(p.Libs, lib)
		p.libsSeen[lib] = true
	}
}

func (p *Parser) expectSep() error {
	utils.Debug("[expectSep]")
	consumed := false
	for {
		if ok, err := p.tryConsume(types.TokenSemicolon); err != nil {
			return err
		} else if ok {
			consumed = true
			continue
		}

		if ok, err := p.tryConsume(types.TokenNewline); err != nil {
			return err
		} else if ok {
			consumed = true
			continue
		}

		if !consumed {
			token, err := p.peek(0)
			if err != nil {
				return err
			}
			return p.error("expected semicolon or newline", token)
		}
		return nil
	}
}

func (p *Parser) tryConsumeSep() (bool, error) {
	utils.Debug("[expectSep]")
	consumed := false
	for {
		if ok, err := p.tryConsume(types.TokenSemicolon); err != nil {
			return false, err
		} else if ok {
			consumed = true
			continue
		}

		if ok, err := p.tryConsume(types.TokenNewline); err != nil {
			return false, err
		} else if ok {
			consumed = true
			continue
		}

		return consumed, nil
	}
}
