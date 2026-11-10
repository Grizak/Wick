package linker

import (
	"fmt"
	"os"
)

func Link(objFiles []string, outFile string, save bool, target string, ldflags []string) error {
	var args []string
	if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
		var kernel32 string
		if target == "x86_64-pc-windows-msvc" {
			kernel32 = "kernel32.amd64.lib"
		} else if target == "aarch64-pc-windows-msvc" {
			kernel32 = "kernel32.arm64.lib"
		}

		// Windows args
	} else {
		// GNU args
	}
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
