package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfigOptimizationBounds(t *testing.T) {
	for _, opt := range []int{0, 1, 2, 3} {
		if err := validateConfig(BuildOptions{Opt: opt}); err != nil {
			t.Errorf("optimization level %d rejected: %v", opt, err)
		}
	}
	for _, opt := range []int{-1, 4} {
		if err := validateConfig(BuildOptions{Opt: opt}); err == nil {
			t.Errorf("optimization level %d accepted, want error", opt)
		}
	}
}

func TestValidateTarget(t *testing.T) {
	for _, target := range []string{
		"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64",
	} {
		if err := validateTarget("linux", target); err != nil {
			t.Errorf("supported target %q rejected on Linux: %v", target, err)
		}
	}

	if err := validateTarget("linux", "plan9/amd64"); err == nil || !strings.Contains(err.Error(), "unsupported target") {
		t.Errorf("unsupported target error = %v", err)
	}
	if err := validateTarget("linux", "windows/amd64"); err == nil || !strings.Contains(err.Error(), "cross-compiling to Windows") {
		t.Errorf("Windows cross-target error = %v", err)
	}
	if err := validateTarget("windows", "windows/arm64"); err != nil {
		t.Errorf("Windows host rejected Windows target: %v", err)
	}
}

func TestPrepareOutputDirCreatesParent(t *testing.T) {
	output := filepath.Join(t.TempDir(), "nested", "program")
	if err := prepareOutputDir(output); err != nil {
		t.Fatalf("prepare output directory: %v", err)
	}
	info, err := os.Stat(filepath.Dir(output))
	if err != nil {
		t.Fatalf("stat output directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("output parent is not a directory")
	}
}

func TestValidateInputs(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "program.wk")
	if err := os.WriteFile(input, []byte("exit(0)"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateInputs([]string{input}); err != nil {
		t.Errorf("regular input rejected: %v", err)
	}
	if err := validateInputs(nil); err != nil {
		t.Errorf("empty input list rejected: %v", err)
	}
	if err := validateInputs([]string{filepath.Join(dir, "missing.wk")}); err == nil {
		t.Error("missing input accepted, want error")
	}
	if err := validateInputs([]string{dir}); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("directory input error = %v", err)
	}
}
