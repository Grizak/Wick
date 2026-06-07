package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
)

type Scope struct {
	symbols map[string]Symbol
	parent  *Scope
}

type Symbol struct {
	llvmName    string
	isConst     bool
	symbolType  typesys.Type // The actual type of this symbol
	staticValue any          // Can be *int or *float64
}

type Generator struct {
	root        *ast.NodeProgram
	output      strings.Builder
	tmpCount    int
	fileName    string
	typeChecker *typesys.TypeChecker
	target      *target.Target
	scope       *Scope
}

func NewGenerator(root *ast.NodeProgram) *Generator {
	g := Generator{
		root:        root,
		tmpCount:    0,
		typeChecker: typesys.NewTypeChecker(),
		scope:       NewScope(nil),
	}

	return &g
}

func (g *Generator) Generate(fileName, triple string) (string, error) {
	g.fileName = fileName
	target, err := target.NewTarget(triple)
	if err != nil {
		return "", err
	}
	g.target = &target
	// Write some metadata about the file (based on target)
	g.writeLine(fmt.Sprintf(`target triple = "%s"`, triple))
	g.writeLine(fmt.Sprintf(`target datalayout = "%s"`, target.DataLayout()))

	g.writeLine(fmt.Sprintf(`source_filename = "%s"`, fileName))

	// Write LLVM IR module header
	g.writeLine("")
	g.writeLine(fmt.Sprintf(`define void @%s() {`, target.EntryPoint()))
	g.writeLine(`entry:`)

	for _, statement := range g.root.Statements {
		if err := g.generateStatement(&statement); err != nil {
			return "", err
		}
	}

	g.writeLine(`    ret void`)
	g.writeLine(`}`)
	g.writeLine("")

	// Exit func
	g.writeLine(`define void @exit(i32 %code) {`)
	g.writeLine("entry:")
	g.writeLine("    %code64 = sext i32 %code to i64")
	g.writeLine("    call void " + target.SysExit())
	g.writeLine("    unreachable")
	g.writeLine("}")

	if triple == "x86_64-pc-windows-msvc" || triple == "aarch64-pc-windows-msvc" {
		g.writeLine(`declare void @ExitProcess(i32)`)
	}

	return g.output.String(), nil
}

func (g *Generator) generateStatement(stmt *ast.NodeStatement) error {
	if stmt.Exit != nil {
		return g.generateExit(stmt.Exit)
	}
	if stmt.VarDecl != nil {
		return g.generateVarDecl(stmt.VarDecl)
	}
	if stmt.VarAssign != nil {
		return g.generateVarAssign(stmt.VarAssign)
	}
	if stmt.Block != nil {
		return g.generateBlock(stmt.Block)
	}
	if stmt.If != nil {
		return g.generateIf(stmt.If)
	}
	if stmt.For != nil {
		return g.generateFor(stmt.For)
	}
	return nil
}

func (g *Generator) generateExit(exit *ast.NodeExit) error {
	expr, err := g.generateExpression(exit.Expr)
	if err != nil {
		return err
	}

	// Infer the type of the exit expression
	exprType, err := g.typeChecker.InferType(&exit.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer exit expression type: %v", err), exit.Pos)
	}

	// If it's a constant int, we can use it directly
	if _, err := strconv.Atoi(expr); err == nil {
		// It's a numeric constant, use as-is but ensure it's in int32 range
		g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", expr))
		return nil
	}

	// Otherwise, convert to i32
	converted := g.tmpVar()
	if isFloatType(exprType) {
		// Convert double to i32
		g.writeLine(fmt.Sprintf("    %s = fptosi double %s to i32", converted, expr))
	} else {
		// Truncate i64 to i32
		g.writeLine(fmt.Sprintf("    %s = trunc i64 %s to i32", converted, expr))
	}
	g.writeLine(fmt.Sprintf("    call void @exit(i32 %s)", converted))
	return nil
}

