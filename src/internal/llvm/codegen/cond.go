package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
	"tinygo.org/x/go-llvm"
)

func (g *Generator) generateCondition(expr ast.NodeExpression) (llvm.Value, error) {
	if expr.BinExpr != nil {
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
		operandType := g.commonNumericType(leftType, rightType)
		isFloat := isFloatType(operandType)
		left, err = g.coerceValue(left, leftType, operandType, g.newValueName("promote"))
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), expr.BinExpr.Pos)
		}
		right, err = g.coerceValue(right, rightType, operandType, g.newValueName("promote"))
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), expr.BinExpr.Pos)
		}

		switch expr.BinExpr.Op {
		case types.BinOpEq:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatOEQ, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntEQ, left, right, g.newValueName("icmp")), nil
		case types.BinOpNotEq:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatONE, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntNE, left, right, g.newValueName("icmp")), nil
		case types.BinOpLt:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatOLT, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntSLT, left, right, g.newValueName("icmp")), nil
		case types.BinOpGt:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatOGT, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntSGT, left, right, g.newValueName("icmp")), nil
		case types.BinOpLtEq:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatOLE, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntSLE, left, right, g.newValueName("icmp")), nil
		case types.BinOpGtEq:
			if isFloat {
				return g.builder.CreateFCmp(llvm.FloatOGE, left, right, g.newValueName("fcmp")), nil
			}
			return g.builder.CreateICmp(llvm.IntSGE, left, right, g.newValueName("icmp")), nil
		default:
			return llvm.Value{}, g.error(fmt.Sprintf("not a comparison operator: %s", expr.BinExpr.Op), expr.BinExpr.Pos)
		}
	}

	if expr.BoolLit != nil || expr.Ident != nil {
		value, err := g.generateExpression(expr)
		if err != nil {
			return llvm.Value{}, err
		}
		return value, nil
	}
	return llvm.Value{}, g.error("expected a boolean expression", expr.Pos)
}
