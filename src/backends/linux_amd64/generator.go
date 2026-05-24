package linux_amd64

import (
	"strconv"
	"strings"

	"github.com/Grizak/Wick/src/types"
)

type Generator struct {
	root   *types.NodeProgram
	output strings.Builder
}

func NewGenerator(root *types.NodeProgram) *Generator {
	g := Generator{
		root: root,
	}

	return &g
}

func (g *Generator) Generate() string {
	g.output.WriteString("section .text\n")
	g.output.WriteString("global _start\n")
	g.output.WriteString("_start:\n")
	for _, statement := range g.root.Statements {
		if statement.Exit != nil {
			g.generateExit(statement.Exit)
		}
	}
	return g.output.String()
}

func (g *Generator) generateExit(exit *types.NodeExit) {
	g.writeLine("mov rax, 60")
	g.writeLine("mov rdi, " + strconv.Itoa(exit.Expr.IntLit))
	g.writeLine("syscall")
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString("    ")
	g.output.WriteString(line)
	g.output.WriteString("\n")
}
