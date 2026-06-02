package types

type NodeProgram struct {
	Statements []NodeStatement
	Filename   string
}

type NodeStatement struct {
	Exit      *NodeExit
	VarDecl   *NodeVarDecl
	VarAssign *NodeVarAssign
}

type NodeExit struct {
	Expr NodeExpression
	Pos  Position
}

type NodeExpression struct {
	BinExpr   *NodeBinExpr
	FuncCall  *NodeFuncCall
	IntLit    *int
	FloatLit  *float64
	BoolLit   *bool
	StringLit *string
	Ident     *string
	Pos       Position
}

type NodeFuncCall struct {
	Name string
	Args []NodeExpression
	Pos  Position
}

type NodeBinExpr struct {
	Left  NodeExpression
	Op    BinOp
	Right NodeExpression
	Pos   Position
}

type NodeVarDecl struct {
	Name  string
	Type  *string // nil if inferred
	Expr  NodeExpression
	Const bool
	Pos   Position
}

type NodeVarAssign struct {
	Name string
	Expr NodeExpression
	Pos  Position
}
