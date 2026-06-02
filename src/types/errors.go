package types

import "fmt"

type CompileError struct {
	File string
	Pos  *Position
	Msg  string
}

func (e *CompileError) Error() string {
	if e.Pos == nil {
		return fmt.Sprintf("%s: error: %s", e.File, e.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: error: %s", e.File, e.Pos.Line, e.Pos.Column, e.Msg)
}
