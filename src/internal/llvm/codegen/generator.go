package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/types"
)

type Symbol struct {
	llvmName    string
	isConst     bool
	symbolType  typesys.Type // The actual type of this symbol
	staticValue any          // Can be *int or *float64
}

type Generator struct {
	root        *ast.NodeProgram
	output      strings.Builder
	tmpCount    int
	symbols     map[string]Symbol
	fileName    string
	typeChecker *typesys.TypeChecker
}

func NewGenerator(root *ast.NodeProgram) *Generator {
	g := Generator{
		root:        root,
		tmpCount:    0,
		symbols:     make(map[string]Symbol),
		typeChecker: typesys.NewTypeChecker(),
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

func (g *Generator) generateExit(exit *ast.NodeExit) error {
	expr, err := g.generateExpression(exit.Expr)
	if err != nil {
		return err
	}

	// Infer the type of the exit expression
	exprType, err := g.typeChecker.InferType(&exit.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer exit expression type: %v", err), exit.Pos)
	}

	// If it's a constant int, we can use it directly
	if _, err := strconv.Atoi(expr); err == nil {
		// It's a numeric constant, use as-is but ensure it's in int32 range
		g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", expr))
		return nil
	}

	// Otherwise, convert to i32
	converted := g.tmpVar()
	if isFloatType(exprType) {
		// Convert double to i32
		g.writeLine(fmt.Sprintf("    %s = fptosi double %s to i32", converted, expr))
	} else {
		// Truncate i64 to i32
		g.writeLine(fmt.Sprintf("    %s = trunc i64 %s to i32", converted, expr))
	}
	g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", converted))
	return nil
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString(line)
	g.output.WriteString("\n")
}

func (g *Generator) generateExpression(expr ast.NodeExpression) (string, error) {
	if expr.IntLit != nil {
		return strconv.Itoa(*expr.IntLit), nil
	}

	if expr.FloatLit != nil {
		// Return the float literal as a double constant
		floatStr := strconv.FormatFloat(*expr.FloatLit, 'f', -1, 64)
		return floatStr, nil
	}

	if expr.BoolLit != nil {
		// Convert bool to i1 (0 or 1)
		val := 0
		if *expr.BoolLit {
			val = 1
		}
		return strconv.Itoa(val), nil
	}

	if expr.StringLit != nil {
		// For now, strings not fully supported in code generation
		return "", g.error("string literals not yet supported in code generation", expr.Pos)
	}

	// Try to constant-fold the expression at compile time
	if staticVal := g.computeStaticValue(expr); staticVal != nil {
		return strconv.Itoa(*staticVal), nil
	}

	if expr.BinExpr != nil {
		// Determine the types of left and right operands
		leftType, err := g.typeChecker.InferType(&expr.BinExpr.Left)
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer left operand type: %v", err), expr.BinExpr.Pos)
		}

		rightType, err := g.typeChecker.InferType(&expr.BinExpr.Right)
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer right operand type: %v", err), expr.BinExpr.Pos)
		}

		// Generate code for operands
		left, err := g.generateExpression(expr.BinExpr.Left)
		if err != nil {
			return "", err
		}
		right, err := g.generateExpression(expr.BinExpr.Right)
		if err != nil {
			return "", err
		}

		// Determine if we're doing float or int operations
		isFloat := isFloatType(leftType) || isFloatType(rightType)

		// Promote int to float if needed
		if isFloat {
			if !isFloatType(leftType) {
				promoted := g.tmpVar()
				g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", promoted, left))
				left = promoted
			}
			if !isFloatType(rightType) {
				promoted := g.tmpVar()
				g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", promoted, right))
				right = promoted
			}
		}

		result := g.tmpVar()

		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fadd double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = add i64 %s, %s", result, left, right))
			}
		case types.BinOpSub:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fsub double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = sub i64 %s, %s", result, left, right))
			}
		case types.BinOpMul:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fmul double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = mul i64 %s, %s", result, left, right))
			}
		case types.BinOpDiv:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fdiv double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = sdiv i64 %s, %s", result, left, right))
			}
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
		llvmType := sym.symbolType.LLVMType()
		g.writeLine(fmt.Sprintf("    %s = load %s, ptr %s", result, llvmType, sym.llvmName))
		return result, nil
	}

	if expr.FuncCall != nil {
		return "", g.error("function calls not yet supported in code generation", expr.Pos)
	}

	return "", g.error("unknown expression type", expr.Pos)
}

