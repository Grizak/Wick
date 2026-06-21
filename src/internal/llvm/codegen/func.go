package codegen

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
)

func (g *Generator) generateNamedFuncDecl(decl *ast.NodeVarDecl) error {
	fd := decl.Expr.FuncDecl

	fnType, err := g.funcDeclType(fd)
	if err != nil {
		return g.error(err.Error(), decl.Pos)
	}
	if decl.Type != nil {
		declType, err := g.typeChecker.ParseTypeString(*decl.Type)
		if err != nil {
			return g.error(fmt.Sprintf("invalid type: %v", err), decl.Pos)
		}
		if !fnType.Equals(declType) {
			return g.error(fmt.Sprintf("type mismatch for '%s'", decl.Name), decl.Pos)
		}
	}

	name := fmt.Sprintf("@func_%d", g.funcCount)
	g.funcCount++

	// Pre-register so the body can call itself by name
	sym := Symbol{llvmName: name, isConst: true, symbolType: fnType}
	if err := g.scope.declare(decl.Name, sym); err != nil {
		return g.error(err.Error(), decl.Pos)
	}
	g.typeChecker.Environment().Define(decl.Name, fnType)

	return g.emitFuncBody(fd, fnType, name)
}

func (g *Generator) funcDeclType(fd *ast.NodeFuncDecl) (*typesys.FunctionType, error) {
	expr := ast.NodeExpression{FuncDecl: fd}
	t, err := g.typeChecker.InferType(&expr)
	if err != nil {
		return nil, err
	}
	return t.(*typesys.FunctionType), nil
}

func (g *Generator) emitFuncBody(fd *ast.NodeFuncDecl, fnType *typesys.FunctionType, name string) error {
	savedOutput := g.output
	funcBuf := &strings.Builder{}
	g.output = funcBuf

	savedScope := g.scope
	g.scope = NewScope(g.globalScope) // no enclosing locals — no closures yet

	savedRet := g.currentFuncReturn
	g.currentFuncReturn = fnType.ReturnType

	savedEnd, savedNext := g.loopEndLabel, g.loopNextLabel
	g.loopEndLabel, g.loopNextLabel = "", ""

	g.typeChecker.EnterScope() // see note below re: isolation

	restore := func() {
		g.output = savedOutput
		g.scope = savedScope
		g.currentFuncReturn = savedRet
		g.loopEndLabel, g.loopNextLabel = savedEnd, savedNext
		g.typeChecker.ExitScope()
	}

	paramStrs := make([]string, len(fd.Params))
	for i, p := range fd.Params {
		paramStrs[i] = fmt.Sprintf("%s %%arg_%s", fnType.ParamTypes[i].LLVMType(), p.Name)
	}
	retLLVM := fnType.ReturnType.LLVMType()
	g.writeLine(fmt.Sprintf("define %s %s(%s) {", retLLVM, name, strings.Join(paramStrs, ", ")))
	g.writeLine("entry:")

	for i, p := range fd.Params {
		sym := Symbol{symbolType: fnType.ParamTypes[i], isConst: p.Const}
		if p.Const {
			sym.llvmName = fmt.Sprintf("%%arg_%s", p.Name)
		} else {
			ptr := fmt.Sprintf("%%param_%s", p.Name)
			llvmType := llvmStorageType(fnType.ParamTypes[i])
			g.writeLine(fmt.Sprintf("    %s = alloca %s", ptr, llvmType))
			g.writeLine(fmt.Sprintf("    store %s %%arg_%s, ptr %s", llvmType, p.Name, ptr))
			sym.llvmName = ptr
		}
		if err := g.scope.declare(p.Name, sym); err != nil {
			restore()
			return g.error(err.Error(), p.Pos)
		}
		g.typeChecker.Environment().Define(p.Name, fnType.ParamTypes[i])
	}

	for _, stmt := range fd.Body.Statements {
		if err := g.generateStatement(&stmt); err != nil {
			restore()
			return err
		}
	}

	if _, isVoid := fnType.ReturnType.(*typesys.VoidType); isVoid {
		g.writeLine("    ret void")
	} else {
		// Type checker should guarantee all paths return; this is a
		// placeholder for now and will produce UB if actually reached.
		g.writeLine("    unreachable")
	}
	g.writeLine("}")
	g.writeLine("")

	g.functions.WriteString(funcBuf.String())
	restore()
	return nil
}

