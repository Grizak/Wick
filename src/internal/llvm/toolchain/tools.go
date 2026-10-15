package toolchain

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var tmpDir string

var versions map[string]string

func Init() error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpDir = dir

	versions = make(map[string]string)
	if err := json.Unmarshal(versionsJSON, &versions); err != nil {
		return err
	}

	for _, b := range []struct {
		dir  string
		name string
		data []byte
	}{
		{LldCacheDir(), LldBinaryName(), lldBinary},
		{LldCacheDir(), "kernel32.amd64.lib", kernel32Amd64Lib},
		{LldCacheDir(), "kernel32.arm64.lib", kernel32Arm64Lib},
	} {
		path := filepath.Join(b.dir, b.name)
		if _, err := os.Stat(path); err == nil {
			continue // already cached
		}
		if err := os.MkdirAll(b.dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, b.data, 0755); err != nil {
			return err
		}
	}
	return nil
}

func LldCacheDir() string {
	return filepath.Join(tmpDir, "lld-"+versions["lld"])
}

func LldPath() string {
	return filepath.Join(LldCacheDir(), LldBinaryName())
}

func LldBinaryName() string {
	switch runtime.GOOS {
	case "windows":
		return "lld-link.exe"
	case "darwin":
		return "ld64.lld"
	default:
		return "ld.lld"
	}
}

func ExecuteCommand(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()

	if err != nil {
		println("ERR:", err.Error())
	}

	return err
}

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "wick", "tools")
	return dir, nil
}
