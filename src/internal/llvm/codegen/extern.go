package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/utils"
	"tinygo.org/x/go-llvm"
)

func (g *Generator) generateExtern(extern *ast.NodeExtern) error {
	utils.Debug(fmt.Sprintf("EXTERN CALLED, name=%s", extern.Name))

	if extern == nil {
		return fmt.Errorf("nil extern declaration")
	}

	var params []ast.NodeParam
	if extern.Params != nil {
		params = *extern.Params
	}

	fd := &ast.NodeFuncDecl{
		Params:     params,
		ReturnType: extern.ReturnType,
		Pos:        extern.Pos,
	}
	fnType, err := g.funcDeclType(fd)
	if err != nil {
		return g.error(err.Error(), extern.Pos)
	}

	fnValue := llvm.AddFunction(g.module, extern.Name, g.llvmType(fnType))
	sym := Symbol{
		llvmValue:  fnValue,
		isConst:    true,
		symbolType: fnType,
	}

	if err := g.scope.declare(extern.Name, sym); err != nil {
		return g.error(err.Error(), extern.Pos)
	}
	g.typeChecker.Environment().Define(extern.Name, fnType)

	if g.currentProgram != nil {
		modulePath := programPath(g.currentProgram)
		if g.moduleFunctions[modulePath] == nil {
			g.moduleFunctions[modulePath] = make(map[string]Symbol)
		}
		g.moduleFunctions[modulePath][extern.Name] = sym
	}

	return nil
}
