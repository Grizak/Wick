package linker

import (
	"fmt"
	"os"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
)

func Link(objFiles []string, outFile string, save bool, target string, ldflags []string) error {
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

		// Windows args - lld-link
		args = append(args, "/OUT:"+outFile+".exe")
		args = append(args, "/ENTRY:main")
		args = append(args, "/SUBSYSTEM:CONSOLE")
		args = append(args, fmt.Sprintf("/MACHINE:%s", machine))
		args = append(args, fmt.Sprintf("/LIBPATH:%s", cache))
		args = append(args, kernel32)
	} else {
		args = append(args, "-o", outFile)
	}
	args = append(args, ldflags...)
	args = append(args, objFiles...)

	err := internalLink(target, args)

	if !save && err == nil {
		for _, objFile := range objFiles {
			if err := os.Remove(objFile); err != nil {
				return err
			}
		}
	} else {
		for counter, objFile := range objFiles {
			if err := os.Rename(objFile, outFile+fmt.Sprint(counter)+".o"); err != nil {
				return err
			}
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