func (g *Generator) generateReturn(node *ast.NodeReturn) error {
	if g.currentFuncReturn == nil {
		return g.error("return statement outside of function", node.Pos)
	}
	if _, isVoid := g.currentFuncReturn.(*typesys.VoidType); isVoid {
		if node.Expr != nil {
			return g.error("cannot return a value from a void function", node.Pos)
		}
		g.writeLine("    ret void")
		return nil
	}
	if node.Expr == nil {
		return g.error("missing return value", node.Pos)
	}

	exprType, err := g.typeChecker.InferType(node.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer return expression type: %v", err), node.Pos)
	}
	if !exprType.Equals(g.currentFuncReturn) && !typesys.TryCoerce(exprType, g.currentFuncReturn) {
		return g.error(fmt.Sprintf("cannot return %s from function returning %s",
			exprType.Name(), g.currentFuncReturn.Name()), node.Pos)
	}

	val, err := g.generateExpression(*node.Expr)
	if err != nil {
		return err
	}
	if isFloatType(g.currentFuncReturn) && !isFloatType(exprType) {
		converted := g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", converted, val))
		val = converted
	}
	g.writeLine(fmt.Sprintf("    ret %s %s", g.currentFuncReturn.LLVMType(), val))
	return nil
}

func (g *Generator) generateFuncCall(call *ast.NodeFuncCall) (string, error) {
	sym, exists := g.scope.lookup(call.Name)
	if !exists {
		return "", g.error(fmt.Sprintf("undeclared function: %s", call.Name), call.Pos)
	}
	fnType, ok := sym.symbolType.(*typesys.FunctionType)
	if !ok {
		return "", g.error(fmt.Sprintf("'%s' is not callable", call.Name), call.Pos)
	}

	required := fnType.RequiredArgCount()
	if len(call.Args) < required || len(call.Args) > len(fnType.ParamTypes) {
		return "", g.error(fmt.Sprintf("'%s' expects %d to %d argument(s), got %d",
			call.Name, required, len(fnType.ParamTypes), len(call.Args)), call.Pos)
	}

	// Fill in any omitted trailing arguments from their defaults.
	callArgs := make([]ast.NodeExpression, len(fnType.ParamTypes))
	copy(callArgs, call.Args)
	for i := len(call.Args); i < len(fnType.ParamTypes); i++ {
		callArgs[i] = *fnType.Defaults[i]
	}

	var callee string
	if sym.isConst {
		callee = sym.llvmName
	} else {
		callee = g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = load ptr, ptr %s", callee, sym.llvmName))
	}

	argStrs := make([]string, len(callArgs))
	for i := range callArgs {
		argType, err := g.typeChecker.InferType(&callArgs[i])
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer argument %d type: %v", i, err), callArgs[i].Pos)
		}
		val, err := g.generateExpression(callArgs[i])
		if err != nil {
			return "", err
		}
		paramType := fnType.ParamTypes[i]
		if !argType.Equals(paramType) {
			if isFloatType(paramType) && !isFloatType(argType) {
				converted := g.tmpVar()
				g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", converted, val))
				val = converted
			} else if !typesys.TryCoerce(argType, paramType) {
				return "", g.error(fmt.Sprintf("argument %d: cannot use %s as %s",
					i, argType.Name(), paramType.Name()), callArgs[i].Pos)
			}
		}
		argStrs[i] = fmt.Sprintf("%s %s", paramType.LLVMType(), val)
	}
	args := strings.Join(argStrs, ", ")

	_, isVoid := fnType.ReturnType.(*typesys.VoidType)

	if sym.isConst {
		if isVoid {
			g.writeLine(fmt.Sprintf("    call void %s(%s)", callee, args))
			return "", nil
		}
		result := g.tmpVar()
		g.writeLine(fmt.Sprintf("    %s = call %s %s(%s)", result, fnType.ReturnType.LLVMType(), callee, args))
		return result, nil
	}

	if isVoid {
		g.writeLine(fmt.Sprintf("    call %s %s(%s)", fnType.LLVMType(), callee, args))
		return "", nil
	}
	result := g.tmpVar()
	g.writeLine(fmt.Sprintf("    %s = call %s %s(%s)", result, fnType.LLVMType(), callee, args))
	return result, nil
}
