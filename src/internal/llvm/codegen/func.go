package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"tinygo.org/x/go-llvm"
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

	name := fmt.Sprintf("func_%d", g.funcCount)
	g.funcCount++
	fnValue := llvm.AddFunction(g.module, name, g.llvmType(fnType))
	sym := Symbol{llvmValue: fnValue, isConst: true, symbolType: fnType}
	if err := g.scope.declare(decl.Name, sym); err != nil {
		return g.error(err.Error(), decl.Pos)
	}
	g.typeChecker.Environment().Define(decl.Name, fnType)

	return g.emitFuncBody(fd, fnType, fnValue, decl.Name)
}

func (g *Generator) funcDeclType(fd *ast.NodeFuncDecl) (*typesys.FunctionType, error) {
	expr := ast.NodeExpression{FuncDecl: fd}
	t, err := g.typeChecker.InferType(&expr)
	if err != nil {
		return nil, err
	}
	fnType, ok := t.(*typesys.FunctionType)
	if !ok {
		return nil, fmt.Errorf("function expression inferred non-function type")
	}
	return fnType, nil
}

func (g *Generator) emitFuncBody(fd *ast.NodeFuncDecl, fnType *typesys.FunctionType, fnValue llvm.Value, selfName string) error {
	previousBlock := g.builder.GetInsertBlock()
	previousScope := g.scope
	previousReturn := g.currentFuncReturn
	previousEnd, previousNext := g.loopEndBlock, g.loopNextBlock

	g.scope = NewScope(g.globalScope)
	g.currentFuncReturn = fnType.ReturnType
	g.loopEndBlock, g.loopNextBlock = llvm.BasicBlock{}, llvm.BasicBlock{}
	g.typeChecker.EnterScope()

	restore := func() {
		g.builder.SetInsertPointAtEnd(previousBlock)
		g.scope = previousScope
		g.currentFuncReturn = previousReturn
		g.loopEndBlock, g.loopNextBlock = previousEnd, previousNext
		g.typeChecker.ExitScope()
	}

	entry := g.context.AddBasicBlock(fnValue, "entry")
	g.builder.SetInsertPointAtEnd(entry)
	if selfName != "" {
		if err := g.scope.declare(selfName, Symbol{
			llvmValue:  fnValue,
			isConst:    true,
			symbolType: fnType,
		}); err != nil {
			restore()
			return g.error(err.Error(), fd.Pos)
		}
		g.typeChecker.Environment().Define(selfName, fnType)
	}

	for i, param := range fd.Params {
		paramValue := fnValue.Param(i)
		paramValue.SetName(param.Name)
		sym := Symbol{symbolType: fnType.ParamTypes[i], isConst: param.Const}
		if param.Const {
			sym.llvmValue = paramValue
		} else {
			ptr := g.builder.CreateAlloca(g.llvmStorageType(fnType.ParamTypes[i]), g.newValueName("param_"+param.Name))
			g.builder.CreateStore(paramValue, ptr)
			sym.llvmValue = ptr
		}
		if err := g.scope.declare(param.Name, sym); err != nil {
			restore()
			return g.error(err.Error(), param.Pos)
		}
		g.typeChecker.Environment().Define(param.Name, fnType.ParamTypes[i])
	}

	if err := g.generateStatements(fd.Body.Statements); err != nil {
		restore()
		return err
	}
	if !g.currentBlockTerminated() {
		if _, isVoid := fnType.ReturnType.(*typesys.VoidType); isVoid {
			g.builder.CreateRetVoid()
		} else {
			g.builder.CreateUnreachable()
		}
	}
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
		g.builder.CreateRetVoid()
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

	value, err := g.generateExpression(*node.Expr)
	if err != nil {
		return err
	}
	value, err = g.coerceValue(value, exprType, g.currentFuncReturn, g.newValueName("coerce"))
	if err != nil {
		return g.error(err.Error(), node.Pos)
	}
	g.builder.CreateRet(value)
	return nil
}

func (g *Generator) generateFuncCall(call *ast.NodeFuncCall) (llvm.Value, error) {
	sym, exists := g.scope.lookup(call.Name)
	if !exists {
		return llvm.Value{}, g.error(fmt.Sprintf("undeclared function: %s", call.Name), call.Pos)
	}
	fnType, ok := sym.symbolType.(*typesys.FunctionType)
	if !ok {
		return llvm.Value{}, g.error(fmt.Sprintf("'%s' is not callable", call.Name), call.Pos)
	}

	required := fnType.RequiredArgCount()
	if len(call.Args) < required || len(call.Args) > len(fnType.ParamTypes) {
		return llvm.Value{}, g.error(fmt.Sprintf("'%s' expects %d to %d argument(s), got %d",
			call.Name, required, len(fnType.ParamTypes), len(call.Args)), call.Pos)
	}

	callArgs := make([]ast.NodeExpression, len(fnType.ParamTypes))
	copy(callArgs, call.Args)
	for i := len(call.Args); i < len(fnType.ParamTypes); i++ {
		callArgs[i] = *fnType.Defaults[i]
	}

	callee := sym.llvmValue
	fnLLVMType := g.llvmType(fnType)
	if !sym.isConst {
		callee = g.builder.CreateLoad(g.llvmStorageType(fnType), sym.llvmValue, g.newValueName("load_fn"))
	}

	args := make([]llvm.Value, len(callArgs))
	for i := range callArgs {
		argType, err := g.typeChecker.InferType(&callArgs[i])
		if err != nil {
			return llvm.Value{}, g.error(fmt.Sprintf("cannot infer argument %d type: %v", i, err), callArgs[i].Pos)
		}
		value, err := g.generateExpression(callArgs[i])
		if err != nil {
			return llvm.Value{}, err
		}
		paramType := fnType.ParamTypes[i]
		if !argType.Equals(paramType) && !typesys.TryCoerce(argType, paramType) {
			return llvm.Value{}, g.error(fmt.Sprintf("argument %d: cannot use %s as %s",
				i, argType.Name(), paramType.Name()), callArgs[i].Pos)
		}
		value, err = g.coerceValue(value, argType, paramType, g.newValueName("coerce"))
		if err != nil {
			return llvm.Value{}, g.error(err.Error(), callArgs[i].Pos)
		}
		args[i] = value
	}

	name := ""
	if _, isVoid := fnType.ReturnType.(*typesys.VoidType); !isVoid {
		name = g.newValueName("call")
	}
	return g.builder.CreateCall(fnLLVMType, callee, args, name), nil
}
