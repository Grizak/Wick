package types

type TokenType string

const (
	TokenExit         TokenType = "exit"
	TokenOpenParen    TokenType = "("
	TokenCloseParen   TokenType = ")"
	TokenOpenBracket  TokenType = "["
	TokenCloseBracket TokenType = "]"
	TokenComma        TokenType = ","
	TokenIntLit       TokenType = "int_lit"
	TokenFloatLit     TokenType = "float_lit"
	TokenStringLit    TokenType = "string_lit"
	TokenTrue         TokenType = "true"
	TokenFalse        TokenType = "false"
	TokenEOF          TokenType = "eof" // Not actual token name
	TokenPlus         TokenType = "+"
	TokenStar         TokenType = "*"
	TokenMinus        TokenType = "-"
	TokenFSlash       TokenType = "/"
	TokenIdent        TokenType = "ident"
	TokenConst        TokenType = "const"
	TokenLet          TokenType = "let"
	TokenColon        TokenType = ":"
	TokenEquals       TokenType = "="
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
