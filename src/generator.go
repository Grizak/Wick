package main

import (
	"strconv"
	"strings"
)

type Generator struct {
	root   NodeProgram
	output strings.Builder
}

func NewGenerator(root NodeProgram) *Generator {
	g := Generator{
		root: root,
	}

	return &g
}

func (g *Generator) Generate() string {
	g.output.WriteString("section .text\n")
	g.output.WriteString("global _start\n")
	g.output.WriteString("_start:\n")
	for _, statement := range g.root.statements {
		if statement.Exit != nil {
			g.generateExit(statement.Exit)
		}
	}
	return g.output.String()
}

func (g *Generator) generateExit(exit *NodeExit) {
	g.writeLine("mov rax, 60")
	g.writeLine("mov rdi, " + strconv.Itoa(exit.Expr.IntLit))
	g.writeLine("syscall")
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString("    ")
	g.output.WriteString(line)
	g.output.WriteString("\n")
}
