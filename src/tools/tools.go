package tools

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var tmpDir string

func Init() error {
	dir, err := os.MkdirTemp("", "wick-tools-*")
	if err != nil {
		return err
	}
	tmpDir = dir

	lldName := "lld"
	if runtime.GOOS == "windows" {
		lldName = "lld.exe"
	}

	for _, b := range []struct {
		name string
		data []byte
	}{
		{"nasm", nasmBinary},
		{lldName, lldBinary},
	} {
		path := filepath.Join(tmpDir, b.name)
		if err := os.WriteFile(path, b.data, 0755); err != nil {
			return err
		}
	}
	return nil
}

func Cleanup() {
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}

func NasmPath() string {
	return toolsPath("nasm")
}

func LldPath() string {
	if runtime.GOOS == "windows" {
		return toolsPath("lld.exe")
	}
	return toolsPath("lld")
}

func toolsPath(name string) string {
	if tmpDir == "" {
		log.Fatal("tools.Init() has not been called")
	}
	return filepath.Join(tmpDir, name)
}

func ExecuteCommand(cmd string) error {
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
