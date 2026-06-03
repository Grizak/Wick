package assembler

import (
	"fmt"
	"os"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
)

func Assemble(asmFile, objFile, outFile string, save bool, idx int) error {
	llcPath := toolchain.LlcPath()

	err := toolchain.ExecuteCommand(llcPath, "-filetype=obj", asmFile, "-o", objFile)

	// If it failed, try to save the .ll file for debugging
	if !save && err == nil {
		if err := os.Remove(asmFile); err != nil {
			return err
		}
	} else {
		if err == nil {
			err := toolchain.ExecuteCommand(llcPath, "-filetype=asm", asmFile, "-o", objFile[:len(objFile)-11]+fmt.Sprint(idx)+".asm")
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
