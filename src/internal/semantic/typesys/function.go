package typesys

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

// FunctionType represents a function signature
type FunctionType struct {
	ParamTypes []types.Type
	ReturnType types.Type
	Defaults   []*ast.NodeExpression // parallel to ParamTypes; nil = required, no default. Deliberately NOT compared in Equals() — two functions with the same signature but different defaults are still the same type for assignability purposes.
}

func (t *FunctionType) Name() string {
	paramNames := make([]string, len(t.ParamTypes))
	for i, p := range t.ParamTypes {
		paramNames[i] = p.Name()
	}
	return fmt.Sprintf("fn(%s) -> %s", strings.Join(paramNames, ",  "), t.ReturnType.Name())
}

func (t *FunctionType) LLVMType() string {
	// LLVM function types are written as: returnType(paramTypes)
	paramLLVMs := make([]string, len(t.ParamTypes))
	for i, p := range t.ParamTypes {
		paramLLVMs[i] = p.LLVMType()
	}
	return fmt.Sprintf("%s (%s)", t.ReturnType.LLVMType(), strings.Join(paramLLVMs, ", "))
}

func (t *FunctionType) SizeBytes() int {
	return 0 // Functions don't have a runtime size
}

func (t *FunctionType) Equals(other types.Type) bool {
	if fn, ok := other.(*FunctionType); ok {
		if len(t.ParamTypes) != len(fn.ParamTypes) {
			return false
		}
		for i := range t.ParamTypes {
			if !t.ParamTypes[i].Equals(fn.ParamTypes[i]) {
				return false
			}
		}
		return t.ReturnType.Equals(fn.ReturnType)
	}
	return false
}

func (t *FunctionType) RequiredArgCount() int {
	for i, d := range t.Defaults {
		if d != nil {
			return i
		}
	}
	return len(t.Defaults)
}
