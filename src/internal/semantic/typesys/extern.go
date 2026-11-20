package typesys

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

func (tc *TypeChecker) CheckExtern(extern *ast.NodeExtern) error {
	paramTypes := make([]types.Type, 0, len(*extern.Params))
	for _, p := range *extern.Params {
		// p.Type is *string, same as in FuncDecl params.
		// If your NodeExtern param type is different, use that field.
		if p.Type == nil {
			return tc.error(fmt.Sprintf("extern parameter '%s' requires a type", p.Name), &p.Pos)
		}
		t, err := tc.ParseTypeString(*p.Type)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, t)
	}

	if extern.ReturnType == nil {
		return tc.error(fmt.Sprintf("extern function '%s' requires a return type", extern.Name), &extern.Pos)
	}
	retType, err := tc.ParseTypeString(*extern.ReturnType)
	if err != nil {
		return err
	}

	tc.env.Define(extern.Name, &FunctionType{
		ParamTypes: paramTypes,
		ReturnType: retType,
		Defaults:   make([]*ast.NodeExpression, len(paramTypes)),
	})
	return nil
}
