package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRuntimeHeaders(t *testing.T) {
	dir := t.TempDir()
	headers := map[string]string{
		"runtime.h": `#ifndef RUNTIME_H
#define RUNTIME_H
/* Runtime entry point and output API. */
int main(int argc, char *argv[]);
void print(const char *message);
void set_status(int);
#endif
`,
		"args.h": `int args_get_argc(void);
char **args_get_argv();
`,
	}
	for name, content := range headers {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	functions, err := parseRuntimeHeaders(dir)
	if err != nil {
		t.Fatalf("parse runtime headers: %v", err)
	}
	got := make(map[string]string, len(functions))
	for _, function := range functions {
		got[function.Name] = externSignature(&function)
	}
	want := map[string]string{
		"main":          "i32(i32,**i8)",
		"print":         "void(*i8)",
		"set_status":    "void(i32)",
		"args_get_argc": "i32()",
		"args_get_argv": "**i8()",
	}
	for name, signature := range want {
		if got[name] != signature {
			t.Errorf("%s signature = %q, want %q", name, got[name], signature)
		}
	}
	if len(got) != len(want) {
		t.Errorf("parsed %d functions, want %d", len(got), len(want))
	}
}

func TestParseRuntimeHeadersRejectsUnsupportedFunctionTypes(t *testing.T) {
	dir := t.TempDir()
	header := filepath.Join(dir, "runtime.h")
	if err := os.WriteFile(header, []byte("void run(struct task *task);\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := parseRuntimeHeaders(dir)
	if err == nil || !strings.Contains(err.Error(), "unsupported C type") || !strings.Contains(err.Error(), header) {
		t.Fatalf("error = %v, want unsupported C type with header location", err)
	}
}
