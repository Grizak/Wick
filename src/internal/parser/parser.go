package parser

import (
	"fmt"
	"strconv"
	"strings"

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

// Read from input, block when input is empty
func (p *Parser) Parse(input chan types.LexerResult) (ast.NodeProgram, error) {
	var program ast.NodeProgram
	program.Filename = p.filename
	p.input = input

	for {
		token := p.peek(0)

		switch token.Type {
		case types.TokenExit:
			_, err := p.consume()
			if err != nil {
				return program, err
			}
			exit, err := p.parseExit()
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, ast.NodeStatement{Exit: exit})
		case types.TokenConst:
			_, err := p.consume()
			if err != nil {
				return program, err
			}
			varDecl, err := p.parseVarDecl(true)
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, ast.NodeStatement{VarDecl: varDecl})
		case types.TokenIdent:
			// Peek ahead to disambiguate
			next := p.peek(1)
			if next.Type == types.TokenEquals {
				varAssign, err := p.parseVarAssign()
				if err != nil {
					return program, err
				}
				program.Statements = append(program.Statements, ast.NodeStatement{VarAssign: varAssign})
			} else {
				return program, p.error("expected `=` after identifier", token)
			}
		case types.TokenLet:
			_, err := p.consume()
			if err != nil {
				return program, err
			}
			varDecl, err := p.parseVarDecl(false)
			if err != nil {
				return program, err
			}
			program.Statements = append(program.Statements, ast.NodeStatement{VarDecl: varDecl})
		case types.TokenOpenParen:
			return program, p.error("unexpected `(`", token)
		case types.TokenCloseParen:
			return program, p.error("unexpected `)`", token)
		case types.TokenIntLit:
			return program, p.error("unexpected int literal", token)
		case types.TokenEOF:
			return program, nil
		default:
			return program, p.error("unexpected token", token)
		}
	}
}

func (p *Parser) parseExit() (*ast.NodeExit, error) {
	_, err := p.expect(types.TokenOpenParen)
	if err != nil {
		return &ast.NodeExit{}, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return &ast.NodeExit{}, err
	}
	_, err = p.expect(types.TokenCloseParen)
	if err != nil {
		return &ast.NodeExit{}, err
	}

	return &ast.NodeExit{Expr: expr, Pos: expr.Pos}, nil
}

func (p *Parser) parseExpression() (ast.NodeExpression, error) {
	term, err := p.parseTerm()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenPlus:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpAdd,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	case types.TokenMinus:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  term,
				Op:    types.BinOpSub,
				Right: right,
				Pos:   term.Pos,
			},
		}, nil
	}

	return term, nil
}

func (p *Parser) parseTerm() (ast.NodeExpression, error) {
	factor, err := p.parseFactor()
	if err != nil {
		return ast.NodeExpression{}, err
	}

	switch p.peek(0).Type {
	case types.TokenStar:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpMul,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	case types.TokenFSlash:
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		right, err := p.parseTerm()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{
			BinExpr: &ast.NodeBinExpr{
				Left:  factor,
				Op:    types.BinOpDiv,
				Right: right,
				Pos:   factor.Pos,
			},
		}, nil
	}
	return factor, nil
}

