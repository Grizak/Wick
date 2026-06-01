package lang_types

// BoolType represents a 1-bit boolean
type BoolType struct{}

func (t *BoolType) Name() string {
	return "bool"
}

func (t *BoolType) LLVMType() string {
	return "i1"
}

func (t *BoolType) SizeBytes() int {
	return 1
}

func (t *BoolType) Equals(other Type) bool {
	_, ok := other.(*BoolType)
	return ok
}