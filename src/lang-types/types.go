package lang_types

import "fmt"

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
	env *TypeEnvironment
}

func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		env: NewTypeEnvironment(),
	}
}

func (tc *TypeChecker) Environment() *TypeEnvironment {
	return tc.env
}

func (tc *TypeChecker) CheckBinaryOp(op string, left, right Type) (Type, error) {
	// Arithmetic operators
	if op == "+" || op == "-" || op == "*" || op == "/" {
		if !isNumeric(left) || !isNumeric(right) {
			return nil, fmt.Errorf("arithmetic operator '%s' requires numeric types, got %s and %s",
				op, left.Name(), right.Name())
		}
		// Type promotion: if either is float, result is float
		if isFloat(left) || isFloat(right) {
			return &FloatType{}, nil
		}
		return left, nil
	}

	// Comparison operators
	if op == "==" || op == "!=" {
		if !left.Equals(right) {
			return nil, fmt.Errorf("comparison operator '%s' requires same types, got %s and %s",
				op, left.Name(), right.Name())
		}
		return &BoolType{}, nil
	}

	if op == "<" || op == ">" || op == "<=" || op == ">=" {
		if !isNumeric(left) || !isNumeric(right) {
			return nil, fmt.Errorf("relational operator '%s' requires numeric types, got %s and %s",
				op, left.Name(), right.Name())
		}
		return &BoolType{}, nil
	}

	// Logical operators
	if op == "&&" || op == "||" {
		if !isBool(left) || !isBool(right) {
			return nil, fmt.Errorf("logical operator '%s' requires bool types, got %s and %s",
				op, left.Name(), right.Name())
		}
		return &BoolType{}, nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op)
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
