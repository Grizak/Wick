package tokenizer

import (
	"unicode"

	"github.com/Grizak/Wick/src/types"
)

type Tokenizer struct {
	content string
	index   int
}

func NewTokenizer(content string) *Tokenizer {
	t := Tokenizer{
		content: content,
		index:   0,
	}

	return &t
}

// Goroutine, write to output, block if output is full
func (t *Tokenizer) Tokenize(output chan types.Token) {
	for {
		r := t.peek(0)

		if r == 0 {
			output <- types.Token{Type: types.TokenEOF, Line: t.line()}
			break
		}

		switch types.TokenType(r) {
		case types.TokenOpenParen:
			t.consume()
			output <- types.Token{Type: types.TokenOpenParen, Line: t.line()}
			continue
		case types.TokenCloseParen:
			t.consume()
			output <- types.Token{Type: types.TokenCloseParen, Line: t.line()}
			continue
		case types.TokenPlus:
			t.consume()
			output <- types.Token{Type: types.TokenPlus, Line: t.line()}
			continue
		default:
			// Check if it's whitespace
			if unicode.IsSpace(r) {
				t.consume()
				continue
			}

			// Otherwise, read it into a buffer until we hit whitespace
			var buffer []rune
			isNumber := unicode.IsDigit(r)

			for {
				r := t.peek(0)
				if r == 0 {
					break
				}
				if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
					break
				}

				buffer = append(buffer, t.consume())

				if len(buffer) > 4096 {
					panic("token too long")
				}
			}

			// Now check what we read
			if isNumber {
				// We have a number
				str := string(buffer)
				output <- types.Token{Type: types.TokenIntLit, Value: &str, Line: t.line()}
			} else {
				// Check if it's a keyword
				switch string(buffer) {
				case "exit":
					output <- types.Token{Type: types.TokenExit, Line: t.line()}
				default: // Identifier
					panic("Not implemented: identifiers")
				}
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
	return r
}

func (t *Tokenizer) line() int {
	line := 1
	for i := 0; i < t.index; i++ {
		if t.content[i] == '\n' {
			line++
		}
	}

	return line
}