func (g *Generator) writeLine(line string) {
	g.output.WriteString(line)
	g.output.WriteString("\n")
}

func (g *Generator) generateExpression(expr ast.NodeExpression) (string, error) {
	if expr.IntLit != nil {
		return strconv.Itoa(*expr.IntLit), nil
	}

	if expr.FloatLit != nil {
		// Return the float literal as a double constant
		floatStr := strconv.FormatFloat(*expr.FloatLit, 'f', -1, 64)
		return floatStr, nil
	}

	if expr.BoolLit != nil {
		// Convert bool to i1 (0 or 1)
		val := 0
		if *expr.BoolLit {
			val = 1
		}
		return strconv.Itoa(val), nil
	}

	if expr.StringLit != nil {
		// For now, strings not fully supported in code generation
		return "", g.error("string literals not yet supported in code generation", expr.Pos)
	}

	// Try to constant-fold the expression at compile time
	if staticVal := g.computeStaticValue(expr); staticVal != nil {
		return strconv.Itoa(*staticVal), nil
	}

	if expr.BinExpr != nil {
		// Determine the types of left and right operands
		leftType, err := g.typeChecker.InferType(&expr.BinExpr.Left)
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer left operand type: %v", err), expr.BinExpr.Pos)
		}

		rightType, err := g.typeChecker.InferType(&expr.BinExpr.Right)
		if err != nil {
			return "", g.error(fmt.Sprintf("cannot infer right operand type: %v", err), expr.BinExpr.Pos)
		}

		// Generate code for operands
		left, err := g.generateExpression(expr.BinExpr.Left)
		if err != nil {
			return "", err
		}
		right, err := g.generateExpression(expr.BinExpr.Right)
		if err != nil {
			return "", err
		}

		// Determine if we're doing float or int operations
		isFloat := isFloatType(leftType) || isFloatType(rightType)

		// Promote int to float if needed
		if isFloat {
			if !isFloatType(leftType) {
				promoted := g.tmpVar()
				g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", promoted, left))
				left = promoted
			}
			if !isFloatType(rightType) {
				promoted := g.tmpVar()
				g.writeLine(fmt.Sprintf("    %s = sitofp i64 %s to double", promoted, right))
				right = promoted
			}
		}

		result := g.tmpVar()

		switch expr.BinExpr.Op {
		case types.BinOpAdd:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fadd double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = add i64 %s, %s", result, left, right))
			}
		case types.BinOpSub:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fsub double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = sub i64 %s, %s", result, left, right))
			}
		case types.BinOpMul:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fmul double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = mul i64 %s, %s", result, left, right))
			}
		case types.BinOpDiv:
			if isFloat {
				g.writeLine(fmt.Sprintf("    %s = fdiv double %s, %s", result, left, right))
			} else {
				g.writeLine(fmt.Sprintf("    %s = sdiv i64 %s, %s", result, left, right))
			}
		default:
			return "", g.error(fmt.Sprintf("unknown operator: %s", expr.BinExpr.Op), expr.BinExpr.Pos)
		}

		return result, nil
	}

	if expr.Ident != nil {
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return "", g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.isConst {
			// Constants are just their value directly
			return sym.llvmName, nil
		}
		// Mutable variables need a load
		result := g.tmpVar()
		llvmType := sym.symbolType.LLVMType()
		g.writeLine(fmt.Sprintf("    %s = load %s, ptr %s", result, llvmType, sym.llvmName))
		return result, nil
	}

	if expr.FuncCall != nil {
		return "", g.error("function calls not yet supported in code generation", expr.Pos)
	}

	return "", g.error("unknown expression type", expr.Pos)
}

func (g *Generator) isStatic(expr ast.NodeExpression) bool {
	if expr.IntLit != nil {
		return true
	}
	if expr.FloatLit != nil {
		return true
	}
	if expr.Ident != nil {
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return false
		}
		return sym.isConst && sym.staticValue != nil
	}
	if expr.BinExpr != nil {
		return g.isStatic(expr.BinExpr.Left) && g.isStatic(expr.BinExpr.Right)
	}
	return false
}

