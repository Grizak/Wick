package codegen

import (
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
	"tinygo.org/x/go-llvm"
)

func TestGenerateBuildsVerifiedLLVMModule(t *testing.T) {
	source := `fn add(a: i64, b: i64 = 2): i64 {
	return a + b
}
	fn increment(value: i32): i32 {
		return value + 1
	}
	let result = add(5)
	let narrow: i32 = increment(1)
	narrow = narrow + 1
	if result > 4 {
		result = result + 1
	} else {
	result = result + 2
}
for let i = 0; i < 2; i = i + 1 {
	if i == 1 {
		break
	}
}
exit(result)
`

	lex := lexer.NewLexer(source, "test.wk")
	tokens := make(chan types.LexerResult, 4096)
	go lex.Tokenize(tokens)
	program, err := parser.NewParser("test.wk").Parse(tokens)
	if err != nil {
		t.Fatalf("parse test program: %v", err)
	}

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("test.wk")
	defer mod.Dispose()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	target, err := target.NewTarget("linux/amd64")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := NewGenerator(&program, "test.wk").Generate(mod, target, builder, ctx); err != nil {
		t.Fatalf("generate LLVM module: %v", err)
	}

	ir := mod.String()
	for _, expected := range []string{"define void @_start()", "define i64 @func_0", "br i1"} {
		if !strings.Contains(ir, expected) {
			t.Errorf("generated IR does not contain %q:\n%s", expected, ir)
		}
	}
}
