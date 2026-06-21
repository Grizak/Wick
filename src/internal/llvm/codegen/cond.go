package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (g *Generator) generateCondition(expr ast.NodeExpression) (string, error) {
	// If it's a bool literal
	if expr.BoolLit != nil {
		if *expr.BoolLit {
			return "1", nil
		}
		return "0", nil
	}

	// If it's a comparison expression
	if expr.BinExpr != nil {
		leftType, err := g.typeChecker.InferType(&expr.BinExpr.Left)
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer left operand type: %v", err), expr.BinExpr.Pos)
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
		isFloat := isFloatType(leftType)

		switch expr.BinExpr.Op {
		case types.BinOpEq:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp oeq double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp eq i64 %s, %s", result, left, right))
			}
		case types.BinOpNotEq:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp one double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp ne i64 %s, %s", result, left, right))
			}
		case types.BinOpLt:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp olt double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp slt i64 %s, %s", result, left, right))
			}
		case types.BinOpGt:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp ogt double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp sgt i64 %s, %s", result, left, right))
			}
		case types.BinOpLtEq:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp ole double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp sle i64 %s, %s", result, left, right))
			}
		case types.BinOpGtEq:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fcmp oge double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = icmp sge i64 %s, %s", result, left, right))
			}
		default:
			return "", g.error(fmt.Sprintf("not a comparison operator: %s", expr.BinExpr.Op), expr.BinExpr.Pos)
		}

		return result, nil
	}

	// If it's an ident of bool type
	if expr.Ident != nil {
		return g.generateExpression(expr)
	}

	return "", g.error("expected a boolean expression", expr.Pos)
}
