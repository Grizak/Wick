package lang_types

import (
	"fmt"
	"strconv"

	"github.com/Grizak/Wick/src/types"
)

// Type represents all possible types in Wick
type Type interface {
	Name() string
	LLVMType() string
	SizeBytes() int
	Equals(other Type) bool
}

// TypeEnvironment tracks variable types
type TypeEnvironment struct {
	vars   map[string]Type
	parent *TypeEnvironment
}

func NewTypeEnvironment() *TypeEnvironment {
	return &TypeEnvironment{
		vars: make(map[string]Type),
	}
}

func (env *TypeEnvironment) Define(name string, typ Type) {
	env.vars[name] = typ
}

func (env *TypeEnvironment) Lookup(name string) (Type, bool) {
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
		vars:   make(map[string]Type),
		parent: env,
	}
}

// TypeChecker validates type correctness
type TypeChecker struct {
	env      *TypeEnvironment
	filename string
}

func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		env: NewTypeEnvironment(),
	}
}

func (tc *TypeChecker) Environment() *TypeEnvironment {
	return tc.env
}

func (tc *TypeChecker) CheckProgram(program *types.NodeProgram) error {
	tc.filename = program.Filename
	for _, stmt := range program.Statements {
		if err := tc.CheckStatement(&stmt); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TypeChecker) CheckStatement(stmt *types.NodeStatement) error {
	if stmt.Exit != nil {
		return tc.CheckExit(stmt.Exit)
	}
	if stmt.VarDecl != nil {
		return tc.CheckVarDecl(stmt.VarDecl)
	}
	if stmt.VarAssign != nil {
		return tc.CheckVarAssign(stmt.VarAssign)
	}
	return tc.error("unknown statement type", nil)
}

func (tc *TypeChecker) CheckExit(exit *types.NodeExit) error {
	_, err := tc.CheckExpression(exit.Expr)
	return err
}

func (tc *TypeChecker) CheckVarDecl(decl *types.NodeVarDecl) error {
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
		if !exprType.Equals(declaredType) && !tryCoerce(exprType, declaredType) {
			return tc.error(fmt.Sprintf("type mismatch for variable '%s': declared as %s but got %s", decl.Name, declaredType.Name(), exprType.Name()), &decl.Pos)
		}
		tc.env.Define(decl.Name, declaredType)
	} else {
		tc.env.Define(decl.Name, exprType)
	}
	return nil
}

func (tc *TypeChecker) CheckVarAssign(assign *types.NodeVarAssign) error {
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

func (tc *TypeChecker) CheckExpression(expr types.NodeExpression) (Type, error) {
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
	if expr.FuncCall != nil {
		return tc.CheckFunctionCall(expr.FuncCall)
	}
	if expr.BinExpr != nil {
		leftType, err := tc.CheckExpression(expr.BinExpr.Left)
		if err != nil {
			return nil, err
		}
		rightType, err := tc.CheckExpression(expr.BinExpr.Right)
		if err != nil {
			return nil, err
		}
		return tc.CheckBinaryOp(expr.BinExpr, leftType, rightType)
	}
	return nil, tc.error("unknown expression type", &expr.Pos)
}

func (tc *TypeChecker) CheckBinaryOp(expr *types.NodeBinExpr, left, right Type) (Type, error) {
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

func isNumeric(t Type) bool {
	switch t.(type) {
	case *Int64Type, *Int32Type, *FloatType:
		return true
	}
	return false
}

func isFloat(t Type) bool {
	_, ok := t.(*FloatType)
	return ok
}

func isBool(t Type) bool {
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

func (tc *TypeChecker) InferType(expr *types.NodeExpression) (Type, error) {
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

func tryCoerce(from, to Type) bool {
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
func (tc *TypeChecker) ParseTypeString(typeStr string) (Type, error) {
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
func (tc *TypeChecker) CheckFunctionCall(call *types.NodeFuncCall) (Type, error) {
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
	if len(call.Args) != len(fn.ParamTypes) {
		return nil, tc.error(fmt.Sprintf("function %s expects %d arguments, got %d",
			call.Name, len(fn.ParamTypes), len(call.Args)), &call.Pos)
	}

	// Check each argument type
	for i, arg := range call.Args {
		argType, err := tc.CheckExpression(arg)
		if err != nil {
			return nil, err
		}

		// Allow coercion if argument type doesn't match exactly
		if !argType.Equals(fn.ParamTypes[i]) && !tryCoerce(argType, fn.ParamTypes[i]) {
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