func (g *Generator) isStatic(expr ast.NodeExpression) bool {
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

func (g *Generator) foldBinExpr(expr *ast.NodeBinExpr) (int, error) {
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
			return 0, g.error("division by zero in constant expression", expr.Pos)
		}
		return left / right, nil
	default:
		return 0, g.error(fmt.Sprintf("unknown operator: %s", expr.Op), expr.Pos)
	}
}

func (g *Generator) foldExpression(expr ast.NodeExpression) (int, error) {
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
		// Type assert the staticValue to *int
		intVal, ok := sym.staticValue.(*int)
		if !ok {
			return 0, g.error(fmt.Sprintf("variable is not an integer constant: %s", *expr.Ident), expr.Pos)
		}
		return *intVal, nil
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

func (g *Generator) generateVarDecl(decl *ast.NodeVarDecl) error {
	if _, exists := g.symbols[decl.Name]; exists {
		return g.error(fmt.Sprintf("variable already declared: %s", decl.Name), decl.Pos)
	}

	// Infer or get the variable type
	var varType typesys.Type
	if decl.Type != nil {
		var err error
		varType, err = g.typeChecker.ParseTypeString(*decl.Type)
		if err != nil {
			return g.error(fmt.Sprintf("invalid type: %v", err), decl.Pos)
		}
	} else {
		// Infer type from expression
		var err error
		varType, err = g.typeChecker.InferType(&decl.Expr)
		if err != nil {
			return g.error(fmt.Sprintf("cannot infer type: %v", err), decl.Pos)
		}
	}

	expr, err := g.generateExpression(decl.Expr)
	if err != nil {
		return err
	}

	sym := Symbol{
		llvmName:   expr,
		isConst:    decl.Const,
		symbolType: varType,
	}

	// Try to compute and store the static value if this expression is constant
	if staticVal := g.computeStaticValue(decl.Expr); staticVal != nil {
		sym.staticValue = staticVal
	}

	if !decl.Const {
		ptr := fmt.Sprintf("%%var_%s", decl.Name)
		llvmType := varType.LLVMType()
		g.writeLine(fmt.Sprintf("    %s = alloca %s", ptr, llvmType))
		g.writeLine(fmt.Sprintf("    store %s %s, ptr %s", llvmType, expr, ptr))
		sym.llvmName = ptr
	}

	// Define the variable in the type checker environment for future type inference
	g.typeChecker.Environment().Define(decl.Name, varType)

	g.symbols[decl.Name] = sym
	return nil
}

func (g *Generator) generateVarAssign(assign *ast.NodeVarAssign) error {
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

	llvmType := sym.symbolType.LLVMType()
	g.writeLine(fmt.Sprintf("    store %s %s, ptr %s", llvmType, expr, sym.llvmName))
	return nil
}

func (g *Generator) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: g.fileName,
		Pos:  &pos,
		Msg:  msg,
	}
}

func (g *Generator) computeStaticValue(expr ast.NodeExpression) *int {
	if !g.isStatic(expr) {
		return nil
	}
	v, err := g.foldExpression(expr)
	if err != nil {
		return nil
	}
	return &v
}

// isFloatType checks if a type is a floating point type
func isFloatType(t typesys.Type) bool {
	_, ok := t.(*typesys.FloatType)
	return ok
}
