package linker

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
)

func Link(objFiles []string, outFile string, save bool, target string) error {
	lldPath := toolchain.LldPath()

	var args []string
	if runtime.GOOS == "windows" {
		args = append(args,
			"/out:"+outFile,
		)
		if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
			args = append(args, "/subsystem:console", "/libpath:"+toolchain.LldCacheDir()+"kernel32.lib")
		}
	} else {
		args = append(args, "-o", outFile)
	}
	args = append(args, objFiles...)

	err := toolchain.ExecuteCommand(lldPath, args...)

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
