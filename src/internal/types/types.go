package types

type BinOp string

const (
	BinOpAdd   BinOp = "+"
	BinOpSub   BinOp = "-"
	BinOpMul   BinOp = "*"
	BinOpDiv   BinOp = "/"
	BinOpEq    BinOp = "=="
	BinOpNotEq BinOp = "!="
	BinOpLt    BinOp = "<"
	BinOpGt    BinOp = ">"
	BinOpLtEq  BinOp = "<="
	BinOpGtEq  BinOp = ">="
)

type LexerResult struct {
	Token Token
	Err   *CompileError
}

// Type represents all possible types in Wick
type Type interface {
	Name() string
	LLVMType() string
	SizeBytes() int
	Equals(other Type) bool
}
