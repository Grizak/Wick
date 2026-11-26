package codegen

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Grizak/Wick/src/internal/types"
)

func (g *Generator) functionName(sourceName string, pos types.Position) string {
	moduleName := "main"
	if g.currentProgram != g.root {
		moduleName = filepath.Base(programPath(g.currentProgram))
		moduleName = strings.TrimSuffix(moduleName, filepath.Ext(moduleName))
	}
	moduleName = sanitizeSymbolComponent(moduleName)
	if moduleName == "" {
		moduleName = "module"
	}

	if sourceName == "" {
		return fmt.Sprintf("wick_%s_anonymous_%d_%d", moduleName, pos.Line, pos.Column)
	}
	return fmt.Sprintf("wick_%s_%s", moduleName, sanitizeSymbolComponent(sourceName))
}

func sanitizeSymbolComponent(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, name)
}
