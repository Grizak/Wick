package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Grizak/Wick/src/types"
)

type Symbol struct {
	llvmName    string
	isConst     bool
	varType     string
	staticValue *int
}

type Generator struct {
	root     *types.NodeProgram
	output   strings.Builder
	tmpCount int
	symbols  map[string]Symbol
	fileName string
}

func NewGenerator(root *types.NodeProgram) *Generator {
	g := Generator{
		root:     root,
		tmpCount: 0,
		symbols:  make(map[string]Symbol),
	}

	return &g
}

func (g *Generator) Generate(fileName, target string) (string, error) {
	g.fileName = fileName
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
			if err := g.generateExit(statement.Exit); err != nil {
				return "", err
			}
		}
		if statement.VarDecl != nil {
			if err := g.generateVarDecl(statement.VarDecl); err != nil {
				return "", err
			}
		}
		if statement.VarAssign != nil {
			if err := g.generateVarAssign(statement.VarAssign); err != nil {
				return "", err
			}
		}
	}

	g.writeLine(`    ret void`)
	g.writeLine(`}`)
	g.writeLine("")

	if exitCount > 0 {
		g.writeLine(`define void @exit(i32 %code) {`)
		g.writeLine("entry:")
		g.writeLine("    %code64 = sext i32 %code to i64")
		if err := g.generateExitSyscall(target); err != nil {
			return "", err
		}
		g.writeLine("    unreachable")
		g.writeLine("}")

		if target == "x86_64-pc-windows-msvc" || target == "aarch64-pc-windows-msvc" {
			g.writeLine(`declare void @ExitProcess(i32)`)
		}
	}

	return g.output.String(), nil
}

func (g *Generator) generateExit(exit *types.NodeExit) error {
	expr, err := g.generateExpression(exit.Expr)
	if err != nil {
		return err
	}

	// If it's a constant, no truncation needed
	if _, err := strconv.Atoi(expr); err == nil {
		g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", expr))
		return nil
	}

	// Otherwise truncate from i64 to i32
	truncated := g.tmpVar()
	g.writeLine(fmt.Sprintf("    %s = trunc i64 %s to i32", truncated, expr))
	g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", truncated))
	return nil
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString(line)
	g.output.WriteString("\n")
}

func (g *Generator) generateExpression(expr types.NodeExpression) (string, error) {
	if expr.IntLit != nil {
		return strconv.Itoa(*expr.IntLit), nil
	}

	if expr.BinExpr != nil {
		if g.isStatic(expr.BinExpr.Left) && g.isStatic(expr.BinExpr.Right) {
			folded, err := g.foldBinExpr(expr.BinExpr)
			return strconv.Itoa(folded), err
		}

		left, err := g.generateExpression(expr.BinExpr.Left)
		if err != nil {
			return "", err
		}
		right, err := g.generateExpression(expr.BinExpr.Right)
		if err != nil {
			return "", err
		}
		result := g.tmpVar()

		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			g.writeLine(fmt.Sprintf("    %s = add i64 %s, %s", result, left, right))
		case types.BinOpSub:
			g.writeLine(fmt.Sprintf("    %s = sub i64 %s, %s", result, left, right))
		case types.BinOpMul:
			g.writeLine(fmt.Sprintf("    %s = mul i64 %s, %s", result, left, right))
		case types.BinOpDiv:
			g.writeLine(fmt.Sprintf("    %s = sdiv i64 %s, %s", result, left, right))
		default:
			return "", g.error(fmt.Sprintf("unknown operator: %s", expr.BinExpr.Op), expr.BinExpr.Pos)
		}

		return result, nil
	}

	if expr.Ident != nil {
		sym, exists := g.symbols[*expr.Ident]
		if !exists {
			return "", g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.isConst {
			// Constants are just their value directly
			return sym.llvmName, nil
		}
		// Mutable variables need a load
		result := g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = load i64, ptr %s", result, sym.llvmName))
		return result, nil
	}

	return "", g.error("unknown expression type", expr.Pos)
}

