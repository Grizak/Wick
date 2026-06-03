package typesys

// Int32Type represents a 32-bit signed integer
type Int32Type struct{}

func (t *Int32Type) Name() string {
	return "i32"
}

func (t *Int32Type) LLVMType() string {
	return "i32"
}

func (t *Int32Type) SizeBytes() int {
	return 4
}

func (t *Int32Type) Equals(other Type) bool {
	_, ok := other.(*Int32Type)
	return ok
}
