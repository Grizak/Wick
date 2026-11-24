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
	stdDir       = "src/internal/llvm/toolchain/std"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "prepare-assets:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(changelogDir, 0o755); err != nil {
		return fmt.Errorf("create changelog directory: %w", err)
	}

	if err := copyFile("LICENSE", filepath.Join(assetsDir, "LICENSE")); err != nil {
		return fmt.Errorf("copy license: %w", err)
	}

	if err := copyReleaseNotes(); err != nil {
		return err
	}

	if err := copyTree("install", runtimeDir); err != nil {
		return fmt.Errorf("copy runtime files: %w", err)
	}

	if err := copyTree("std", stdDir); err != nil {
		return fmt.Errorf("copy standard library files: %w", err)
	}

	return nil
}

func copyReleaseNotes() error {
	notes, err := filepath.Glob("release-notes/*.md")
	if err != nil {
		return fmt.Errorf("find release notes: %w", err)
	}
	if len(notes) == 0 {
		return fmt.Errorf("no release notes found in release-notes/*.md")
	}

	for _, note := range notes {
		dst := filepath.Join(changelogDir, filepath.Base(note))
		if err := copyFile(note, dst); err != nil {
			return fmt.Errorf("copy release note %q: %w", note, err)
		}
	}

	return nil
}

func copyTree(srcRoot, dstRoot string) error {
	// Delete the old dstRoot
	err := os.RemoveAll(dstRoot)
	if err != nil {
		return err
	}

	// Copy the new files
	return filepath.WalkDir(srcRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("get relative path for %q: %w", path, err)
		}

		dst := filepath.Join(dstRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("create directory for %q: %w", dst, err)
		}

		if err := copyFile(path, dst); err != nil {
			return fmt.Errorf("copy %q to %q: %w", path, dst, err)
		}

		return nil
	})
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
