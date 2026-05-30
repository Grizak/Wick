package types

type TokenType string

const (
	TokenExit       TokenType = "exit"
	TokenOpenParen  TokenType = "("
	TokenCloseParen TokenType = ")"
	TokenIntLit     TokenType = "int_lit"
	TokenEOF        TokenType = "eof" // Not actual token name
	TokenPlus       TokenType = "+"
	TokenStar       TokenType = "*"
	TokenMinus      TokenType = "-"
	TokenFSlash     TokenType = "/"
	TokenIdent      TokenType = "ident"
	TokenConst      TokenType = "const"
	TokenLet        TokenType = "let"
	TokenColon      TokenType = ":"
	TokenEquals     TokenType = "="
)

type Token struct {
	Type  TokenType
	Value *string
	Pos   Position
}

type Position struct {
	Line   int
	Column int
	Index  int
}
