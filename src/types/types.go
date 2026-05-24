package types

type TokenType string

const (
	TokenExit       TokenType = "exit"
	TokenOpenParen  TokenType = "("
	TokenCloseParen TokenType = ")"
	TokenIntLit     TokenType = "int_lit"
	TokenEOF        TokenType = "eof"
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
	IntLit int
}

type Backend interface {
	Generate(program NodeProgram, outFile string) error
	Assemble(asmFile, objFile string, save bool) error
	Link(objFiles []string, outFile string, save bool) error
}
