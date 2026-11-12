package lexer_test

import (
	"testing"

	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/types"
)

func TestTokenizeValuesOperatorsAndPositions(t *testing.T) {
	input := `let café = 12.5 + 3 * "hi" // comment
exit(café >= 4)`
	results := tokenize(t, input)

	wantTypes := []types.TokenType{
		types.TokenLet, types.TokenIdent, types.TokenEquals, types.TokenFloatLit,
		types.TokenPlus, types.TokenIntLit, types.TokenStar, types.TokenStringLit,
		types.TokenExit, types.TokenOpenParen, types.TokenIdent, types.TokenGtEq,
		types.TokenIntLit, types.TokenCloseParen, types.TokenEOF,
	}
	if len(results) != len(wantTypes) {
		t.Fatalf("got %d results, want %d", len(results), len(wantTypes))
	}
	for i, result := range results {
		if result.Err != nil {
			t.Fatalf("result %d: unexpected error: %v", i, result.Err)
		}
		if result.Token.Type != wantTypes[i] {
			t.Errorf("result %d type = %q, want %q", i, result.Token.Type, wantTypes[i])
		}
	}

	if got := *results[1].Token.Value; got != "café" {
		t.Errorf("identifier = %q, want café", got)
	}
	if got := results[1].Token.Pos; got != (types.Position{Line: 1, Column: 5, Index: 4}) {
		t.Errorf("identifier position = %+v, want line 1, column 5, index 4", got)
	}
	if got := results[8].Token.Pos.Line; got != 2 {
		t.Errorf("exit token line = %d, want 2", got)
	}
}

func TestTokenizeRejectsUnexpectedAndUnterminatedInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    string
		wantColumn int
	}{
		{name: "unexpected character", input: "let x = @", wantErr: "unexpected character", wantColumn: 9},
		{name: "standalone bang", input: "!", wantErr: "unexpected character '!'", wantColumn: 1},
		{name: "unterminated string", input: `"unterminated`, wantErr: "unterminated string literal", wantColumn: 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := tokenize(t, tt.input)
			if len(results) == 0 || results[len(results)-1].Err == nil {
				t.Fatalf("got results %#v, want a lexer error", results)
			}
			for i, result := range results[:len(results)-1] {
				if result.Err != nil {
					t.Fatalf("result %d: unexpected earlier error: %v", i, result.Err)
				}
			}
			err := results[len(results)-1].Err
			if got := err.Msg; got != tt.wantErr {
				t.Errorf("error message = %q, want %q", got, tt.wantErr)
			}
			if got := err.Pos.Column; got != tt.wantColumn {
				t.Errorf("error column = %d, want %d", got, tt.wantColumn)
			}
		})
	}
}

func tokenize(t *testing.T, input string) []types.LexerResult {
	t.Helper()
	output := make(chan types.LexerResult, len([]rune(input))+1)
	lexer.NewLexer(input, "test.wk").Tokenize(output)

	var results []types.LexerResult
	for result := range output {
		results = append(results, result)
	}
	return results
}
