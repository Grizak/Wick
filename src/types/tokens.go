package types

type TokenType string

const (
	TokenExit       TokenType = "exit"
	TokenOpenParen  TokenType = "("
	TokenCloseParen TokenType = ")"
	TokenIntLit     TokenType = "int_lit"
	TokenEOF        TokenType = "eof"
	TokenPlus       TokenType = "+"
	TokenStar       TokenType = "*"
)

type Token struct {
	Type  TokenType
	Value *string
	Pos Position
}

type Position struct {
	Line int
	Column int
	Index int
}