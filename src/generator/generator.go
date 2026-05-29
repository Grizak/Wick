package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Grizak/Wick/src/types"
)

type Generator struct {
	root         *types.NodeProgram
	output       strings.Builder
	nextRegister int
	nextLabelID  int
}

func NewGenerator(root *types.NodeProgram) *Generator {
	g := Generator{
		root:         root,
		nextRegister: 0,
		nextLabelID:  0,
	}

	return &g
}

func (g *Generator) Generate(fileName, target string) string {
	// Write some metadata about the file (based on target)
	g.writeLine(fmt.Sprintf(`target triple = "%s"`, target))
	g.writeLine(`target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"`)

	g.writeLine(fmt.Sprintf(`source_filename = "%s"`, fileName))

	// Write LLVM IR module header
	g.writeLine("")
	g.writeLine(fmt.Sprintf(`define void @%s() {`, entryPoint(target)))
	g.writeLine(`entry:`)

	exitCount := 0

	for _, statement := range g.root.Statements {
		if statement.Exit != nil {
			exitCount++
			g.generateExit(statement.Exit)
		}
	}

	g.writeLine(`  ret void`)
	g.writeLine(`}`)
	g.writeLine("")

	if exitCount > 0 {
		g.writeLine(`define void @exit(i32 %code) {`)
		g.writeLine("entry:")
		g.writeLine("    %code64 = sext i32 %code to i64")
		g.generateExitSyscall(target)
		g.writeLine("    unreachable")
		g.writeLine("}")

		if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
			g.writeLine(`declare void @ExitProcess(i32)`)
		}
	}

	return g.output.String()
}

func (g *Generator) generateExit(exit *types.NodeExit) {
	expr := g.generateExpression(exit.Expr)
	g.writeLine(fmt.Sprintf(`  call void @exit(i32 %s)`, expr))
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString(line)
	g.output.WriteString("\n")
}

func (g *Generator) getRegister() string {
	reg := fmt.Sprintf("%%%d", g.nextRegister)
	g.nextRegister++
	return reg
}

func (g *Generator) generateExpression(expr types.NodeExpression) string {
	if expr.IntLit != nil {
		return strconv.Itoa(*expr.IntLit)
	}

	if expr.BinExpr != nil {
		// Constant folding - if both sides are static, evaluate at compile time
		if isStatic(expr.BinExpr.Left) && isStatic(expr.BinExpr.Right) {
			return strconv.Itoa(foldBinExpr(expr.BinExpr))
		}

		// Generate LLVM IR for runtime arithmetic
		left := g.generateExpression(expr.BinExpr.Left)
		right := g.generateExpression(expr.BinExpr.Right)

		reg := g.getRegister()
		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			g.writeLine(fmt.Sprintf(`  %s = add i32 %s, %s`, reg[1:], left, right))
		default:
			panic("unknown operator")
		}
		return reg[1:] // Remove the % prefix since it's already in the register
	}

	panic("unknown expression type")
}

func isStatic(expr types.NodeExpression) bool {
	if expr.IntLit != nil {
		return true
	}
	if expr.BinExpr != nil {
		return isStatic(expr.BinExpr.Left) && isStatic(expr.BinExpr.Right)
	}
	return false
}

func foldBinExpr(expr *types.NodeBinExpr) int {
	left := foldExpression(expr.Left)
	right := foldExpression(expr.Right)

	switch expr.Op {
	case types.BinOpAdd:
		return left + right
	default:
		panic("unknown operator")
	}
}

func foldExpression(expr types.NodeExpression) int {
	if expr.IntLit != nil {
		return *expr.IntLit
	}
	if expr.BinExpr != nil {
		return foldBinExpr(expr.BinExpr)
	}
	panic("unknown expression type")
}

func (g *Generator) generateExitSyscall(target string) {
	switch target {
	case "x86_64-pc-linux-gnu":
		g.writeLine("    call void asm sideeffect \"syscall\", \"{rax},{rdi}\" (i64 60, i64 %code64)")
	case "aarch64-pc-linux-gnu":
		g.writeLine("    call void asm sideeffect \"svc #0\", \"{x8},{x0}\" (i64 93, i64 %code64)")
	case "x86_64-apple-darwin":
		g.writeLine("    call void asm sideeffect \"syscall\", \"{rax},{rdi}\" (i64 0x2000001, i64 %code64)")
	case "aarch64-apple-darwin":
		g.writeLine("    call void asm sideeffect \"svc #0x80\", \"{x16},{x0}\" (i64 1, i64 %code64)")
	case "x86_64-pc-windows-msvc", "aarch64-pc-windows-msvc":
		g.writeLine("    call void @ExitProcess(i32 %code)")
	default:
		panic(fmt.Sprintf("unsupported target: %s", target))
	}
}

func entryPoint(target string) string {
	switch target {
	case "x86_64-apple-darwin", "aarch64-apple-darwin":
		return "_main"
	case "x86_64-pc-windows-msvc", "aarch64-pc-windows-msvc":
		return "mainCRTStartup"
	default:
		return "_start"
	}
}
