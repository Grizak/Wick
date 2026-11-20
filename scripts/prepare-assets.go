package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	assetsDir    = "src/internal/assets"
	changelogDir = "src/internal/assets/changelog"
	runtimeDir   = "src/internal/llvm/toolchain/runtime"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "prepare-assets:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(changelogDir, 0o755); err != nil {
		return err
	}
	if err := copyFile("LICENSE", filepath.Join(assetsDir, "LICENSE")); err != nil {
		return err
	}

	notes, err := filepath.Glob("release-notes/*.md")
	if err != nil {
		return err
	}
	if len(notes) == 0 {
		return fmt.Errorf("no release notes found in release-notes/*.md")
	}
	for _, note := range notes {
		if err := copyFile(note, filepath.Join(changelogDir, filepath.Base(note))); err != nil {
			return err
		}
	}

	err = filepath.WalkDir("install", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel("install", path)
		if err != nil {
			return err
		}
		dst := filepath.Join(runtimeDir, rel)

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return copyFile(path, dst)
	})
	if err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
