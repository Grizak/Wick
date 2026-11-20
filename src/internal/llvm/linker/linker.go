package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/Grizak/Wick/src/internal/utils"
	"tinygo.org/x/go-llvm"
)

func Link(obj, outFile string, save bool, target string, ldflags []string, targetMachine llvm.TargetMachine) error {
	var args []string

	switch target {
	case "x86_64-pc-windows-msvc", "aarch64-pc-windows-msvc":
		var machine, platform string
		if target == "x86_64-pc-windows-msvc" {
			machine = "x64"
			platform = "windows_amd64"
		} else {
			machine = "arm64"
			platform = "windows_arm64"
		}
		libcDir, err := toolchain.LibcDir(platform)
		if err != nil {
			return err
		}

		stub, err := buildCtorStub(target, targetMachine, libcDir)
		if err != nil {
			return err
		}

		exeOut := outFile
		if !strings.HasSuffix(exeOut, ".exe") {
			exeOut += ".exe"
		}

		args = append(args, "/machine:"+machine)
		args = append(args, "/subsystem:console")
		args = append(args, "/out:"+exeOut)
		args = append(args, "/entry:mainCRTStartup") // <-- MinGW's C entry, calls your main
		args = append(args, "/libpath:"+libcDir)

		args = append(args, obj)
		args = append(args, stub)
		args = append(args, filepath.Join(libcDir, "crt2.o"))
		args = append(args, filepath.Join(libcDir, "libmingw32.a"))
		args = append(args, filepath.Join(libcDir, "libmingwex.a"))
		args = append(args, filepath.Join(libcDir, "libmsvcrt.a"))
		args = append(args, filepath.Join(libcDir, "libucrt.a"))
		args = append(args, filepath.Join(libcDir, "libkernel32.a"))
		args = append(args, filepath.Join(libcDir, "libgcc.a"))
		args = append(args, filepath.Join(libcDir, "libmingw32.a"))
		args = append(args, filepath.Join(libcDir, "crtend.o"))

	case "x86_64-apple-darwin", "aarch64-apple-darwin":
		platform := "darwin_amd64"
		arch := "x86_64"
		if target == "aarch64-apple-darwin" {
			platform = "darwin_arm64"
			arch = "arm64"
		}

		libcDir, err := toolchain.LibcDir(platform)
		if err != nil {
			return err
		}

		args = append(args, "-arch", arch)
		args = append(args, "-platform_version", "macos", "11.0", "11.0")
		args = append(args, "-L", libcDir)
		args = append(args, "-lSystem")
		args = append(args, "-o", outFile)

		args = append(args, obj)

	case "x86_64-pc-linux-gnu", "aarch64-pc-linux-gnu":
		platform := "linux_amd64"
		if target == "aarch64-pc-linux-gnu" {
			platform = "linux_arm64"
		}

		libcDir, err := toolchain.LibcDir(platform)
		if err != nil {
			return err
		}

		// Before obj: _start + init prologue from musl
		args = append(args,
			filepath.Join(libcDir, "crt1.o"),
			filepath.Join(libcDir, "crti.o"),
			"-no-pie",
			"-o", outFile,
		)

		args = append(args, obj)

	default:
		args = append(args, "-o", outFile)
		args = append(args, obj)
	}

	// After obj: libc + init epilogue
	if target == "x86_64-pc-linux-gnu" || target == "aarch64-pc-linux-gnu" {
		platform := "linux_amd64"
		if target == "aarch64-pc-linux-gnu" {
			platform = "linux_arm64"
		}
		libcDir, err := toolchain.LibcDir(platform)
		if err != nil {
			return err
		}
		args = append(args,
			filepath.Join(libcDir, "libc.a"),
			filepath.Join(libcDir, "libm.a"),
			filepath.Join(libcDir, "crtn.o"),
		)
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
