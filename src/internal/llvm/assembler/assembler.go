package assembler

import (
	"fmt"
	"os"

	"tinygo.org/x/go-llvm"
)

func Assemble(mod *llvm.Module, outFile string, save bool, idx int, targetMachine llvm.TargetMachine) (string, error) {
	buf, err := targetMachine.EmitToMemoryBuffer(*mod, llvm.ObjectFile)
	if err != nil {
		return "", err
	}

	bytes := buf.Bytes()
	objFile := outFile + fmt.Sprint(idx) + ".o"
	if err := os.WriteFile(objFile, bytes, 0644); err != nil {
		return "", err
	}

	if save {
		buf, err := targetMachine.EmitToMemoryBuffer(*mod, llvm.AssemblyFile)
		if err != nil {
			return "", err
		}

		bytes := buf.Bytes()
		objFile := outFile + fmt.Sprint(idx) + ".s"
		if err := os.WriteFile(objFile, bytes, 0644); err != nil {
			return "", err
		}

		ir := mod.String()
		if err := os.WriteFile(outFile+fmt.Sprint(idx)+".ll", []byte(ir), 0644); err != nil {
			return "", err
		}
	}

	return objFile, nil
}
