package lang_types

import "fmt"

// ArrayType represents a fixed-size array
type ArrayType struct {
	ElementType Type
	Length      int
}

func (t *ArrayType) Name() string {
	return fmt.Sprintf("[%d]%s", t.Length, t.ElementType.Name())
}

func (t *ArrayType) LLVMType() string {
	return fmt.Sprintf("[%d x %s]", t.Length, t.ElementType.LLVMType())
}

func (t *ArrayType) SizeBytes() int {
	return t.Length * t.ElementType.SizeBytes()
}

func (t *ArrayType) Equals(other Type) bool {
	if arr, ok := other.(*ArrayType); ok {
		return t.Length == arr.Length && t.ElementType.Equals(arr.ElementType)
	}
	return false
}
