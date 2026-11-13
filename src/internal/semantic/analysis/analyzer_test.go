package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/types"
)

func TestCheckProgramResolvesModulesAndQualifiedCalls(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "math.wk")
	if err := os.WriteFile(modulePath, []byte(`fn add(a: i64, b: i64): i64 {
	return a + b
}`), 0600); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "main.wk")
	program := parseTestProgram(t, rootPath, `import "math.wk"
let result = math.add(1, 2)`)

	if err := NewAnalyzer().CheckProgram(&program); err != nil {
		t.Fatalf("check program: %v", err)
	}

	importNode := program.Statements[0].Import
	if importNode.Namespace != "math" {
		t.Fatalf("namespace = %q, want math", importNode.Namespace)
	}
	if importNode.ResolvedPath != modulePath {
		t.Fatalf("resolved path = %q, want %q", importNode.ResolvedPath, modulePath)
	}
}

func TestCheckProgramRejectsMissingModule(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "main.wk")
	program := parseTestProgram(t, rootPath, `import "missing.wk"`)

	err := NewAnalyzer().CheckProgram(&program)
	if err == nil || !strings.Contains(err.Error(), "failed to read module") {
		t.Fatalf("expected missing-module error, got %v", err)
	}
}

func TestCheckProgramRejectsModuleCycles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.wk"), []byte(`import "b.wk"`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.wk"), []byte(`import "a.wk"`), 0600); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "main.wk")
	program := parseTestProgram(t, rootPath, `import "a.wk"`)

	err := NewAnalyzer().CheckProgram(&program)
	if err == nil || !strings.Contains(err.Error(), "cyclic module import") {
		t.Fatalf("expected cyclic-import error, got %v", err)
	}
}

func TestCheckProgramRejectsUnknownModuleFunction(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "math.wk"), []byte(`fn add(a: i64): i64 {
	return a
}`), 0600); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "main.wk")
	program := parseTestProgram(t, rootPath, `import "math.wk"
let result = math.subtract(1, 2)`)

	err := NewAnalyzer().CheckProgram(&program)
	if err == nil || !strings.Contains(err.Error(), "undeclared function: math.subtract") {
		t.Fatalf("expected unknown module function error, got %v", err)
	}
}

func TestCheckProgramRejectsNestedImports(t *testing.T) {
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "main.wk")
	program := parseTestProgram(t, rootPath, `if true {
	import "math.wk"
}`)

	err := NewAnalyzer().CheckProgram(&program)
	if err == nil || !strings.Contains(err.Error(), "imports are only allowed at the top level") {
		t.Fatalf("expected nested-import error, got %v", err)
	}
}

func parseTestProgram(t *testing.T, filename, source string) ast.NodeProgram {
	t.Helper()
	lex := lexer.NewLexer(source, filename)
	tokens := make(chan types.LexerResult, 4096)
	go lex.Tokenize(tokens)
	program, err := parser.NewParser().Parse(tokens)
	if err != nil {
		t.Fatalf("parse program: %v", err)
	}
	return program
}
