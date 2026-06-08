package opt

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
)

func Optimize(inputFile, outputFile string, level int) error {
	optPath := toolchain.OptPath()

	if err := toolchain.ExecuteCommand(optPath, fmt.Sprintf("-O%d", level), "-S", inputFile, "-o", outputFile); err != nil {
		return err
	}

	return nil
}
