package parser

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/types"
)

// parseType parses type annotations (e.g., i32, *i64, [10]bool, string)
func (p *Parser) parseType() (string, error) {
	var typeStr strings.Builder

	token, err := p.peek(0)
	if err != nil {
		return "", err
	}

	// Handle pointer types
	for {
		tok, err := p.peek(0)
		if err != nil {
			return "", err
		}
		if tok.Type != types.TokenStar {
			break
		}
		_, err = p.consume()
		if err != nil {
			return "", err
		}
		typeStr.WriteString("*")
	}

	// Handle array types
	if token, err := p.peek(0); err != nil {
		return "", err
	} else if token.Type == types.TokenOpenBracket {
		_, err := p.consume()
		if err != nil {
			return "", err
		}
		lenToken, err := p.peek(0)
		if err != nil {
			return "", err
		}
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
	token, err = p.peek(0)
	if err != nil {
		return "", err
	}
	switch token.Type {
	case types.TokenIdent:
		_, err := p.consume()
		if err != nil {
			return "", err
		}
		switch *token.Value {
		case "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "f32", "f64", "bool", "string", "int", "float":
			typeStr.WriteString(*token.Value)
		default:
			return "", p.error(fmt.Sprintf("unknown type: %s", *token.Value), token)
		}
	default:
		return "", p.error(fmt.Sprintf("expected type, got %s", token.Type), token)
	}

	return typeStr.String(), nil
}
