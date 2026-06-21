package lexer

import (
	"unicode"

	"github.com/Grizak/Wick/src/internal/types"
)

const (
	EOF            rune = -1
	maxTokenLength int  = 4096
)

var singleCharTokens = map[rune]types.TokenType{
	'(': types.TokenOpenParen,
	')': types.TokenCloseParen,
	'{': types.TokenOpenBrace,
	'}': types.TokenCloseBrace,
	'[': types.TokenOpenBracket,
	']': types.TokenCloseBracket,
	',': types.TokenComma,
	'+': types.TokenPlus,
	'-': types.TokenMinus,
	'*': types.TokenStar,
	'/': types.TokenFSlash,
	':': types.TokenColon,
	//'=': types.TokenEquals, // Handled by lexOperator
	';': types.TokenSemicolon,
}

var keywords = map[string]types.TokenType{
	"exit":     types.TokenExit,
	"const":    types.TokenConst,
	"let":      types.TokenLet,
	"true":     types.TokenTrue,
	"false":    types.TokenFalse,
	"if":       types.TokenIf,
	"else":     types.TokenElse,
	"for":      types.TokenFor,
	"break":    types.TokenBreak,
	"continue": types.TokenContinue,
	"fn":       types.TokenFunctionDecl,
	"return":   types.TokenReturnStmt,
}

type Lexer struct {
	content []rune // []rune instead of string for correct Unicode indexing
	index   int
	line    int
	column  int
	file    string
}

func NewLexer(content, filename string) *Lexer {
	return &Lexer{
		content: []rune(content),
		line:    1,
		column:  1,
		file:    filename,
	}
}

func (t *Lexer) Tokenize(output chan types.LexerResult) {
	for {
		r := t.peek(0)

		if r == EOF {
			output <- types.LexerResult{Token: types.Token{Type: types.TokenEOF, Pos: t.pos()}}
			return
		}

		if unicode.IsSpace(r) {
			t.consume()
			continue
		}

		if r == '/' && t.peek(1) == '/' {
			t.skipLineComment()
			continue
		}

		if r == '<' || r == '>' || r == '=' || r == '!' {
			if !t.lexOperator(output) {
				return
			}
			continue
		}

		if r == '"' {
			if !t.lexString(output) {
				return
			}
			continue
		}

		if tokenType, ok := singleCharTokens[r]; ok {
			pos := t.pos()
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: tokenType, Pos: pos}}
			continue
		}

		if unicode.IsDigit(r) {
			if !t.lexNumber(output) {
				return
			}
			continue
		}

		if unicode.IsLetter(r) || r == '_' {
			if !t.lexIdentifier(output) {
				return
			}
			continue
		}

		pos := t.pos()
		t.consume()
		output <- types.LexerResult{Err: t.error("unexpected character", pos)}
		return
	}
}

func (t *Lexer) skipLineComment() {
	t.consume() // first /
	t.consume() // second /
	for t.peek(0) != '\n' && t.peek(0) != EOF {
		t.consume()
	}
}

func (t *Lexer) lexString(output chan types.LexerResult) bool {
	pos := t.pos()
	t.consume() // opening "
	var buf []rune
	for {
		r := t.peek(0)
		if r == EOF {
			output <- types.LexerResult{Err: t.error("unterminated string literal", t.pos())}
			return false
		}
		if r == '"' {
			t.consume() // closing "
			break
		}
		buf = append(buf, t.consume())
	}
	str := string(buf)
	output <- types.LexerResult{Token: types.Token{Type: types.TokenStringLit, Value: &str, Pos: pos}}
	return true
}

func (t *Lexer) lexNumber(output chan types.LexerResult) bool {
	pos := t.pos()
	var buf []rune
	hasDecimal := false
	for {
		r := t.peek(0)
		if unicode.IsDigit(r) {
			buf = append(buf, t.consume())
		} else if r == '.' && !hasDecimal {
			hasDecimal = true
			buf = append(buf, t.consume())
		} else {
			break
		}
		if len(buf) > maxTokenLength {
			output <- types.LexerResult{Err: t.error("token too long", pos)}
			return false
		}
	}
	str := string(buf)
	tokenType := types.TokenIntLit
	if hasDecimal {
		tokenType = types.TokenFloatLit
	}
	output <- types.LexerResult{Token: types.Token{Type: tokenType, Value: &str, Pos: pos}}
	return true
}

func (t *Lexer) lexIdentifier(output chan types.LexerResult) bool {
	pos := t.pos()
	var buf []rune
	for {
		r := t.peek(0)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			buf = append(buf, t.consume())
		} else {
			break
		}
		if len(buf) > maxTokenLength {
			output <- types.LexerResult{Err: t.error("token too long", pos)}
			return false
		}
	}
	str := string(buf)
	if tokenType, ok := keywords[str]; ok {
		output <- types.LexerResult{Token: types.Token{Type: tokenType, Pos: pos}}
	} else {
		output <- types.LexerResult{Token: types.Token{Type: types.TokenIdent, Value: &str, Pos: pos}}
	}
	return true
}

func (t *Lexer) peek(offset int) rune {
	if t.index+offset >= len(t.content) {
		return EOF
	}
	return t.content[t.index+offset]
}

func (t *Lexer) consume() rune {
	if t.index >= len(t.content) {
		return EOF
	}
	r := t.content[t.index]
	t.index++
	if r == '\n' {
		t.line++
		t.column = 1
	} else {
		t.column++
	}
	return r
}

func (t *Lexer) pos() types.Position {
	return types.Position{Line: t.line, Column: t.column, Index: t.index}
}

func (t *Lexer) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{File: t.file, Pos: &pos, Msg: msg}
}

func (t *Lexer) lexOperator(output chan types.LexerResult) bool {
	pos := t.pos()
	r := t.consume()

	switch r {
	case '<':
		if t.peek(0) == '=' {
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: types.TokenLtEq, Pos: pos}}
		} else {
			output <- types.LexerResult{Token: types.Token{Type: types.TokenLt, Pos: pos}}
		}
	case '>':
		if t.peek(0) == '=' {
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: types.TokenGtEq, Pos: pos}}
		} else {
			output <- types.LexerResult{Token: types.Token{Type: types.TokenGt, Pos: pos}}
		}
	case '=':
		if t.peek(0) == '=' {
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: types.TokenEqEq, Pos: pos}}
		} else {
			output <- types.LexerResult{Token: types.Token{Type: types.TokenEquals, Pos: pos}}
		}
	case '!':
		if t.peek(0) == '=' {
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: types.TokenNotEq, Pos: pos}}
		} else {
			output <- types.LexerResult{Err: t.error("unexpected character '!'", pos)}
			return false
		}
	}
	return true
}
