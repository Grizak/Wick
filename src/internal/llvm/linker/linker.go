package linker

import (
	"fmt"
	"os"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
)

func Link(obj, outFile string, save bool, target string, ldflags []string) error {
	var args []string
	if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
		var kernel32, machine string
		if target == "x86_64-pc-windows-msvc" {
			kernel32 = "kernel32.amd64.lib"
			machine = "X64"
		} else if target == "aarch64-pc-windows-msvc" {
			kernel32 = "kernel32.arm64.lib"
			machine = "ARM64"
		}

		cache, err := toolchain.Kernel32CacheDir()
		if err != nil {
			return err
		}

		args = append(args, "/OUT:"+outFile+".exe")
		args = append(args, "/ENTRY:main")
		args = append(args, "/SUBSYSTEM:CONSOLE")
		args = append(args, fmt.Sprintf("/MACHINE:%s", machine))
		args = append(args, fmt.Sprintf("/LIBPATH:%s", cache))
		args = append(args, kernel32)
	} else {
		if target == "x86_64-apple-darwin" || target == "aarch64-apple-darwin" {
			var arch string
			switch target {
			case "x86_64-apple-darwin":
				arch = "x86_64"
			case "aarch64-apple-darwin":
				arch = "arm64"
			}
			args = append(args, "-arch", arch, "-platform_version", "macos", "13.0", "13.0")
		}
		args = append(args, "-o", outFile)
	}
	args = append(args, ldflags...)
	args = append(args, obj)

	err := internalLink(target, args)

	if !save && err == nil {
		if err := os.Remove(obj); err != nil {
			return err
		}
	} else {
		if err := os.Rename(obj, outFile+".o"); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func pushTarget(target string, args []string) ([]string, error) {
	var driver string

	switch target {
	case "elf":
		driver = "ld.lld"
	case "macho":
		driver = "ld64.lld"
	case "coff":
		driver = "lld-link"
	default:
		return nil, fmt.Errorf("unsupported target %q", target)
	}

	return append([]string{driver}, args...), nil
}
