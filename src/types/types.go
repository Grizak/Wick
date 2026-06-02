package types

type BinOp string

const (
	BinOpAdd BinOp = "+"
	BinOpSub BinOp = "-"
	BinOpMul BinOp = "*"
	BinOpDiv BinOp = "/"
)

type LexerResult struct {
	Token Token
	Err   *CompileError
}
