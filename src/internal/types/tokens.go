package types

type TokenType string

const (
	TokenExit         TokenType = "exit"
	TokenOpenParen    TokenType = "("
	TokenCloseParen   TokenType = ")"
	TokenOpenBracket  TokenType = "["
	TokenCloseBracket TokenType = "]"
	TokenOpenBrace    TokenType = "{"
	TokenCloseBrace   TokenType = "}"
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
	TokenIf           TokenType = "if"
	TokenElse         TokenType = "else"
	TokenColon        TokenType = ":"
	TokenEquals       TokenType = "="
	TokenLt           TokenType = "<"
	TokenGt           TokenType = ">"
	TokenLtEq         TokenType = "<="
	TokenGtEq         TokenType = ">="
	TokenEqEq         TokenType = "=="
	TokenNotEq        TokenType = "!="
	TokenFor          TokenType = "for"
	TokenSemicolon    TokenType = ";"
	TokenBreak        TokenType = "break"
	TokenContinue     TokenType = "continue"
	TokenFunctionDecl TokenType = "fn"
	TokenReturnStmt   TokenType = "return"
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
