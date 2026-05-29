package backend

import (
	"os"

	"github.com/Grizak/Wick/src/tools"
)

func Assemble(asmFile, objFile string, save bool) error {
	llcPath := tools.LlcPath()

	err := tools.ExecuteCommand(llcPath, "-filetype=obj", asmFile, "-o", objFile)
	if err != nil {
		return err
	}

	if !save {
		if err := os.Remove(asmFile); err != nil {
			return err
		}
	}

	return nil
}

func Link(objFiles []string, outFile string, save bool, target string) error {
	lldPath := tools.LldPath()

	var args []string
	if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
		args = append(args, "/subsystem:console", "kernel32.lib", "/out:"+outFile)
	} else {
		args = append(args, "-o", outFile)
	}
	args = append(args, objFiles...)

	err := tools.ExecuteCommand(lldPath, args...)
	if err != nil {
		return err
	}

	if !save {
		for _, objFile := range objFiles {
			if err := os.Remove(objFile); err != nil {
				return err
			}
		}
	}

	return nil
}
