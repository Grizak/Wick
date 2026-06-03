package typesys

// StringType represents a pointer to a null-terminated string
type StringType struct{}

func (t *StringType) Name() string {
	return "string"
}

func (t *StringType) LLVMType() string {
	return "i8*"
}

func (t *StringType) SizeBytes() int {
	return 8 // pointer size
}

func (t *StringType) Equals(other Type) bool {
	_, ok := other.(*StringType)
	return ok
}
