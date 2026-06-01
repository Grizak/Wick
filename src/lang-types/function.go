package lang_types

import "fmt"

// FunctionType represents a function signature
type FunctionType struct {
	ParamTypes []Type
	ReturnType Type
}

func (t *FunctionType) Name() string {
	paramNames := make([]string, len(t.ParamTypes))
	for i, p := range t.ParamTypes {
		paramNames[i] = p.Name()
	}
	return fmt.Sprintf("fn(%s) -> %s", fmt.Sprint(paramNames), t.ReturnType.Name())
}

func (t *FunctionType) LLVMType() string {
	// LLVM function types are written as: returnType(paramTypes)
	paramLLVMs := make([]string, len(t.ParamTypes))
	for i, p := range t.ParamTypes {
		paramLLVMs[i] = p.LLVMType()
	}
	return fmt.Sprintf("%s (%s)", t.ReturnType.LLVMType(), fmt.Sprint(paramLLVMs))
}

func (t *FunctionType) SizeBytes() int {
	return 0 // Functions don't have a runtime size
}

func (t *FunctionType) Equals(other Type) bool {
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
