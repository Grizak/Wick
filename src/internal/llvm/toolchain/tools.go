package toolchain

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed libc/windows_amd64/kernel32.lib
var kernel32Amd64Lib []byte

//go:embed libc/windows_arm64/kernel32.lib
var kernel32Arm64Lib []byte

func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "wick", "tools")
	return dir, nil
}

func Init() error {
	dir, err := CacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cache, err := Kernel32CacheDir()
	if err != nil {
		return err
	}

	for _, b := range []struct {
		dir  string
		name string
		data []byte
	}{
		{cache, "kernel32.amd64.lib", kernel32Amd64Lib},
		{cache, "kernel32.arm64.lib", kernel32Arm64Lib},
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

func Kernel32CacheDir() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "kernel32"), nil
}
