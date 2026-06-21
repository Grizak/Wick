package codegen

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

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
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return "", g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.isConst {
			// Constants are just their value directly
			return sym.llvmName, nil
		}
		// Mutable variables need a load
		result := g.tmpVar()
		llvmType := llvmStorageType(sym.symbolType)
		g.writeLine(fmt.Sprintf("    %s = load %s, ptr %s", result, llvmType, sym.llvmName))
		return result, nil
	}

	if expr.FuncCall != nil {
		return g.generateFuncCall(expr.FuncCall)
	}

	if expr.FuncDecl != nil {
		fnType, err := g.funcDeclType(expr.FuncDecl)
		if err != nil {
			return "", g.error(err.Error(), expr.Pos)
		}
		name := fmt.Sprintf("@func_%d", g.funcCount)
		g.funcCount++
		if err := g.emitFuncBody(expr.FuncDecl, fnType, name); err != nil {
			return "", err
		}
		return name, nil
	}

	return "", g.error("unknown expression type", expr.Pos)
}

func (g *Generator) isStatic(expr ast.NodeExpression) bool {
	if expr.IntLit != nil {
		return true
	}
	if expr.FloatLit != nil {
		return true
	}
	if expr.Ident != nil {
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return false
		}
		return sym.isConst && sym.staticValue != nil
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
		sym, exists := g.scope.lookup(*expr.Ident)
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
