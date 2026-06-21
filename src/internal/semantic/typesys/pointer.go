package typesys

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/types"
)

// PointerType represents a pointer to another type
type PointerType struct {
	PointsTo types.Type
}

func (t *PointerType) Name() string {
	return fmt.Sprintf("*%s", t.PointsTo.Name())
}

func (t *PointerType) LLVMType() string {
	return t.PointsTo.LLVMType() + "*"
}

func (t *PointerType) SizeBytes() int {
	return 8
}

func (t *PointerType) Equals(other types.Type) bool {
	if ptr, ok := other.(*PointerType); ok {
		return t.PointsTo.Equals(ptr.PointsTo)
	}
	return false
}