func (g *Generator) foldBinExpr(expr *ast.NodeBinExpr) (int, error) {
	left, err := g.foldExpression(expr.Left)
	if err != nil {
		return 0, err
	}
	right, err := g.foldExpression(expr.Right)
	if err != nil {
		return 0, err
	}

	switch expr.Op {
	case types.BinOpAdd:
		return left + right, nil
	case types.BinOpSub:
		return left - right, nil
	case types.BinOpMul:
		return left * right, nil
	case types.BinOpDiv:
		if right == 0 {
			return 0, g.error("division by zero in constant expression", expr.Pos)
		}
		return left / right, nil
	default:
		return 0, g.error(fmt.Sprintf("unknown operator: %s", expr.Op), expr.Pos)
	}
}

func (g *Generator) foldExpression(expr ast.NodeExpression) (int, error) {
	if expr.IntLit != nil {
		return *expr.IntLit, nil
	}
	if expr.Ident != nil {
		sym, exists := g.scope.lookup(*expr.Ident)
		if !exists {
			return 0, g.error(fmt.Sprintf("undeclared variable: %s", *expr.Ident), expr.Pos)
		}
		if sym.staticValue == nil {
			return 0, g.error(fmt.Sprintf("variable is not statically known: %s", *expr.Ident), expr.Pos)
		}
		// Type assert the staticValue to *int
		intVal, ok := sym.staticValue.(*int)
		if !ok {
			return 0, g.error(fmt.Sprintf("variable is not an integer constant: %s", *expr.Ident), expr.Pos)
		}
		return *intVal, nil
	}
	if expr.BinExpr != nil {
		return g.foldBinExpr(expr.BinExpr)
	}
	return 0, g.error("unknown expression type", expr.Pos)
}

func (g *Generator) tmpVar() string {
	g.tmpCount++
	return fmt.Sprintf("%%tmp%d", g.tmpCount)
}

func (g *Generator) generateVarDecl(decl *ast.NodeVarDecl) error {
	// Always infer the expression type first
	exprType, err := g.typeChecker.InferType(&decl.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer type: %v", err), decl.Pos)
	}

	var varType typesys.Type
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
		llvmType := varType.LLVMType()
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

	llvmType := sym.symbolType.LLVMType()
	g.writeLine(fmt.Sprintf("    store %s %s, ptr %s", llvmType, expr, sym.llvmName))
	return nil
}

func (g *Generator) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: g.fileName,
		Pos:  &pos,
		Msg:  msg,
	}
}

func (g *Generator) computeStaticValue(expr ast.NodeExpression) *int {
	if !g.isStatic(expr) {
		return nil
	}
	v, err := g.foldExpression(expr)
	if err != nil {
		return nil // folding failed, treat as non-static
	}
	return &v
}

// isFloatType checks if a type is a floating point type
func isFloatType(t typesys.Type) bool {
	_, ok := t.(*typesys.FloatType)
	return ok
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		symbols: make(map[string]Symbol),
		parent:  parent,
	}
}

func (s *Scope) lookup(name string) (Symbol, bool) {
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}
	if s.parent != nil {
		return s.parent.lookup(name)
	}
	return Symbol{}, false
}

func (s *Scope) declare(name string, sym Symbol) error {
	if _, exists := s.symbols[name]; exists {
		return fmt.Errorf("variable already declared: %s", name)
	}
	s.symbols[name] = sym
	return nil
}

func (g *Generator) enterScope() {
	g.scope = NewScope(g.scope)
}

func (g *Generator) exitScope() {
	g.scope = g.scope.parent
}

func (s *Scope) update(name string, sym Symbol) bool {
	if _, ok := s.symbols[name]; ok {
		s.symbols[name] = sym
		return true
	}
	if s.parent != nil {
		return s.parent.update(name, sym)
	}
	return false
}

