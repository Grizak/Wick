package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
	"tinygo.org/x/go-llvm"
)

func (g *Generator) generateExpression(expr ast.NodeExpression) (llvm.Value, error) {
	switch {
	case expr.IntLit != nil:
		return llvm.ConstInt(g.context.Int64Type(), uint64(*expr.IntLit), true), nil
	case expr.FloatLit != nil:
		return llvm.ConstFloat(g.context.DoubleType(), *expr.FloatLit), nil
	case expr.BoolLit != nil:
		value := uint64(0)
		if *expr.BoolLit {
			value = 1
		}
		return llvm.ConstInt(g.context.Int1Type(), value, false), nil
	case expr.StringLit != nil:
		return llvm.Value{}, g.error("string literals not yet supported in code generation", expr.Pos)
	}

	if staticValue := g.computeStaticValue(expr); staticValue != nil {
		exprType, err := g.typeChecker.InferType(&expr)
		if err != nil {
			return llvm.Value{}, err
		}
		return llvm.ConstInt(g.llvmType(exprType), uint64(*staticValue), true), nil
	}

	switch {
	case expr.BinExpr != nil:
		leftType, err := g.typeChecker.InferType(&expr.BinExpr.Left)
		if err != nil {
			return llvm.Value{}, g.error(fmt.Sprintf("cannot infer left operand type: %v", err), expr.BinExpr.Pos)
		}
		rightType, err := g.typeChecker.InferType(&expr.BinExpr.Right)
		if err != nil {
			return llvm.Value{}, g.error(fmt.Sprintf("cannot infer right operand type: %v", err), expr.BinExpr.Pos)
		}
		left, err := g.generateExpression(expr.BinExpr.Left)
		if err != nil {
			return llvm.Value{}, err
		}
		right, err := g.generateExpression(expr.BinExpr.Right)
		if err != nil {
			return llvm.Value{}, err
		}

		resultType, err := g.typeChecker.InferType(&expr)
		if err != nil {
			return llvm.Value{}, err
		}
		isFloat := isFloatType(resultType)
		left, err = g.coerceValue(left, leftType, resultType, g.newValueName("promote"))
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), expr.BinExpr.Pos)
		}
		right, err = g.coerceValue(right, rightType, resultType, g.newValueName("promote"))
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), expr.BinExpr.Pos)
		}

		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			if isFloat {
				return g.builder.CreateFAdd(left, right, g.newValueName("fadd")), nil
			}
			return g.builder.CreateAdd(left, right, g.newValueName("add")), nil
		case types.BinOpSub:
			if isFloat {
				return g.builder.CreateFSub(left, right, g.newValueName("fsub")), nil
			}
			return g.builder.CreateSub(left, right, g.newValueName("sub")), nil
		case types.BinOpMul:
			if isFloat {
				return g.builder.CreateFMul(left, right, g.newValueName("fmul")), nil
			}
			return g.builder.CreateMul(left, right, g.newValueName("mul")), nil
		case types.BinOpDiv:
			if isFloat {
				return g.builder.CreateFDiv(left, right, g.newValueName("fdiv")), nil
			}
			return g.builder.CreateSDiv(left, right, g.newValueName("sdiv")), nil
		default:
			return llvm.Value{}, g.error(fmt.Sprintf("unknown operator: %s", expr.BinExpr.Op), expr.BinExpr.Pos)
		}
	case expr.Ident != nil:
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return llvm.Value{}, g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.isConst {
			return sym.llvmValue, nil
		}
		return g.builder.CreateLoad(g.llvmStorageType(sym.symbolType), sym.llvmValue, g.newValueName("load")), nil
	case expr.FuncCall != nil:
		return g.generateFuncCall(expr.FuncCall)
	case expr.FuncDecl != nil:
		fnType, err := g.funcDeclType(expr.FuncDecl)
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), expr.Pos)
		}
		name := fmt.Sprintf("func_%d", g.funcCount)
		g.funcCount++
		fnValue := llvm.AddFunction(g.module, name, g.llvmType(fnType))
		if err := g.emitFuncBody(expr.FuncDecl, fnType, fnValue, ""); err != nil {
			return llvm.Value{}, err
		}
		return fnValue, nil
	default:
		return llvm.Value{}, g.error("unknown expression type", expr.Pos)
	}
}

func (g *Generator) isStatic(expr ast.NodeExpression) bool {
	if expr.IntLit != nil {
		return true
	}
	if expr.Ident != nil {
		sym, exists := g.scope.lookup(*expr.Ident)
		return exists && sym.isConst && sym.staticValue != nil
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
		intValue, ok := sym.staticValue.(*int)
		if !ok {
			return 0, g.error(fmt.Sprintf("variable is not an integer constant: %s", *expr.Ident), expr.Pos)
		}
		return *intValue, nil
	}
	if expr.BinExpr != nil {
		return g.foldBinExpr(expr.BinExpr)
	}
	return 0, g.error("unknown expression type", expr.Pos)
}
