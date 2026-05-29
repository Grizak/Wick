package types

type TokenType string

const (
	TokenExit       TokenType = "exit"
	TokenOpenParen  TokenType = "("
	TokenCloseParen TokenType = ")"
	TokenIntLit     TokenType = "int_lit"
	TokenEOF        TokenType = "eof"
	TokenPlus       TokenType = "+"
)

type Token struct {
	Type  TokenType
	Value *string
	Line  int
}

type NodeProgram struct {
	Statements []NodeStatement
}

type NodeStatement struct {
	Exit *NodeExit
}

type NodeExit struct {
	Expr NodeExpression
}

type NodeExpression struct {
	BinExpr *NodeBinExpr
	IntLit  *int
}

type NodeBinExpr struct {
	Left  NodeExpression
	Op    BinOp
	Right NodeExpression
}

type BinOp string

const BinOpAdd BinOp = "+"
