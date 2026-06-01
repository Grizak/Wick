package lang_types

// FloatType represents a 64-bit floating point
type FloatType struct{}

func (t *FloatType) Name() string {
	return "float"
}

func (t *FloatType) LLVMType() string {
	return "double"
}

func (t *FloatType) SizeBytes() int {
	return 8
}

func (t *FloatType) Equals(other Type) bool {
	_, ok := other.(*FloatType)
	return ok
}
