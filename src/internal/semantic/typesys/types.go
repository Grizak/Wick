package typesys

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

// TypeEnvironment tracks variable types
type TypeEnvironment struct {
	vars   map[string]types.Type
	parent *TypeEnvironment
}

func NewTypeEnvironment() *TypeEnvironment {
	return &TypeEnvironment{
		vars: make(map[string]types.Type),
	}
}

func (env *TypeEnvironment) Define(name string, typ types.Type) {
	env.vars[name] = typ
}

func (env *TypeEnvironment) Lookup(name string) (types.Type, bool) {
	if typ, exists := env.vars[name]; exists {
		return typ, true
	}
	if env.parent != nil {
		return env.parent.Lookup(name)
	}
	return nil, false
}

func (env *TypeEnvironment) Child() *TypeEnvironment {
	return &TypeEnvironment{
		vars:   make(map[string]types.Type),
		parent: env,
	}
}

// TypeChecker validates type correctness
type TypeChecker struct {
	env       *TypeEnvironment
	filename  string
	loopDepth int // incremented when entering a loop, decremented when leaving

}

func NewTypeChecker(filename string) *TypeChecker {
	return &TypeChecker{
		filename: filename,
		env:      NewTypeEnvironment(),
	}
}

func (tc *TypeChecker) Environment() *TypeEnvironment {
	return tc.env
}

