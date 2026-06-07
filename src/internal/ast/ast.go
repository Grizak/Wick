package ast

import "github.com/Grizak/Wick/src/internal/types"

type NodeProgram struct {
	Statements []NodeStatement
	Filename   string
}

type NodeStatement struct {
	Exit      *NodeExit
	VarDecl   *NodeVarDecl
	VarAssign *NodeVarAssign
	Block     *NodeBlock
	If        *NodeIf
	For       *NodeFor
	Break     *NodeBreak
	Continue  *NodeContinue
}

type NodeExit struct {
	Expr NodeExpression
	Pos  types.Position
}

type NodeExpression struct {
	BinExpr   *NodeBinExpr
	FuncCall  *NodeFuncCall
	IntLit    *int
	FloatLit  *float64
	BoolLit   *bool
	StringLit *string
	Ident     *string
	Pos       types.Position
}

type NodeFuncCall struct {
	Name string
	Args []NodeExpression
	Pos  types.Position
}

type NodeBinExpr struct {
	Left  NodeExpression
	Op    types.BinOp
	Right NodeExpression
	Pos   types.Position
}

type NodeVarDecl struct {
	Name  string
	Type  *string // nil if inferred
	Expr  NodeExpression
	Const bool
	Pos   types.Position
}

type NodeVarAssign struct {
	Name string
	Expr NodeExpression
	Pos  types.Position
}

type NodeBlock struct {
	Statements []NodeStatement
	Pos        types.Position
}

type NodeIf struct {
	Condition NodeExpression
	Then      NodeBlock
	Else      *NodeBlock // Nil if no else branch
	Pos       types.Position
}

type NodeFor struct {
	Init      *NodeStatement  // nil for while-style and infinite
	Condition *NodeExpression // nil for infinite loop
	Post      *NodeStatement  // nil for while-style and infinite
	Body      NodeBlock
	Pos       types.Position
}

type NodeBreak struct {
	Pos types.Position
}

type NodeContinue struct {
	Pos types.Position
}