func (g *Generator) generateBlock(block *ast.NodeBlock) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	for _, stmt := range block.Statements {
		g.generateStatement(&stmt)
	}
	return nil
}

func (g *Generator) generateIf(node *ast.NodeIf) error {
	// Generate the condition
	cond, err := g.generateCondition(node.Condition)
	if err != nil {
		return err
	}

	// Create unique labels
	thenLabel := g.newLabel("then")
	endLabel := g.newLabel("end")
	elseLabel := endLabel
	if node.Else != nil {
		elseLabel = g.newLabel("else")
	}

	// Conditional branch
	g.writeLine(fmt.Sprintf("    br i1 %s, label %%%s, label %%%s", cond, thenLabel, elseLabel))

	// Then block
	g.writeLine(fmt.Sprintf("%s:", thenLabel))
	g.enterScope()
	g.typeChecker.EnterScope()
	for _, stmt := range node.Then.Statements {
		if err := g.generateStatement(&stmt); err != nil {
			return err
		}
	}
	g.exitScope()
	g.typeChecker.ExitScope()
	g.writeLine(fmt.Sprintf("    br label %%%s", endLabel))

	// Else block
	if node.Else != nil {
		g.writeLine(fmt.Sprintf("%s:", elseLabel))
		g.enterScope()
		g.typeChecker.EnterScope()
		for _, stmt := range node.Else.Statements {
			if err := g.generateStatement(&stmt); err != nil {
				return err
			}
		}
		g.exitScope()
		g.typeChecker.ExitScope()
		g.writeLine(fmt.Sprintf("    br label %%%s", endLabel))
	}

	// End label
	g.writeLine(fmt.Sprintf("%s:", endLabel))
	return nil
}

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

func (g *Generator) newLabel(prefix string) string {
	g.tmpCount++
	return fmt.Sprintf("%s_%d", prefix, g.tmpCount)
}

func (g *Generator) generateFor(node *ast.NodeFor) error {
	g.enterScope()
	g.typeChecker.EnterScope()
	defer g.exitScope()
	defer g.typeChecker.ExitScope()

	// Clear static values for all mutable variables since they may change
	g.clearMutableStaticValues()

	loopLabel := g.newLabel("loop")
	bodyLabel := g.newLabel("body")
	endLabel := g.newLabel("loop_end")

	// Init statement
	if node.Init != nil {
		if err := g.generateStatement(node.Init); err != nil {
			return err
		}
	}

	// Jump to loop header
	g.writeLine(fmt.Sprintf("    br label %%%s", loopLabel))
	g.writeLine(fmt.Sprintf("%s:", loopLabel))

	// Condition check
	if node.Condition != nil {
		cond, err := g.generateCondition(*node.Condition)
		if err != nil {
			return err
		}
		g.writeLine(fmt.Sprintf("    br i1 %s, label %%%s, label %%%s", cond, bodyLabel, endLabel))
	} else {
		// Infinite loop — unconditional branch to body
		g.writeLine(fmt.Sprintf("    br label %%%s", bodyLabel))
	}

	// Body
	g.writeLine(fmt.Sprintf("%s:", bodyLabel))
	for _, stmt := range node.Body.Statements {
		if err := g.generateStatement(&stmt); err != nil {
			return err
		}
	}

	// Post statement
	if node.Post != nil {
		if err := g.generateStatement(node.Post); err != nil {
			return err
		}
	}

	// Jump back to loop header
	g.writeLine(fmt.Sprintf("    br label %%%s", loopLabel))

	// End label
	g.writeLine(fmt.Sprintf("%s:", endLabel))
	return nil
}

func (g *Generator) clearMutableStaticValues() {
	scope := g.scope
	for scope != nil {
		for name, sym := range scope.symbols {
			if !sym.isConst {
				sym.staticValue = nil
				scope.symbols[name] = sym
			}
		}
		scope = scope.parent
	}
}
