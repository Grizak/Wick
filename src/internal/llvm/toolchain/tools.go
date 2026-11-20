package toolchain

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed libc
var embeddedLibc embed.FS

func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "wick", "tools")
	return dir, nil
}

// LibcCacheDir returns e.g. ~/.cache/wick/tools/libc/linux_amd64
func LibcCacheDir(platform string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "libc", platform), nil
}

// LibcDir is what Link() should call.
// platform is your existing style: linux_amd64, darwin_arm64, windows_amd64, etc.
func LibcDir(platform string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}
	return LibcCacheDir(platform)
}

func Init() error {
	dir, err := CacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Extract everything under libc/ to ~/.cache/wick/tools/libc/<platform>/
	err = fs.WalkDir(embeddedLibc, "libc", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := embeddedLibc.ReadFile(path)
		if err != nil {
			return err
		}

		// path = libc/linux_amd64/crt1.o
		// rel  = linux_amd64/crt1.o
		rel := strings.TrimPrefix(path, "libc/")
		outPath := filepath.Join(dir, "libc", rel)

		// Skip if already cached - same logic you had
		if _, err := os.Stat(outPath); err == nil {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(outPath, data, 0644)
	})
	if err != nil {
		return err
	}

	return nil
}
