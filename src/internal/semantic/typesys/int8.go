package typesys

import "github.com/Grizak/Wick/src/internal/types"

// Int8Type represents a 32-bit signed integer
type Int8Type struct{}

func (t *Int8Type) Name() string {
	return "i8"
}

func (t *Int8Type) LLVMType() string {
	return "i8"
}

func (t *Int8Type) SizeBytes() int {
	return 1
}

func (t *Int8Type) Equals(other types.Type) bool {
	_, ok := other.(*Int8Type)
	return ok
}