func (tc *TypeChecker) CheckProgram(program *ast.NodeProgram) error {
	tc.filename = program.Filename
	for _, stmt := range program.Statements {
		if err := tc.CheckStatement(&stmt); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TypeChecker) CheckStatement(stmt *ast.NodeStatement) error {
	if stmt.Exit != nil {
		return tc.CheckExit(stmt.Exit)
	}
	if stmt.VarDecl != nil {
		return tc.CheckVarDecl(stmt.VarDecl)
	}
	if stmt.VarAssign != nil {
		return tc.CheckVarAssign(stmt.VarAssign)
	}
	if stmt.Block != nil {
		return tc.CheckBlock(stmt.Block)
	}
	if stmt.If != nil {
		return tc.CheckIf(stmt.If)
	}
	if stmt.For != nil {
		return tc.CheckFor(stmt.For)
	}
	if stmt.Break != nil {
		return tc.CheckBreak(stmt.Break)
	}
	if stmt.Continue != nil {
		return tc.CheckContinue(stmt.Continue)
	}
	return tc.error("unknown statement type", nil)
}

func (tc *TypeChecker) CheckBlock(block *ast.NodeBlock) error {
	tc.EnterScope()
	defer tc.ExitScope()

	for _, stmt := range block.Statements {
		if err := tc.CheckStatement(&stmt); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TypeChecker) CheckExit(exit *ast.NodeExit) error {
	_, err := tc.CheckExpression(exit.Expr)
	return err
}

func (tc *TypeChecker) CheckVarDecl(decl *ast.NodeVarDecl) error {
	exprType, err := tc.CheckExpression(decl.Expr)
	if err != nil {
		return err
	}
	if decl.Type != nil {
		// Parse the declared type string
		declaredType, err := tc.ParseTypeString(*decl.Type)
		if err != nil {
			return tc.error(fmt.Sprintf("invalid type annotation for variable '%s': %v", decl.Name, err), &decl.Pos)
		}
		if !exprType.Equals(declaredType) && !TryCoerce(exprType, declaredType) {
			return tc.error(fmt.Sprintf("type mismatch for variable '%s': declared as %s but got %s", decl.Name, declaredType.Name(), exprType.Name()), &decl.Pos)
		}
		tc.env.Define(decl.Name, declaredType)
	} else {
		tc.env.Define(decl.Name, exprType)
	}
	return nil
}

func (tc *TypeChecker) CheckVarAssign(assign *ast.NodeVarAssign) error {
	exprType, err := tc.CheckExpression(assign.Expr)
	if err != nil {
		return err
	}
	varType, exists := tc.env.Lookup(assign.Name)
	if !exists {
		return tc.error(fmt.Sprintf("undeclared variable: %s", assign.Name), &assign.Pos)
	}
	if !exprType.Equals(varType) {
		return tc.error(fmt.Sprintf("type mismatch in assignment to variable '%s': expected %s but got %s", assign.Name, varType.Name(), exprType.Name()), &assign.Pos)
	}
	return nil
}

func (tc *TypeChecker) CheckExpression(expr ast.NodeExpression) (types.Type, error) {
	return tc.InferType(&expr)
}

func (tc *TypeChecker) CheckBinaryOp(expr *ast.NodeBinExpr, left, right types.Type) (types.Type, error) {
	// Arithmetic operators
	if expr.Op == "+" || expr.Op == "-" || expr.Op == "*" || expr.Op == "/" {
		if !isNumeric(left) || !isNumeric(right) {
			return nil, tc.error(fmt.Sprintf("arithmetic operator '%s' requires numeric types, got %s and %s", expr.Op, left.Name(), right.Name()), &expr.Pos)
		}
		// Type promotion: if either is float, result is float
		if isFloat(left) || isFloat(right) {
			return &FloatType{}, nil
		}
		return left, nil
	}

	// Comparison operators
	if expr.Op == "==" || expr.Op == "!=" {
		if !left.Equals(right) {
			return nil, tc.error(fmt.Sprintf("comparison operator '%s' requires same types, got %s and %s", expr.Op, left.Name(), right.Name()), &expr.Pos)
		}
		return &BoolType{}, nil
	}

	if expr.Op == "<" || expr.Op == ">" || expr.Op == "<=" || expr.Op == ">=" {
		if !isNumeric(left) || !isNumeric(right) {
			return nil, tc.error(fmt.Sprintf("relational operator '%s' requires numeric types, got %s and %s", expr.Op, left.Name(), right.Name()), &expr.Pos)
		}
		return &BoolType{}, nil
	}

	// Logical operators
	if expr.Op == "&&" || expr.Op == "||" {
		if !isBool(left) || !isBool(right) {
			return nil, tc.error(fmt.Sprintf("logical operator '%s' requires bool types, got %s and %s", expr.Op, left.Name(), right.Name()), &expr.Pos)
		}
		return &BoolType{}, nil
	}

	return nil, tc.error(fmt.Sprintf("unknown operator: %s", expr.Op), &expr.Pos)
}

func isNumeric(t types.Type) bool {
	switch t.(type) {
	case *Int64Type, *Int32Type, *FloatType:
		return true
	}
	return false
}

func isFloat(t types.Type) bool {
	_, ok := t.(*FloatType)
	return ok
}

func isBool(t types.Type) bool {
	_, ok := t.(*BoolType)
	return ok
}

func (tc *TypeChecker) error(msg string, pos *types.Position) *types.CompileError {
	return &types.CompileError{
		File: tc.filename,
		Pos:  pos,
		Msg:  msg,
	}
}

func (tc *TypeChecker) InferType(expr *ast.NodeExpression) (types.Type, error) {
	if expr.IntLit != nil {
		return &Int64Type{}, nil
	}
	if expr.FloatLit != nil {
		return &FloatType{}, nil
	}
	if expr.BoolLit != nil {
		return &BoolType{}, nil
	}
	if expr.StringLit != nil {
		return &StringType{}, nil
	}
	if expr.Ident != nil {
		typ, exists := tc.env.Lookup(*expr.Ident)
		if !exists {
			return nil, tc.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), &expr.Pos)
		}
		return typ, nil
	}
	if expr.FuncDecl != nil {
		paramTypes := make([]types.Type, len(expr.FuncDecl.Params))
		defaults := make([]*ast.NodeExpression, len(expr.FuncDecl.Params))
		seenDefault := false

		for i, p := range expr.FuncDecl.Params {
			if p.Default == nil {
				if seenDefault {
					return nil, tc.error(fmt.Sprintf(
						"parameter '%s' without a default cannot follow a defaulted parameter", p.Name), &p.Pos)
				}
				if p.Type == nil {
					return nil, tc.error(fmt.Sprintf(
						"parameter '%s' requires an explicit type or a default value", p.Name), &p.Pos)
				}
				t, err := tc.ParseTypeString(*p.Type)
				if err != nil {
					return nil, err
				}
				paramTypes[i] = t
				continue
			}

			seenDefault = true
			defaultType, err := tc.InferType(p.Default)
			if err != nil {
				return nil, err
			}
			if p.Type != nil {
				declaredType, err := tc.ParseTypeString(*p.Type)
				if err != nil {
					return nil, err
				}
				if !declaredType.Equals(defaultType) && !TryCoerce(defaultType, declaredType) {
					return nil, tc.error(fmt.Sprintf(
						"parameter '%s': default value type %s does not match declared type %s",
						p.Name, defaultType.Name(), declaredType.Name()), &p.Pos)
				}
				paramTypes[i] = declaredType
			} else {
				paramTypes[i] = defaultType // e.g. `a = 5` infers i64 from the literal
			}
			defaults[i] = p.Default
		}

		if expr.FuncDecl.ReturnType == nil {
			return nil, fmt.Errorf("function literal requires an explicit return type")
		}
		retType, err := tc.ParseTypeString(*expr.FuncDecl.ReturnType)
		if err != nil {
			return nil, err
		}
		return &FunctionType{ParamTypes: paramTypes, ReturnType: retType, Defaults: defaults}, nil
	}
	if expr.FuncCall != nil {
		return tc.CheckFunctionCall(expr.FuncCall)
	}
	if expr.BinExpr != nil {
		leftType, err := tc.InferType(&expr.BinExpr.Left)
		if err != nil {
			return nil, err
		}
		rightType, err := tc.InferType(&expr.BinExpr.Right)
		if err != nil {
			return nil, err
		}
		return tc.CheckBinaryOp(expr.BinExpr, leftType, rightType)
	}
	return nil, tc.error("unknown expression type", &expr.Pos)
}