func (g *Generator) isStatic(expr types.NodeExpression) bool {
	if expr.IntLit != nil {
		return true
	}
	if expr.Ident != nil {
		sym, exists := g.symbols[*expr.Ident]
		if !exists {
			return false
		}
		return sym.staticValue != nil
	}
	if expr.BinExpr != nil {
		return g.isStatic(expr.BinExpr.Left) && g.isStatic(expr.BinExpr.Right)
	}
	return false
}

func (g *Generator) foldBinExpr(expr *types.NodeBinExpr) (int, error) {
	left, err := g.foldExpression(expr.Left)
	if err != nil {
		return 0, err
	}
	right, err := g.foldExpression(expr.Right)
	if err != nil {
		return 0, err
	}

	switch expr.Op {
	case types.BinOpAdd:
		return left + right, nil
	case types.BinOpSub:
		return left - right, nil
	case types.BinOpMul:
		return left * right, nil
	case types.BinOpDiv:
		if right == 0 {
			return 0, &types.CompileError{
				File: g.fileName,
				Pos:  expr.Pos,
				Msg:  "Divide by zero",
			}
		}
		return left / right, nil
	default:
		return 0, g.error(fmt.Sprintf("unknown operator: %s", expr.Op), expr.Pos)
	}
}

func (g *Generator) foldExpression(expr types.NodeExpression) (int, error) {
	if expr.IntLit != nil {
		return *expr.IntLit, nil
	}
	if expr.Ident != nil {
		sym, exists := g.symbols[*expr.Ident]
		if !exists {
			return 0, g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.staticValue == nil {
			return 0, g.error(fmt.Sprintf("variable is not statically known: %s", *expr.Ident), expr.Pos)
		}
		return *sym.staticValue, nil
	}
	if expr.BinExpr != nil {
		return g.foldBinExpr(expr.BinExpr)
	}
	return 0, g.error("unknown expression type", expr.Pos)
}

func (g *Generator) generateExitSyscall(target string) error {
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
		return g.error(fmt.Sprintf("unsupported target: %s", target), types.Position{})
	}
	return nil
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

func (g *Generator) generateVarDecl(decl *types.NodeVarDecl) error {
	if _, exists := g.symbols[decl.Name]; exists {
		return g.error(fmt.Sprintf("variable already declared: %s", decl.Name), decl.Pos)
	}

	expr, err := g.generateExpression(decl.Expr)
	if err != nil {
		return err
	}

	sym := Symbol{
		llvmName:    expr,
		isConst:     decl.Const,
		varType:     "i64",
		staticValue: g.computeStaticValue(decl.Expr),
	}

	if !decl.Const {
		ptr := fmt.Sprintf("%%var_%s", decl.Name)
		g.writeLine(fmt.Sprintf("    %s = alloca i64", ptr))
		g.writeLine(fmt.Sprintf("    store i64 %s, ptr %s", expr, ptr))
		sym.llvmName = ptr
	}

	g.symbols[decl.Name] = sym
	return nil
}

func (g *Generator) generateVarAssign(assign *types.NodeVarAssign) error {
	sym, exists := g.symbols[assign.Name]
	if !exists {
		return g.error(fmt.Sprintf("undeclared variable: %s", assign.Name), assign.Pos)
	}
	if sym.isConst {
		return g.error(fmt.Sprintf("cannot reassign const variable: %s", assign.Name), assign.Pos)
	}

	expr, err := g.generateExpression(assign.Expr)
	if err != nil {
		return err
	}

	sym.staticValue = g.computeStaticValue(assign.Expr)
	g.symbols[assign.Name] = sym

	g.writeLine(fmt.Sprintf("  store i64 %s, ptr %s", expr, sym.llvmName))
	return nil
}

func (g *Generator) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: g.fileName,
		Pos:  pos,
		Msg:  msg,
	}
}

func (g *Generator) computeStaticValue(expr types.NodeExpression) *int {
	if !g.isStatic(expr) {
		return nil
	}
	v, err := g.foldExpression(expr)
	if err != nil {
		return nil
	}
	return &v
}
