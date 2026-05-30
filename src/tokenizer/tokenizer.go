package tokenizer

import (
	"unicode"

	"github.com/Grizak/Wick/src/types"
)

type Tokenizer struct {
	content string
	index   int
	line    int
	column  int
}

var singleCharTokens = map[rune]types.TokenType{
	'(': types.TokenOpenParen,
	')': types.TokenCloseParen,
	'+': types.TokenPlus,
	'-': types.TokenMinus,
	'*': types.TokenStar,
	'/': types.TokenFSlash,
}

func NewTokenizer(content string) *Tokenizer {
	t := Tokenizer{
		content: content,
		index:   0,
		line:    1,
		column:  1,
	}

	return &t
}

// Goroutine, write to output, block if output is full
func (t *Tokenizer) Tokenize(output chan types.Token) {
	for {
		r := t.peek(0)

		if r == 0 {
			output <- types.Token{Type: types.TokenEOF, Pos: t.pos()}
			break
		}

		// Single character tokens
		if tokenType, ok := singleCharTokens[r]; ok {
			t.consume()
			output <- types.Token{Type: tokenType, Pos: t.pos()}
			continue
		}

		// Whitespace
		if unicode.IsSpace(r) {
			t.consume()
			continue
		}

		// Multi-character tokens
		var buffer []rune
		isNumber := unicode.IsDigit(r)

		for {
			r := t.peek(0)
			if r == 0 || !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				break
			}
			buffer = append(buffer, t.consume())
			if len(buffer) > 4096 {
				panic("token too long")
			}
		}

		if isNumber {
			str := string(buffer)
			output <- types.Token{Type: types.TokenIntLit, Value: &str, Pos: t.pos()}
		} else {
			switch string(buffer) {
			case "exit":
				output <- types.Token{Type: types.TokenExit, Pos: t.pos()}
			default:
				panic("not implemented: identifiers")
			}
		}
	}
}

func (t *Tokenizer) peek(offset int) rune {
	// Bounds check
	if t.index+offset >= len(t.content) {
		return 0
	}

	return rune(t.content[t.index+offset])
}

func (t *Tokenizer) consume() rune {
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

func (t *Tokenizer) pos() types.Position {
	return types.Position{
		Line:   t.line,
		Column: t.column,
		Index:  t.index,
	}
}