func TryCoerce(from, to types.Type) bool {
	if from.Equals(to) {
		return true
	}
	if isNumeric(from) && isNumeric(to) {
		// Allow coercion between numeric types (e.g. int to float)
		return true
	}
	return false
}

// ParseTypeString converts a type annotation string into a Type object
// Supports: i32, i64, f64, bool, string, int, float, *Type, [N]Type
func (tc *TypeChecker) ParseTypeString(typeStr string) (types.Type, error) {
	// Handle pointer types
	if len(typeStr) > 0 && typeStr[0] == '*' {
		innerType, err := tc.ParseTypeString(typeStr[1:])
		if err != nil {
			return nil, err
		}
		return &PointerType{PointsTo: innerType}, nil
	}

	// Handle array types [N]Type
	if len(typeStr) > 0 && typeStr[0] == '[' {
		endIdx := -1
		for i, c := range typeStr {
			if c == ']' {
				endIdx = i
				break
			}
		}
		if endIdx == -1 {
			return nil, fmt.Errorf("invalid array type syntax: %s", typeStr)
		}

		lengthStr := typeStr[1:endIdx]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array length: %s", lengthStr)
		}

		innerTypeStr := typeStr[endIdx+1:]
		innerType, err := tc.ParseTypeString(innerTypeStr)
		if err != nil {
			return nil, err
		}
		return &ArrayType{ElementType: innerType, Length: length}, nil
	}

	// Handle base types
	switch typeStr {
	case "i32":
		return &Int32Type{}, nil
	case "i64", "int":
		return &Int64Type{}, nil
	case "f64", "float":
		return &FloatType{}, nil
	case "bool":
		return &BoolType{}, nil
	case "string":
		return &StringType{}, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", typeStr)
	}
}

