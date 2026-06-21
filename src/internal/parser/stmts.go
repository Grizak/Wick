package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
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
		token := p.peek(0)

		if token.Type == terminator {
			return stmts, nil
		}

		switch token.Type {
		case types.TokenExit:
			p.consume()
			exit, err := p.parseExit()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Exit: exit})
		case types.TokenLet:
			p.consume()
			decl, err := p.parseVarDecl(false)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{VarDecl: decl})
		case types.TokenConst:
			p.consume()
			decl, err := p.parseVarDecl(true)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{VarDecl: decl})
		case types.TokenIdent:
			next := p.peek(1)
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
			p.consume()
			ifStmt, err := p.parseIf()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{If: ifStmt})
		case types.TokenFor:
			p.consume()
			forStmt, err := p.parseFor()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{For: forStmt})
		case types.TokenBreak:
			token, err := p.consume()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Break: &ast.NodeBreak{Pos: token.Pos}})

		case types.TokenContinue:
			token, err := p.consume()
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, ast.NodeStatement{Continue: &ast.NodeContinue{Pos: token.Pos}})
		case types.TokenFunctionDecl:
			_, err := p.consume() // fn
			if err != nil {
				return nil, err
			}

			name, err := p.expect(types.TokenIdent)
			if err != nil {
				return nil, err
			}

			expr, err := p.parseComparison()
			if err != nil {
				return nil, err
			}

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
		default:
			return nil, p.error("unexpected token", token)
		}
	}
}