func (p *Parser) parseFactor() (ast.NodeExpression, error) {
	token := p.peek(0)

	if token.Type == types.TokenOpenParen {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		expr, err := p.parseExpression()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		_, err = p.expect(types.TokenCloseParen)
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return expr, nil
	}

	if token.Type == types.TokenIdent {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		// Check if this is a function call
		if p.peek(0).Type == types.TokenOpenParen {
			_, err := p.consume() // consume (
			if err != nil {
				return ast.NodeExpression{}, err
			}
			var args []ast.NodeExpression
			if p.peek(0).Type != types.TokenCloseParen {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return ast.NodeExpression{}, err
					}
					args = append(args, arg)
					if p.peek(0).Type != types.TokenComma {
						break
					}
					_, err = p.consume() // consume comma
					if err != nil {
						return ast.NodeExpression{}, err
					}
				}
			}
			_, err = p.expect(types.TokenCloseParen)
			if err != nil {
				return ast.NodeExpression{}, err
			}
			return ast.NodeExpression{
				FuncCall: &ast.NodeFuncCall{
					Name: *token.Value,
					Args: args,
					Pos:  token.Pos,
				},
			}, nil
		}
		return ast.NodeExpression{Ident: token.Value, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenIntLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		i, err := strconv.Atoi(*token.Value)
		if err != nil {
			return ast.NodeExpression{}, p.error("invalid int literal", token)
		}
		return ast.NodeExpression{IntLit: &i, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenFloatLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		f, err := strconv.ParseFloat(*token.Value, 64)
		if err != nil {
			return ast.NodeExpression{}, p.error("invalid float literal", token)
		}
		return ast.NodeExpression{FloatLit: &f, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenStringLit {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		return ast.NodeExpression{StringLit: token.Value, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenTrue {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		t := true
		return ast.NodeExpression{BoolLit: &t, Pos: token.Pos}, nil
	}

	if token.Type == types.TokenFalse {
		_, err := p.consume()
		if err != nil {
			return ast.NodeExpression{}, err
		}
		f := false
		return ast.NodeExpression{BoolLit: &f, Pos: token.Pos}, nil
	}

	return ast.NodeExpression{}, p.error(fmt.Sprintf("expected literal or identifier, got %s", token.Type), token)
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

func (p *Parser) parseVarDecl(isConst bool) (*ast.NodeVarDecl, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return nil, err
	}

	var typeName *string
	if p.peek(0).Type == types.TokenColon {
		_, err := p.consume()
		if err != nil {
			return nil, err
		}
		typeStr, err := p.parseType()
		if err != nil {
			return nil, err
		}
		typeName = &typeStr
	}

	_, err = p.expect(types.TokenEquals)
	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return &ast.NodeVarDecl{}, err
	}

	return &ast.NodeVarDecl{
		Name:  *name.Value,
		Type:  typeName,
		Expr:  expr,
		Const: isConst,
		Pos:   expr.Pos,
	}, nil
}

func (p *Parser) parseVarAssign() (*ast.NodeVarAssign, error) {
	name, err := p.expect(types.TokenIdent)
	if err != nil {
		return &ast.NodeVarAssign{}, err
	}
	_, err = p.expect(types.TokenEquals)
	if err != nil {
		return &ast.NodeVarAssign{}, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return &ast.NodeVarAssign{}, err
	}

	return &ast.NodeVarAssign{
		Name: *name.Value,
		Expr: expr,
		Pos:  expr.Pos,
	}, nil
}

// parseType parses type annotations (e.g., i32, *i64, [10]bool, string)
func (p *Parser) parseType() (string, error) {
	var typeStr strings.Builder

	// Handle pointer types
	for p.peek(0).Type == types.TokenStar {
		_, err := p.consume()
		if err != nil {
			return "", err
		}
		typeStr.WriteString("*")
	}

	// Handle array types
	if p.peek(0).Type == types.TokenOpenBracket {
		_, err := p.consume()
		if err != nil {
			return "", err
		}
		lenToken := p.peek(0)
		if lenToken.Type != types.TokenIntLit {
			return "", p.error("expected array length", lenToken)
		}
		_, err = p.consume()
		if err != nil {
			return "", err
		}
		typeStr.WriteString("[")
		typeStr.WriteString(*lenToken.Value)
		typeStr.WriteString("]")
		_, err = p.expect(types.TokenCloseBracket)
		if err != nil {
			return "", err
		}
	}

	// Parse base type
	token := p.peek(0)
	switch token.Type {
	case types.TokenIdent:
		_, err := p.consume()
		if err != nil {
			return "", err
		}
		switch *token.Value {
		case "i32", "i64", "f64", "bool", "string", "int", "float":
			typeStr.WriteString(*token.Value)
		default:
			return "", p.error(fmt.Sprintf("unknown type: %s", *token.Value), token)
		}
	default:
		return "", p.error(fmt.Sprintf("expected type, got %s", token.Type), token)
	}

	return typeStr.String(), nil
}
