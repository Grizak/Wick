package types_test

import (
	"testing"

	"github.com/Grizak/Wick/src/internal/types"
)

func TestCompileErrorFormatting(t *testing.T) {
	pos := types.Position{Line: 3, Column: 7}
	tests := []struct {
		name string
		err  types.CompileError
		want string
	}{
		{
			name: "without position",
			err:  types.CompileError{File: "main.wk", Msg: "bad input"},
			want: "main.wk: error: bad input",
		},
		{
			name: "with position",
			err:  types.CompileError{File: "main.wk", Pos: &pos, Msg: "bad input"},
			want: "main.wk:3:7: error: bad input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
