package linux_amd64

import (
	"os"

	"github.com/Grizak/Wick/src/tools"
	"github.com/Grizak/Wick/src/types"
)

type LinuxAMD64Backend struct{}

// Implement Generate in [backend.Backend]
func (b *LinuxAMD64Backend) Generate(program types.NodeProgram, outFile string) error {
	generator := NewGenerator(&program)
	asm := generator.Generate()

	// Write the generated assembly code to the output file
	err := os.WriteFile(outFile, []byte(asm), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *LinuxAMD64Backend) Assemble(asmFile, objFile string, save bool) error {
	// Implementation for assembling the assembly code into an object file
	nasmPath := tools.NasmPath()

	cmd := nasmPath + " -f elf64 " + asmFile + " -o " + objFile
	err := tools.ExecuteCommand(cmd)
	if err != nil {
		return err
	}
	// Remove the assembly file after successful assembly
	if !save {
		err = os.Remove(asmFile)
		if err != nil {
			return err
		}
	}
	// Return nil if everything succeeded
	return nil
}

func (b *LinuxAMD64Backend) Link(objFiles []string, outFile string, save bool) error {
	// Implementation for linking object files
	lldPath := tools.LldPath()
	cmd := lldPath + " -flavor gnu -o " + outFile
	for _, objFile := range objFiles {
		cmd += " " + objFile
	}
	err := tools.ExecuteCommand(cmd)
	if err != nil {
		return err
	}
	// Remove the object files after successful linking
	if !save {
		for _, objFile := range objFiles {
			err = os.Remove(objFile)
			if err != nil {
				return err
			}
		}
	}
	// Return nil if everything succeeded
	return nil
}

func NewLinuxAMD64Backend() types.Backend {
	return &LinuxAMD64Backend{}
}
