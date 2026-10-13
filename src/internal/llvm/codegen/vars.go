package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/types"
)

func (g *Generator) generateVarDecl(decl *ast.NodeVarDecl) error {
	if decl.Const && decl.Expr.FuncDecl != nil {
		return g.generateNamedFuncDecl(decl)
	}
	// Always infer the expression type first
	exprType, err := g.typeChecker.InferType(&decl.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer type: %v", err), decl.Pos)
	}

	var varType types.Type
	if decl.Type != nil {
		varType, err = g.typeChecker.ParseTypeString(*decl.Type)
		if err != nil {
			return g.error(fmt.Sprintf("invalid type: %v", err), decl.Pos)
		}
		// Validate expression type against declared type
		if !exprType.Equals(varType) && !typesys.TryCoerce(exprType, varType) {
			return g.error(fmt.Sprintf(
				"type mismatch for '%s': declared %s but expression is %s",
				decl.Name, varType.Name(), exprType.Name(),
			), decl.Pos)
		}
	} else {
		varType = exprType
	}

	expr, err := g.generateExpression(decl.Expr)
	if err != nil {
		return err
	}

	// Emit int → float coercion if needed
	if isFloatType(varType) && !isFloatType(exprType) {
		converted := g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", converted, expr))
		expr = converted
	}

	sym := Symbol{
		llvmName:   expr,
		isConst:    decl.Const,
		symbolType: varType,
	}

	if staticVal := g.computeStaticValue(decl.Expr); staticVal != nil {
		sym.staticValue = staticVal
	}

	if !decl.Const {
		ptr := fmt.Sprintf("%%var_%s_%d", decl.Name, g.tmpCount+1)
		g.tmpCount++
		llvmType := llvmStorageType(varType)
		g.writeLine(fmt.Sprintf("    %s = alloca %s", ptr, llvmType))
		g.writeLine(fmt.Sprintf("    store %s %s, ptr %s", llvmType, expr, ptr))
		sym.llvmName = ptr
	}

	g.typeChecker.Environment().Define(decl.Name, varType)
	g.scope.declare(decl.Name, sym)
	return nil
}

func (g *Generator) generateVarAssign(assign *ast.NodeVarAssign) error {
	sym, exists := g.scope.lookup(assign.Name)
	if !exists {
		return g.error(fmt.Sprintf("undeclared variable: %s", assign.Name), assign.Pos)
	}
	if sym.isConst {
		return g.error(fmt.Sprintf("cannot reassign const variable: %s", assign.Name), assign.Pos)
	}

	// Type-check the assignment
	exprType, err := g.typeChecker.InferType(&assign.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer type: %v", err), assign.Pos)
	}
	if !exprType.Equals(sym.symbolType) && !typesys.TryCoerce(exprType, sym.symbolType) {
		return g.error(fmt.Sprintf(
			"type mismatch in assignment to '%s': expected %s but got %s",
			assign.Name, sym.symbolType.Name(), exprType.Name(),
		), assign.Pos)
	}

	expr, err := g.generateExpression(assign.Expr)
	if err != nil {
		return err
	}

	// Emit int → float coercion if needed
	if isFloatType(sym.symbolType) && !isFloatType(exprType) {
		converted := g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", converted, expr))
		expr = converted
	}

	// Update static value tracking
	sym.staticValue = g.computeStaticValue(assign.Expr)
	g.scope.update(assign.Name, sym)

	llvmType := llvmStorageType(sym.symbolType)
	g.writeLine(fmt.Sprintf("    store %s %s, ptr %s", llvmType, expr, sym.llvmName))
	return nil
}
