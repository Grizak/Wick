package opt

import (
	"fmt"

	"tinygo.org/x/go-llvm"
)

var (
	opts llvm.PassBuilderOptions
)

func Initialize() {
	opts := llvm.NewPassBuilderOptions()
	opts.SetVerifyEach(true)
}

func Deinitialize() {
	opts.Dispose()
}

func Optimize(mod *llvm.Module, level int, targetMachine llvm.TargetMachine) error {
	return mod.RunPasses(fmt.Sprintf("default<O%d>", level), targetMachine, opts)
}
