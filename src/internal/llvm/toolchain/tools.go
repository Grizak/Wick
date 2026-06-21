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
		{LlcCacheDir(), LlcBinaryName(), llcBinary},
		{LldCacheDir(), LldBinaryName(), lldBinary},
		{LldCacheDir(), "kernel32.lib", kernel32Lib},
		{OptCacheDir(), OptBinaryName(), optBinary},
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

func LlcCacheDir() string {
	return filepath.Join(tmpDir, "llc-"+versions["llc"])
}

func OptCacheDir() string {
	return filepath.Join(tmpDir, "opt-"+versions["opt"])
}

func LldPath() string {
	return filepath.Join(LldCacheDir(), LldBinaryName())
}

func LlcPath() string {
	return filepath.Join(LlcCacheDir(), LlcBinaryName())
}

func OptPath() string {
	return filepath.Join(OptCacheDir(), OptBinaryName())
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

func LlcBinaryName() string {
	if runtime.GOOS == "windows" {
		return "llc.exe"
	}
	return "llc"
}

func OptBinaryName() string {
	if runtime.GOOS == "windows" {
		return "opt.exe"
	}
	return "opt"
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
