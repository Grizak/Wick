package tools

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var tmpDir string

func Init() error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpDir = dir

	versions := make(map[string]string)
	if err := json.Unmarshal(versionsJSON, &versions); err != nil {
		return err
	}

	lldName := "lld-" + versions["lld"]
	if runtime.GOOS == "windows" {
		lldName = lldName + ".exe"
	}

	nasmName := "nasm-" + versions["nasm"]
	if runtime.GOOS == "windows" {
		nasmName = nasmName + ".exe"
	}

	for _, b := range []struct {
		name string
		data []byte
	}{
		{nasmName, nasmBinary},
		{lldName, lldBinary},
	} {
		path := filepath.Join(tmpDir, b.name)
		if err := os.WriteFile(path, b.data, 0755); err != nil {
			return err
		}
	}
	return nil
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

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "wick", "tools")
	return dir, nil
}
