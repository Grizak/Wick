package parser_test

import (
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/types"
)

func TestParsePreservesProgramAndExpressionStructure(t *testing.T) {
	program := parse(t, "sample.wk", `let answer: i64 = 2 + 3 * 4
exit(answer)`)
	if len(program.Statements) != 2 {
		t.Fatalf("got %d statements, want 2", len(program.Statements))
	}

	decl := program.Statements[0].VarDecl
	if decl == nil || decl.Name != "answer" || decl.Type == nil || *decl.Type != "i64" {
		t.Fatalf("unexpected variable declaration: %#v", program.Statements[0])
	}
	sum := decl.Expr.BinExpr
	if sum == nil || sum.Op != types.BinOpAdd {
		t.Fatalf("initializer root = %#v, want addition", decl.Expr)
	}
	product := sum.Right.BinExpr
	if product == nil || product.Op != types.BinOpMul {
		t.Fatalf("right operand = %#v, want multiplication", sum.Right)
	}
	if *sum.Left.IntLit != 2 || *product.Left.IntLit != 3 || *product.Right.IntLit != 4 {
		t.Fatalf("unexpected arithmetic operands: %#v", decl.Expr)
	}

	exit := program.Statements[1].Exit
	if exit == nil || exit.Expr.Ident == nil || *exit.Expr.Ident != "answer" {
		t.Fatalf("unexpected exit statement: %#v", program.Statements[1])
	}
}

func TestParseReturnsLocatedSyntaxAndLexerErrors(t *testing.T) {
	t.Run("syntax error", func(t *testing.T) {
		_, err := parseWithError(t, "broken.wk", "let value =")
		if err == nil || !strings.Contains(err.Error(), "broken.wk:1:12: error:") {
			t.Fatalf("expected located syntax error, got %v", err)
		}
	})

	t.Run("lexer error", func(t *testing.T) {
		_, err := parseWithError(t, "broken.wk", "let value = @")
		if err == nil || !strings.Contains(err.Error(), "unexpected character") {
			t.Fatalf("expected lexer error, got %v", err)
		}
	})
}

func parse(t *testing.T, filename, source string) ast.NodeProgram {
	t.Helper()
	program, err := parseWithError(t, filename, source)
	if err != nil {
		t.Fatalf("parse program: %v", err)
	}
	return program
}

func parseWithError(t *testing.T, filename, source string) (ast.NodeProgram, error) {
	t.Helper()
	input := make(chan types.LexerResult, len([]rune(source))+1)
	lexer.NewLexer(source, filename).Tokenize(input)
	return parser.NewParser().Parse(input)
}
