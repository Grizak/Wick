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
	tmpCount     int
}

func NewGenerator(root *types.NodeProgram) *Generator {
	g := Generator{
		root:         root,
		nextRegister: 0,
		nextLabelID:  0,
		tmpCount:     0,
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

	// If it's a constant, no truncation needed
	if _, err := strconv.Atoi(expr); err == nil {
		g.writeLine(fmt.Sprintf("  call void @exit(i32 %s)", expr))
		return
	}

	// Otherwise truncate from i64 to i32
	truncated := g.tmpVar()
	g.writeLine(fmt.Sprintf("  %s = trunc i64 %s to i32", truncated, expr))
	g.writeLine(fmt.Sprintf("  call void @exit(i32 %s)", truncated))
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
		if isStatic(expr.BinExpr.Left) && isStatic(expr.BinExpr.Right) {
			return strconv.Itoa(foldBinExpr(expr.BinExpr))
		}

		left := g.generateExpression(expr.BinExpr.Left)
		right := g.generateExpression(expr.BinExpr.Right)
		result := g.tmpVar()

		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			g.writeLine(fmt.Sprintf("  %s = add i64 %s, %s", result, left, right))
		case types.BinOpSub:
			g.writeLine(fmt.Sprintf("  %s = sub i64 %s, %s", result, left, right))
		case types.BinOpMul:
			g.writeLine(fmt.Sprintf("  %s = mul i64 %s, %s", result, left, right))
		case types.BinOpDiv:
			g.writeLine(fmt.Sprintf("  %s = sdiv i64 %s, %s", result, left, right))
		default:
			panic(fmt.Sprintf("unknown operator: %s", expr.BinExpr.Op))
		}

		return result
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
	case types.BinOpSub:
		return left - right
	case types.BinOpMul:
		return left * right
	case types.BinOpDiv:
		if right == 0 {
			panic("division by zero")
		}
		return left / right
	default:
		panic(fmt.Sprintf("unknown operator: %s", expr.Op))
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

func (g *Generator) tmpVar() string {
	g.tmpCount++
	return fmt.Sprintf("%%tmp%d", g.tmpCount)
}