// CheckFunctionCall validates a function call and returns the return type
func (tc *TypeChecker) CheckFunctionCall(call *ast.NodeFuncCall) (types.Type, error) {
	// Look up the function type
	fnType, exists := tc.env.Lookup(call.Name)
	if !exists {
		return nil, tc.error(fmt.Sprintf("undeclared function: %s", call.Name), &call.Pos)
	}

	// Verify it's actually a function type
	fn, ok := fnType.(*FunctionType)
	if !ok {
		return nil, tc.error(fmt.Sprintf("%s is not a function", call.Name), &call.Pos)
	}

	// Check argument count
	required := fn.RequiredArgCount()
	if len(call.Args) < required || len(call.Args) > len(fn.ParamTypes) {
		return nil, tc.error(fmt.Sprintf("function %s expects %d to %d arguments, got %d",
			call.Name, required, len(fn.ParamTypes), len(call.Args)), &call.Pos)
	}

	// Check each argument type
	for i, arg := range call.Args {
		argType, err := tc.CheckExpression(arg)
		if err != nil {
			return nil, err
		}

		// Allow coercion if argument type doesn't match exactly
		if !argType.Equals(fn.ParamTypes[i]) && !TryCoerce(argType, fn.ParamTypes[i]) {
			return nil, tc.error(fmt.Sprintf("function %s argument %d: expected %s, got %s",
				call.Name, i+1, fn.ParamTypes[i].Name(), argType.Name()), &call.Pos)
		}
	}

	return fn.ReturnType, nil
}

// EnterScope creates a new child scope for variables
func (tc *TypeChecker) EnterScope() {
	tc.env = tc.env.Child()
}

// ExitScope returns to the parent scope
func (tc *TypeChecker) ExitScope() {
	if tc.env.parent != nil {
		tc.env = tc.env.parent
	}
}

func (tc *TypeChecker) CheckIf(node *ast.NodeIf) error {
	// Check condition is a bool expression
	condType, err := tc.CheckExpression(node.Condition)
	if err != nil {
		return err
	}
	if !isBool(condType) {
		return tc.error(fmt.Sprintf("if condition must be a bool, got %s", condType.Name()), &node.Pos)
	}

	// Check then block
	if err := tc.CheckBlock(&node.Then); err != nil {
		return err
	}

	// Check else block if present
	if node.Else != nil {
		if err := tc.CheckBlock(node.Else); err != nil {
			return err
		}
	}

	return nil
}

func (tc *TypeChecker) CheckFor(node *ast.NodeFor) error {
	tc.EnterScope()
	tc.loopDepth++
	defer func() {
		tc.ExitScope()
		tc.loopDepth--
	}()

	// Check init statement
	if node.Init != nil {
		if err := tc.CheckStatement(node.Init); err != nil {
			return err
		}
	}

	// Check condition is bool
	if node.Condition != nil {
		condType, err := tc.CheckExpression(*node.Condition)
		if err != nil {
			return err
		}
		if !isBool(condType) {
			return tc.error(fmt.Sprintf("for condition must be a bool, got %s", condType.Name()), &node.Pos)
		}
	}

	// Check post statement
	if node.Post != nil {
		if err := tc.CheckStatement(node.Post); err != nil {
			return err
		}
	}

	// Check body — don't enter a new scope here since we already entered one
	// to include the init variable
	for _, stmt := range node.Body.Statements {
		if err := tc.CheckStatement(&stmt); err != nil {
			return err
		}
	}

	return nil
}

func (tc *TypeChecker) CheckBreak(node *ast.NodeBreak) error {
	if tc.loopDepth == 0 {
		return tc.error("break outside of loop", &node.Pos)
	}
	return nil
}

func (tc *TypeChecker) CheckContinue(node *ast.NodeContinue) error {
	if tc.loopDepth == 0 {
		return tc.error("continue outside of loop", &node.Pos)
	}
	return nil
}
