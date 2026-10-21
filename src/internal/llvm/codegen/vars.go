package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
)

func (g *Generator) generateVarDecl(decl *ast.NodeVarDecl) error {
	if decl.Const && decl.Expr.FuncDecl != nil {
		return g.generateNamedFuncDecl(decl)
	}

	exprType, err := g.typeChecker.InferType(&decl.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer type: %v", err), decl.Pos)
	}

	varType := exprType
	if decl.Type != nil {
		varType, err = g.typeChecker.ParseTypeString(*decl.Type)
		if err != nil {
			return g.error(fmt.Sprintf("invalid type: %v", err), decl.Pos)
		}
		if !exprType.Equals(varType) && !typesys.TryCoerce(exprType, varType) {
			return g.error(fmt.Sprintf(
				"type mismatch for '%s': declared %s but expression is %s",
				decl.Name, varType.Name(), exprType.Name(),
			), decl.Pos)
		}
	}

	value, err := g.generateExpression(decl.Expr)
	if err != nil {
		return err
	}
	value, err = g.coerceValue(value, exprType, varType, g.newValueName("coerce"))
	if err != nil {
		return g.error(err.Error(), decl.Pos)
	}

	sym := Symbol{
		llvmValue:  value,
		isConst:    decl.Const,
		symbolType: varType,
	}
	if staticValue := g.computeStaticValue(decl.Expr); staticValue != nil {
		sym.staticValue = staticValue
	}

	if !decl.Const {
		storageType := g.llvmStorageType(varType)
		ptr := g.builder.CreateAlloca(storageType, g.newValueName(decl.Name))
		g.builder.CreateStore(value, ptr)
		sym.llvmValue = ptr
	}

	if err := g.scope.declare(decl.Name, sym); err != nil {
		return g.error(err.Error(), decl.Pos)
	}
	g.typeChecker.Environment().Define(decl.Name, varType)
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

	value, err := g.generateExpression(assign.Expr)
	if err != nil {
		return err
	}
	value, err = g.coerceValue(value, exprType, sym.symbolType, g.newValueName("coerce"))
	if err != nil {
		return g.error(err.Error(), assign.Pos)
	}

	sym.staticValue = g.computeStaticValue(assign.Expr)
	g.scope.update(assign.Name, sym)
	g.builder.CreateStore(value, sym.llvmValue)
	return nil
}

func (g *Generator) clearMutableStaticValues() {
	for scope := g.scope; scope != nil; scope = scope.parent {
		for name, sym := range scope.symbols {
			if !sym.isConst {
				sym.staticValue = nil
				scope.symbols[name] = sym
			}
		}
	}
}
