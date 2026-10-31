package parser

import (
	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (p *Parser) parseImportStmt() (*ast.NodeImport, error) {
	// Consume the 'import' keyword
	_, err := p.expect(types.TokenImport)
	if err != nil {
		return nil, err
	}

	// Expect a string literal for the module path
	pathToken, err := p.expect(types.TokenStringLiteral)
	if err != nil {
		return nil, err
	}

	// Create the import statement node
	importStmt := &ast.NodeImport{
		Module: pathToken.Literal,
		Pos:    pathToken.Pos,
	}
	return importStmt, nil
}