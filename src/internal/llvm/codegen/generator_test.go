package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/semantic/analysis"
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

func TestGenerateEmitsImportedModuleFunctions(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "math.wk")
	if err := os.WriteFile(modulePath, []byte(`fn add(a: i64, b: i64): i64 {
	return a + b
}`), 0600); err != nil {
		t.Fatal(err)
	}
	wrapperPath := filepath.Join(dir, "arithmetic.wk")
	if err := os.WriteFile(wrapperPath, []byte(`import "math.wk"
fn addFive(value: i64): i64 {
	return math.add(value, 5)
}`), 0600); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "main.wk")
	source := `import "arithmetic.wk"
let result = arithmetic.addFive(5)
exit(result)`

	lex := lexer.NewLexer(source, rootPath)
	tokens := make(chan types.LexerResult, 4096)
	go lex.Tokenize(tokens)
	program, err := parser.NewParser(rootPath).Parse(tokens)
	if err != nil {
		t.Fatalf("parse test program: %v", err)
	}
	if err := analysis.NewAnalyzer(rootPath).CheckProgram(&program); err != nil {
		t.Fatalf("analyze test program: %v", err)
	}

	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("main.wk")
	defer mod.Dispose()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	target, err := target.NewTarget("linux/amd64")
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := NewGenerator(&program, rootPath).Generate(mod, target, builder, ctx); err != nil {
		t.Fatalf("generate LLVM module: %v", err)
	}

	ir := mod.String()
	for _, expected := range []string{
		"define i64 @func_0",
		"define i64 @func_1",
		"call i64 @func_0",
		"call i64 @func_1",
	} {
		if !strings.Contains(ir, expected) {
			t.Errorf("generated IR does not contain %q:\n%s", expected, ir)
		}
	}
}
