package linker

import (
	"os"
	"path/filepath"
	"tinygo.org/x/go-llvm"
	"github.com/Grizak/Wick/src/internal/llvm/assembler"
)

func buildCtorStub(targetTriple string, targetMachine llvm.TargetMachine, libcDir string) (string, error) {
	outBase := filepath.Join(libcDir, "ctor")
	outObj := outBase + ".o"
	if _, err := os.Stat(outObj); err == nil {
		return outObj, nil
	}

	ctx := llvm.NewContext()
	defer ctx.Dispose()

	mod := ctx.NewModule("ctor_stub")
	// don't Dispose mod here — Assemble() needs it, it takes *llvm.Module
	mod.SetTarget(targetTriple)

	// TargetMachine.CreateTargetData(), not .TargetData
	td := targetMachine.CreateTargetData()
	defer td.Dispose()
	mod.SetDataLayout(td.String())

	// [2 x i64] { -1, 0 }
	i64 := ctx.Int64Type()
	arrTy := llvm.ArrayType(i64, 2)
	vals := []llvm.Value{
		llvm.ConstInt(i64, ^uint64(0), true), // -1
		llvm.ConstInt(i64, 0, false),         // 0
	}
	init := llvm.ConstArray(arrTy, vals)

	ctor := llvm.AddGlobal(mod, arrTy, "__CTOR_LIST__")
	ctor.SetInitializer(init)
	ctor.SetLinkage(llvm.ExternalLinkage)
	ctor.SetAlignment(8)
	ctor.SetGlobalConstant(false)

	dtor := llvm.AddGlobal(mod, arrTy, "__DTOR_LIST__")
	dtor.SetInitializer(init)
	dtor.SetLinkage(llvm.ExternalLinkage)
	dtor.SetAlignment(8)
	dtor.SetGlobalConstant(false)

	_, err := assembler.Assemble(&mod, outBase, false, targetMachine)
	mod.Dispose()
	if err != nil {
		return "", err
	}
	return outObj, nil
}
