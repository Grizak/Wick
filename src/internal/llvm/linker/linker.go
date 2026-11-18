package linker

import (
	"fmt"
	"os"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/Grizak/Wick/src/internal/utils"
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
			// ... your existing macOS branch, unchanged ...
		}

		if target == "x86_64-pc-linux-gnu" || target == "aarch64-pc-linux-gnu" {
			// Temporary CRT hack, same style as the -L flags below.
			// TODO: Embed these files in the binary and extract them to predicable paths
			var crtDir string
			if target == "x86_64-pc-linux-gnu" {
				crtDir = "/usr/lib/x86_64-linux-gnu"
			} else {
				crtDir = "/usr/lib/aarch64-linux-gnu"
			}

			// Before obj: real _start + init prologue
			args = append(args,
				crtDir+"/crt1.o",
				crtDir+"/crti.o",
				"-no-pie",
				"-dynamic-linker", "/lib64/ld-linux-x86-64.so.2",
			)

			libDirs := []string{
				"-L", "/usr/lib/x86_64-linux-gnu",
				"-L", "/lib/x86_64-linux-gnu",
				"-L", "/usr/lib",
			}
			args = append(args, libDirs...)

			args = append(args, "-o", outFile)
		} else {
			args = append(args, "-o", outFile)
		}
	}

	args = append(args, obj)

	if target == "x86_64-pc-linux-gnu" || target == "aarch64-pc-linux-gnu" {
		// After obj: libc + init epilogue
		var crtDir string
		if target == "x86_64-pc-linux-gnu" {
			crtDir = "/usr/lib/x86_64-linux-gnu"
		} else {
			crtDir = "/usr/lib/aarch64-linux-gnu"
		}
		args = append(args, "-lc", crtDir+"/crtn.o")
	}

	args = append(args, ldflags...)

	utils.Debug(fmt.Sprintf("Calling lld with %d args: %v", len(args), args))

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
