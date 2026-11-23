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

//go:embed runtime
var embeddedRuntime embed.FS

//go:embed std
var embeddedStdlib embed.FS

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

// RuntimeCacheDir returns e.g. ~/.cache/wick/tools/runtime/linux_amd64
func RuntimeCacheDir(platform string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "runtime", platform), nil
}

// LibcDir is what Link() should call.
// platform is your existing style: linux_amd64, darwin_arm64, windows_amd64, etc.
func LibcDir(platform string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}
	return LibcCacheDir(platform)
}

func RuntimeDir(platform string) (string, error) {
	if err := Init(); err != nil {
		return "", err
	}
	return RuntimeCacheDir(platform)
}

func RuntimeLibPath(platform string) (string, error) {
	dir, err := RuntimeDir(platform)
	if err != nil {
		return "", err
	}
	libPath := filepath.Join(dir, "lib")
	return libPath, nil
}

func RuntimeIncludePath(platform string) (string, error) {
	dir, err := RuntimeDir(platform)
	if err != nil {
		return "", err
	}
	includePath := filepath.Join(dir, "include")
	return includePath, nil
}

func RuntimeLibFile(platform string) (string, error) {
	dir, err := RuntimeLibPath(platform)
	if err != nil {
		return "", err
	}
	var libFile string
	if strings.HasPrefix(platform, "windows") {
		libFile = filepath.Join(dir, "wick_runtime.lib")
	} else {
		libFile = filepath.Join(dir, "libwick_runtime.a")
	}
	return libFile, nil
}

// Init initializes the cache directory and extracts the embedded libc and runtime files to it. Should look like
// ~/.cache/wick/tools/
// ├── libc
// │   ├── darwin_amd64
// │   │   ├── libSystem.B.tbd
// │   │   └── libSystem.tbd
// │   ├── darwin_arm64
// │   │   ├── libSystem.B.tbd
// │   │   └── libSystem.tbd
// │   ├── linux_amd64
// │   │   ├── crt1.o
// │   │   ├── crti.o
// │   │   ├── crtn.o
// │   │   ├── libc.a
// │   │   └── libm.a
// │   ├── linux_arm64
// │   │   ├── crt1.o
// │   │   ├── crti.o
// │   │   ├── crtn.o
// │   │   ├── libc.a
// │   │   └── libm.a
// │   ├── windows_amd64
// │   │   ├── crt2.o
// │   │   ├── crtend.o
// │   │   ├── libgcc.a
// │   │   ├── libkernel32.a
// │   │   ├── libmingw32.a
// │   │   ├── libmingwex.a
// │   │   ├── libmsvcrt.a
// │   │   └── libucrt.a
// │   └── windows_arm64
// │       ├── crt2.o
// │       ├── crtend.o
// │       ├── libgcc.a
// │       ├── libkernel32.a
// │       ├── libmingw32.a
// │       ├── libmingwex.a
// │       ├── libmsvcrt.a
// │       └── libucrt.a
// └── runtime
//     ├── linux-amd64
//     │   ├── include
//     │   │   ├── args.h
//     │   │   └── runtime.h
//     │   └── lib
//     │       └── libwick_runtime.a
//     ├── linux-arm64
//     │   ├── include
//     │   │   ├── args.h
//     │   │   └── runtime.h
//     │   └── lib
//     │       └── libwick_runtime.a
//     ├── macos-amd64
//     │   ├── include
//     │   │   ├── args.h
//     │   │   └── runtime.h
//     │   └── lib
//     │       └── libwick_runtime.a
//     ├── macos-arm64
//     │   ├── include
//     │   │   ├── args.h
//     │   │   └── runtime.h
//     │   └── lib
//     │       └── libwick_runtime.a
//     ├── windows-amd64
//     │   ├── include
//     │   │   ├── args.h
//     │   │   └── runtime.h
//     │   └── lib
//     │       └── wick_runtime.lib
//     └── windows-arm64
//         ├── include
//         │   ├── args.h
//         │   └── runtime.h
//         └── lib
//             └── wick_runtime.lib
//
// 27 directories, 48 files

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

	// Extract everything under runtime/ to ~/.cache/wick/tools/runtime/<platform>/
	err = fs.WalkDir(embeddedRuntime, "runtime", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := embeddedRuntime.ReadFile(path)
		if err != nil {
			return err
		}

		// path = runtime/runtime-linux-amd64/lib/libwick_runtime.a
		// rel  = linux_amd64/lib/libwick_runtime.a
		rel := strings.TrimPrefix(path, "runtime/runtime-")
		rel = strings.ReplaceAll(rel, "-", "_")
		outPath := filepath.Join(dir, "runtime", rel)

		// Skip if already cached
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

	// Extract everything under std/ to ~/.cache/wick/tools/std/<platform>/
	err = fs.WalkDir(embeddedStdlib, "std", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := embeddedStdlib.ReadFile(path)
		if err != nil {
			return err
		}

		// path = std/linux_amd64/std/fmt/fmt.wk
		// rel  = linux_amd64/std/fmt/fmt.wk
		rel := strings.TrimPrefix(path, "std/")
		outPath := filepath.Join(dir, "std", rel)

		// Skip if already cached
		if _, err := os.Stat(outPath); err != nil {
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
