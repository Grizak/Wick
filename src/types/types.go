package types

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

const (
	BinOpAdd BinOp = "+"
	BinOpSub BinOp = "-"
	BinOpMul BinOp = "*"
	BinOpDiv BinOp = "/"
)
