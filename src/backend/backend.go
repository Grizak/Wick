package backend

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Grizak/Wick/src/tools"
)

func Assemble(asmFile, objFile, outFile string, save bool, idx int) error {
	llcPath := tools.LlcPath()

	err := tools.ExecuteCommand(llcPath, "-filetype=obj", asmFile, "-o", objFile)

	// If it failed, try to save the .ll file for debugging
	if !save && err == nil {
		if err := os.Remove(asmFile); err != nil {
			return err
		}
	} else {
		if err == nil {
			err := tools.ExecuteCommand(llcPath, "-filetype=asm", asmFile, "-o", objFile[:len(objFile)-11]+fmt.Sprint(idx)+".asm")
			if err != nil {
				return err
			}
		}
		if err := os.Rename(asmFile, outFile+fmt.Sprint(idx)+".ll"); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func Link(objFiles []string, outFile string, save bool, target string) error {
	lldPath := tools.LldPath()

	var args []string
	if runtime.GOOS == "windows" {
		args = append(args,
			"/out:"+outFile,
		)
		if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
			args = append(args, "/subsystem:console", "/libpath:"+tools.LldCacheDir()+"kernel32.lib")
		}
	} else {
		args = append(args, "-o", outFile)
	}
	args = append(args, objFiles...)

	err := tools.ExecuteCommand(lldPath, args...)

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
