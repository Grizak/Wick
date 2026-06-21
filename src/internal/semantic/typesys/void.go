package typesys

import "github.com/Grizak/Wick/src/internal/types"

type VoidType struct{}

func (VoidType) Name() string {
	return "void"
}

func (VoidType) LLVMType() string {
	return "void"
}

func (VoidType) SizeBytes() int {
	return 0
}

func (VoidType) Equals(other types.Type) bool {
	switch other.(type) {
	case VoidType, *VoidType:
		return true
	default:
		return false
	}
}
