package typesys

import "github.com/Grizak/Wick/src/internal/types"

// Int64Type represents a 64-bit signed integer
type Int64Type struct{}

func (t *Int64Type) Name() string {
	return "int"
}

func (t *Int64Type) LLVMType() string {
	return "i64"
}

func (t *Int64Type) SizeBytes() int {
	return 8
}

func (t *Int64Type) Equals(other types.Type) bool {
	_, ok := other.(*Int64Type)
	return ok
}
