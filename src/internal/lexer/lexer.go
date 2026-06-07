package lexer

import (
	"unicode"

	"github.com/Grizak/Wick/src/internal/types"
)

type Lexer struct {
	content string
	index   int
	line    int
	column  int
	file    string
}

const EOF rune = 0

var singleCharTokens = map[rune]types.TokenType{
	'(': types.TokenOpenParen,
	')': types.TokenCloseParen,
	'[': types.TokenOpenBracket,
	']': types.TokenCloseBracket,
	',': types.TokenComma,
	'+': types.TokenPlus,
	'-': types.TokenMinus,
	'*': types.TokenStar,
	'/': types.TokenFSlash,
	':': types.TokenColon,
	'=': types.TokenEquals,
}

func NewLexer(content, filename string) *Lexer {
	t := Lexer{
		content: content,
		index:   0,
		line:    1,
		column:  1,
		file:    filename,
	}

	return &t
}

// Goroutine, write to output, block if output is full
func (t *Lexer) Tokenize(output chan types.LexerResult) {
	for {
		r := t.peek(0)

		if r == 0 {
			output <- types.LexerResult{Token: types.Token{Type: types.TokenEOF, Pos: t.pos()}}
			break
		}

		// String literals
		if r == '"' {
			t.consume() // consume opening quote
			var buffer []rune
			for {
				r := t.peek(0)
				if r == 0 {
					output <- types.LexerResult{Err: t.error("unterminated string literal", t.pos())}
					return
				}
				if r == '"' {
					t.consume() // consume closing quote
					break
				}
				buffer = append(buffer, t.consume())
			}
			str := string(buffer)
			output <- types.LexerResult{Token: types.Token{Type: types.TokenStringLit, Value: &str, Pos: t.pos()}}
			continue
		}

		// Single character tokens
		if tokenType, ok := singleCharTokens[r]; ok {
			t.consume()
			output <- types.LexerResult{Token: types.Token{Type: tokenType, Pos: t.pos()}}
			continue
		}

		// Comments
		if t.peek(0) == '/' && t.peek(1) == '/' {
			for {
				if t.peek(0) == '\n' || t.peek(0) == EOF {
					break
				}
				t.consume()
			}
		}

		// Whitespace
		if unicode.IsSpace(r) {
			t.consume()
			continue
		}

		// Multi-character tokens (numbers and identifiers)
		var buffer []rune
		isNumber := unicode.IsDigit(r)

		for {
			r := t.peek(0)
			if r == 0 {
				break
			}
			// For numbers, allow digits and a single decimal point
			if isNumber {
				if unicode.IsDigit(r) {
					buffer = append(buffer, t.consume())
				} else if r == '.' && !containsDecimal(string(buffer)) {
					buffer = append(buffer, t.consume())
				} else {
					break
				}
			} else {
				// For identifiers, allow letters and digits
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					buffer = append(buffer, t.consume())
				} else {
					break
				}
			}
			if len(buffer) > 4096 {
				output <- types.LexerResult{Err: t.error("token too long", t.pos())}
			}
		}

		str := string(buffer)

		if isNumber {
			// Check if it's a float or int
			if containsDecimal(str) {
				output <- types.LexerResult{Token: types.Token{Type: types.TokenFloatLit, Value: &str, Pos: t.pos()}}
			} else {
				output <- types.LexerResult{Token: types.Token{Type: types.TokenIntLit, Value: &str, Pos: t.pos()}}
			}
		} else {
			switch str {
			case "exit":
				output <- types.LexerResult{Token: types.Token{Type: types.TokenExit, Pos: t.pos()}}
			case "const":
				output <- types.LexerResult{Token: types.Token{Type: types.TokenConst, Pos: t.pos()}}
			case "let":
				output <- types.LexerResult{Token: types.Token{Type: types.TokenLet, Pos: t.pos()}}
			case "true":
				output <- types.LexerResult{Token: types.Token{Type: types.TokenTrue, Pos: t.pos()}}
			case "false":
				output <- types.LexerResult{Token: types.Token{Type: types.TokenFalse, Pos: t.pos()}}
			default:
				output <- types.LexerResult{Token: types.Token{Type: types.TokenIdent, Value: &str, Pos: t.pos()}}
			}
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			t.consume()
			output <- types.LexerResult{Err: t.error("unexpected character", t.pos())}
			return
		}
	}
}

func (t *Lexer) peek(offset int) rune {
	// Bounds check
	if t.index+offset >= len(t.content) {
		return 0
	}

	return rune(t.content[t.index+offset])
}

func (t *Lexer) consume() rune {
	if t.index >= len(t.content) {
		return 0
	}

	r := rune(t.content[t.index])
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
	return types.Position{
		Line:   t.line,
		Column: t.column,
		Index:  t.index,
	}
}

func (t *Lexer) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: t.file,
		Pos:  &pos,
		Msg:  msg,
	}
}

func containsDecimal(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}
