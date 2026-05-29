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
		{llcCacheDir(), llcBinaryName(), llcBinary},
		{lldCacheDir(), lldBinaryName(), lldBinary},
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

func lldCacheDir() string {
	return filepath.Join(tmpDir, "lld-"+versions["lld"])
}

func llcCacheDir() string {
	return filepath.Join(tmpDir, "llc-"+versions["llc"])
}

func LldPath() string {
	return filepath.Join(lldCacheDir(), lldBinaryName())
}

func LlcPath() string {
	return filepath.Join(llcCacheDir(), llcBinaryName())
}

func lldBinaryName() string {
	switch runtime.GOOS {
	case "windows":
		return "lld-link.exe"
	case "darwin":
		return "ld64.lld"
	default:
		return "ld.lld"
	}
}

func llcBinaryName() string {
	if runtime.GOOS == "windows" {
		return "llc.exe"
	}
	return "llc"
}

func toolsPath(name string) string {
	if tmpDir == "" {
		log.Fatal("tools.Init() has not been called")
	}
	return filepath.Join(tmpDir, name)
}

func ExecuteCommand(name string, args ...string) error {
	c := exec.Command(name, args...)
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
