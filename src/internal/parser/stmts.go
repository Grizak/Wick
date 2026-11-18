package parser

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
	"github.com/Grizak/Wick/src/internal/utils"
)

func (p *Parser) parseBlock() (*ast.NodeBlock, error) {
	openBrace, err := p.expect(types.TokenOpenBrace)
	if err != nil {
		return nil, err
	}
	stmts, err := p.parseStatements(types.TokenCloseBrace)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(types.TokenCloseBrace); err != nil {
		return nil, err
	}
	return &ast.NodeBlock{
		Statements: stmts,
		Pos:        openBrace.Pos,
	}, nil
}

func (p *Parser) parseStatements(terminator types.TokenType) ([]ast.NodeStatement, error) {
	var stmts []ast.NodeStatement

	for {
		utils.Debug("[parseStatements loop]")
		tok, _ := p.peek(0)
		utils.Debug(fmt.Sprintf("[parseStatements peek=%s]", tok.Type))

		token, err := p.peek(0)
		if err != nil {
			return nil, err
		}

		if token.Type == types.TokenSemicolon || token.Type == types.TokenNewline {
			if _, err := p.consume(); err != nil {
				return nil, err
			}
			continue
		}

		if token.Type == terminator {
			return stmts, nil
		}
		if token.Type == types.TokenEOF {
			if terminator != types.TokenEOF {
				return nil, p.error("unexpected end of file", token)
			}
			return stmts, nil
		}

		switch token.Type {
		case types.TokenExit:
			if _, err := p.consume(); err != nil {
				return nil, err
			}
			exit, err := p.parseExit()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Exit: exit})
		case types.TokenLet:
			if _, err := p.consume(); err != nil {
				return nil, err
			}
			decl, err := p.parseVarDecl(false)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{VarDecl: decl})
		case types.TokenConst:
			if _, err := p.consume(); err != nil {
				return nil, err
			}
			decl, err := p.parseVarDecl(true)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{VarDecl: decl})
		case types.TokenIdent:
			next, err := p.peek(1)
			if err != nil {
				return nil, err
			}
			if next.Type == types.TokenEquals {
				assign, err := p.parseVarAssign()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, ast.NodeStatement{VarAssign: assign})
			} else {
				return nil, p.error("expected `=` after identifier", token)
			}
		case types.TokenOpenBrace:
			block, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Block: block})
		case types.TokenIf:
			ifTok, err := p.consume()
			if err != nil {
				return nil, err
			}
			ifStmt, err := p.parseIf(ifTok.Pos)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{If: ifStmt})
		case types.TokenFor:
			forTok, err := p.consume()
			if err != nil {
				return nil, err
			}
			forStmt, err := p.parseFor(forTok.Pos)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{For: forStmt})
		case types.TokenBreak:
			brkTok, err := p.consume()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Break: &ast.NodeBreak{Pos: brkTok.Pos}})

		case types.TokenContinue:
			contTok, err := p.consume()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Continue: &ast.NodeContinue{Pos: contTok.Pos}})
		case types.TokenFunctionDecl:
			fnToken, err := p.consume()
			if err != nil {
				return nil, err
			}
			name, err := p.expect(types.TokenIdent)
			if err != nil {
				return nil, err
			}
			funcDecl, err := p.parseFuncDecl()
			if err != nil {
				return nil, err
			}
			funcDecl.Pos = fnToken.Pos
			expr := ast.NodeExpression{FuncDecl: funcDecl, Pos: fnToken.Pos}
			stmts = append(stmts, ast.NodeStatement{VarDecl: &ast.NodeVarDecl{
				Name:  *name.Value,
				Type:  nil,
				Expr:  expr,
				Const: false,
				Pos:   expr.Pos,
			}})
		case types.TokenReturnStmt:
			returnstmt, err := p.parseReturnStmt()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Return: returnstmt})
		case types.TokenImport:
			importstmt, err := p.parseImportStmt()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Import: importstmt})
		case types.TokenAt:
			if _, err := p.consume(); err != nil {
				return nil, err
			}
			token, err := p.peek(0)
			if err != nil {
				return nil, err
			}
			switch token.Type {
			case types.TokenExtern:
				if _, err := p.consume(); err != nil {
					return nil, err
				}
				externStmt, err := p.parseExternStmt()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, ast.NodeStatement{Extern: externStmt})
			default:
				return nil, p.error("unexpected token", token)
			}
		default:
			return nil, p.error("unexpected token", token)
		}

		token, err = p.peek(0)
		if err != nil {
			return nil, err
		}
		if token.Type != terminator && token.Type != types.TokenEOF &&
			token.Type != types.TokenSemicolon && token.Type != types.TokenNewline {
			if err := p.expectSep(); err != nil {
				return nil, err
			}
		}
	}
}
