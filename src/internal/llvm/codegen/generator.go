package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
	"github.com/Grizak/Wick/src/internal/utils"
	"tinygo.org/x/go-llvm"
)

type Scope struct {
	symbols map[string]Symbol
	parent  *Scope
}

type Symbol struct {
	llvmValue   llvm.Value
	isConst     bool
	symbolType  types.Type
	staticValue any
}

type Generator struct {
	root              *ast.NodeProgram
	fileName          string
	typeChecker       *typesys.TypeChecker
	target            target.Target
	module            llvm.Module
	builder           llvm.Builder
	context           llvm.Context
	scope             *Scope
	globalScope       *Scope
	currentFuncReturn types.Type
	loopEndBlock      llvm.BasicBlock
	loopNextBlock     llvm.BasicBlock
	stringCount       int
	blockCount        int
	valueCount        int
	exitFunction      llvm.Value
	currentProgram    *ast.NodeProgram
	moduleFunctions   map[string]map[string]Symbol
	generatedModules  map[string]bool
	generatingModules map[string]bool
}

func NewGenerator(root *ast.NodeProgram) *Generator {
	return &Generator{
		root:              root,
		typeChecker:       typesys.NewTypeChecker(),
		scope:             NewScope(nil),
		moduleFunctions:   make(map[string]map[string]Symbol),
		generatedModules:  make(map[string]bool),
		generatingModules: make(map[string]bool),
	}
}

func (g *Generator) Generate(mod llvm.Module, target target.Target, builder llvm.Builder, ctx llvm.Context) error {
	g.module = mod
	g.target = target
	g.builder = builder
	g.context = ctx
	g.currentProgram = g.root

	if err := g.validateRootProgram(); err != nil {
		return err
	}

	exitFn, err := g.createExitFunction()
	if err != nil {
		return err
	}
	g.exitFunction = exitFn
	if err := g.generateModule(g.root); err != nil {
		return err
	}

	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		ir := mod.String()
		if err := os.WriteFile("debug.ll", []byte(ir), 0644); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (g *Generator) validateRootProgram() error {
	hasMain := false
	for _, stmt := range g.root.Statements {
		switch {
		case stmt.Import != nil || stmt.Extern != nil:
			continue
		case stmt.VarDecl != nil && stmt.VarDecl.Expr.FuncDecl != nil:
			if stmt.VarDecl.Name == "main" {
				hasMain = true
			}
		default:
			return g.error("executable statements must be inside fn main", statementPosition(&stmt))
		}
	}
	if !hasMain {
		return fmt.Errorf("root module must define fn main")
	}
	return nil
}

func statementPosition(stmt *ast.NodeStatement) types.Position {
	switch {
	case stmt.Exit != nil:
		return stmt.Exit.Pos
	case stmt.FuncCall != nil:
		return stmt.FuncCall.Pos
	case stmt.VarDecl != nil:
		return stmt.VarDecl.Pos
	case stmt.VarAssign != nil:
		return stmt.VarAssign.Pos
	case stmt.Block != nil:
		return stmt.Block.Pos
	case stmt.If != nil:
		return stmt.If.Pos
	case stmt.For != nil:
		return stmt.For.Pos
	case stmt.Break != nil:
		return stmt.Break.Pos
	case stmt.Continue != nil:
		return stmt.Continue.Pos
	case stmt.Return != nil:
		return stmt.Return.Pos
	case stmt.Import != nil:
		return stmt.Import.Pos
	case stmt.Extern != nil:
		return stmt.Extern.Pos
	default:
		return types.Position{}
	}
}

func (g *Generator) defineProgramModules(program *ast.NodeProgram) error {
	for _, stmt := range program.Statements {
		if stmt.Import == nil || stmt.Import.Program == nil {
			continue
		}
		functions := make(map[string]types.Type)
		for _, moduleStmt := range stmt.Import.Program.Statements {
			decl := moduleStmt.VarDecl
			if decl == nil || decl.Expr.FuncDecl == nil {
				continue
			}
			expr := decl.Expr
			fnType, err := g.typeChecker.InferType(&expr)
			if err != nil {
				return err
			}
			functions[decl.Name] = fnType
		}
		g.typeChecker.DefineModule(stmt.Import.Namespace, functions)
	}
	return nil
}

func (g *Generator) generateImport(importNode *ast.NodeImport) error {
	if importNode.Program == nil {
		return g.error("module import was not resolved by analysis", importNode.Pos)
	}
	return g.generateModule(importNode.Program)
}

func (g *Generator) generateModule(program *ast.NodeProgram) error {
	utils.Debug(fmt.Sprintf("generateModule called for: %s", programPath(program)))
	modulePath := programPath(program)
	if g.generatedModules[modulePath] {
		return nil
	}
	if g.generatingModules[modulePath] {
		return fmt.Errorf("cyclic module reached during code generation: %s", modulePath)
	}
	g.generatingModules[modulePath] = true
	defer delete(g.generatingModules, modulePath)

	for _, stmt := range program.Statements {
		if stmt.Import != nil {
			if err := g.generateImport(stmt.Import); err != nil {
				return err
			}
		}
	}

	previousScope := g.scope
	previousGlobalScope := g.globalScope
	previousTypeChecker := g.typeChecker
	previousFileName := g.fileName
	previousProgram := g.currentProgram
	defer func() {
		g.scope = previousScope
		g.globalScope = previousGlobalScope
		g.typeChecker = previousTypeChecker
		g.fileName = previousFileName
		g.currentProgram = previousProgram
	}()

	g.currentProgram = program
	g.scope = NewScope(nil)
	g.globalScope = g.scope
	g.typeChecker = typesys.NewTypeChecker()
	g.typeChecker.Environment().Define("exit", &typesys.FunctionType{
		ParamTypes: []types.Type{&typesys.Int64Type{}},
		ReturnType: &typesys.VoidType{},
		Defaults:   []*ast.NodeExpression{nil},
	})
	if err := g.scope.declare("exit", Symbol{
		llvmValue: g.exitFunction,
		isConst:   true,
		symbolType: &typesys.FunctionType{
			ParamTypes: []types.Type{&typesys.Int64Type{}},
			ReturnType: &typesys.VoidType{},
			Defaults:   []*ast.NodeExpression{nil},
		},
	}); err != nil {
		return err
	}
	if err := g.defineProgramModules(program); err != nil {
		return err
	}
	g.moduleFunctions[modulePath] = make(map[string]Symbol)
	for _, stmt := range program.Statements {
		utils.Debug(fmt.Sprintf("MODULE STMT: file=%s Extern=%v VarDecl=%v FuncDeclInVar=%v Import=%v\n",
			programPath(program),
			stmt.Extern != nil,
			stmt.VarDecl != nil,
			stmt.VarDecl != nil && stmt.VarDecl.Expr.FuncDecl != nil,
			stmt.Import != nil),
		)
		if stmt.Extern != nil {
			if err := g.generateExtern(stmt.Extern); err != nil {
				return err
			}
			continue
		}
		if stmt.VarDecl == nil || stmt.VarDecl.Expr.FuncDecl == nil {
			continue
		}
		if err := g.generateNamedFuncDecl(stmt.VarDecl); err != nil {
			return err
		}
		symbol, exists := g.scope.lookup(stmt.VarDecl.Name)
		if !exists {
			return g.error(fmt.Sprintf("failed to register module function: %s", stmt.VarDecl.Name), stmt.VarDecl.Pos)
		}
		g.moduleFunctions[modulePath][stmt.VarDecl.Name] = symbol
	}
	g.generatedModules[modulePath] = true
	return nil
}

func programPath(program *ast.NodeProgram) string {
	path := program.Dir
	for _, stmt := range program.Statements {
		switch {
		case stmt.Import != nil:
			if stmt.Import.Pos.Filename != "" {
				path = stmt.Import.Pos.Filename
			}
		case stmt.VarDecl != nil:
			if stmt.VarDecl.Pos.Filename != "" {
				path = stmt.VarDecl.Pos.Filename
			}
		}
		if path != program.Dir {
			break
		}
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(path)
}

func (g *Generator) createExitFunction() (llvm.Value, error) {
	i32 := g.context.Int32Type()
	fnType := llvm.FunctionType(g.context.VoidType(), []llvm.Type{i32}, false)
	exitFn := llvm.AddFunction(g.module, "exit", fnType)
	exitFn.SetLinkage(llvm.InternalLinkage)
	entry := g.context.AddBasicBlock(exitFn, "entry")
	g.builder.SetInsertPointAtEnd(entry)

	code := exitFn.Param(0)
	switch g.target.Conv() {
	case "windows/amd64", "windows/arm64":
		exitProcessType := llvm.FunctionType(g.context.VoidType(), []llvm.Type{i32}, false)
		exitProcess := g.module.NamedFunction("ExitProcess")
		if exitProcess.IsNil() {
			exitProcess = llvm.AddFunction(g.module, "ExitProcess", exitProcessType)
		}
		g.builder.CreateCall(exitProcessType, exitProcess, []llvm.Value{code}, "")
	case "linux/amd64":
		g.emitSyscallExit(code, "mov $$60, %rax; syscall", "{rdi}")
	case "linux/arm64":
		g.emitSyscallExit(code, "mov x8, #93; svc #0", "{x0}")
	case "darwin/amd64":
		g.emitSyscallExit(code, "mov $$0x2000001, %rax; syscall", "{rdi}")
	case "darwin/arm64":
		g.emitSyscallExit(code, "mov x16, #1; svc #0x80", "{x0}")
	default:
		return llvm.Value{}, fmt.Errorf("unsupported target for exit code generation: %s", g.target.Triple())
	}

	g.builder.CreateUnreachable()
	return exitFn, nil
}

func (g *Generator) emitSyscallExit(code llvm.Value, assembly, constraints string) {
	i64 := g.context.Int64Type()
	code64 := g.builder.CreateSExt(code, i64, "exit_code")
	asmType := llvm.FunctionType(g.context.VoidType(), []llvm.Type{i64}, false)
	asm := llvm.InlineAsm(asmType, assembly, constraints, true, false, llvm.InlineAsmDialectATT, false)
	g.builder.CreateCall(asmType, asm, []llvm.Value{code64}, "")
}

func (g *Generator) generateExit(exit *ast.NodeExit) error {
	value, err := g.generateExpression(exit.Expr)
	if err != nil {
		return err
	}
	exprType, err := g.typeChecker.InferType(&exit.Expr)
	if err != nil {
		return g.error(fmt.Sprintf("cannot infer exit expression type: %v", err), exit.Pos)
	}
	i32 := g.context.Int32Type()
	if isFloatType(exprType) {
		value = g.builder.CreateFPToSI(value, i32, "exit_code")
	} else if isType(exprType, &typesys.BoolType{}) {
		value = g.builder.CreateZExt(value, i32, "exit_code")
	} else if !isType(exprType, &typesys.Int32Type{}) {
		value = g.builder.CreateTrunc(value, i32, "exit_code")
	}

	exitFn := g.module.NamedFunction("exit")
	exitType := llvm.FunctionType(g.context.VoidType(), []llvm.Type{i32}, false)
	g.builder.CreateCall(exitType, exitFn, []llvm.Value{value}, "")
	g.builder.CreateUnreachable()
	return nil
}

func (g *Generator) error(msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: pos.Filename,
		Pos:  &pos,
		Msg:  msg,
	}
}

func (g *Generator) computeStaticValue(expr ast.NodeExpression) *int {
	if !g.isStatic(expr) {
		return nil
	}
	value, err := g.foldExpression(expr)
	if err != nil {
		return nil
	}
	return &value
}

func isFloatType(t types.Type) bool {
	_, ok := t.(*typesys.FloatType)
	return ok
}

func isType(t, expected types.Type) bool {
	return t != nil && t.Equals(expected)
}

func (g *Generator) llvmType(t types.Type) llvm.Type {
	switch t := t.(type) {
	case *typesys.Int64Type:
		return g.context.Int64Type()
	case *typesys.Int32Type:
		return g.context.Int32Type()
	case *typesys.FloatType:
		return g.context.DoubleType()
	case *typesys.BoolType:
		return g.context.Int1Type()
	case typesys.VoidType, *typesys.VoidType:
		return g.context.VoidType()
	case *typesys.FunctionType:
		params := make([]llvm.Type, len(t.ParamTypes))
		for i, param := range t.ParamTypes {
			params[i] = g.llvmType(param)
		}
		return llvm.FunctionType(g.llvmType(t.ReturnType), params, false)
	case *typesys.PointerType:
		return llvm.PointerType(g.llvmType(t.PointsTo), 0)
	case *typesys.ArrayType:
		return llvm.ArrayType(g.llvmType(t.ElementType), t.Length)
	case *typesys.StringType:
		return llvm.PointerType(g.context.Int8Type(), 0)
	default:
		return g.context.Int64Type()
	}
}

func (g *Generator) llvmStorageType(t types.Type) llvm.Type {
	if _, ok := t.(*typesys.FunctionType); ok {
		return llvm.PointerType(g.context.Int8Type(), 0)
	}
	return g.llvmType(t)
}

func (g *Generator) coerceValue(value llvm.Value, from, to types.Type, name string) (llvm.Value, error) {
	if from.Equals(to) {
		return value, nil
	}

	// string -> *i8 (C string)
	// A string literal already generates a *i8 pointer,
	// so there's nothing to convert — just pass it through.
	if _, isStr := from.(*typesys.StringType); isStr {
		if ptr, isPtr := to.(*typesys.PointerType); isPtr {
			if _, isInt8 := ptr.PointsTo.(*typesys.Int8Type); isInt8 {
				return value, nil
			}
		}
	}

	fromType := g.llvmType(from)
	toType := g.llvmType(to)

	if isFloatType(to) && fromType.TypeKind() == llvm.IntegerTypeKind {
		return g.builder.CreateSIToFP(value, toType, name), nil
	}
	if isFloatType(from) && toType.TypeKind() == llvm.IntegerTypeKind {
		return g.builder.CreateFPToSI(value, toType, name), nil
	}
	if fromType.TypeKind() == llvm.IntegerTypeKind && toType.TypeKind() == llvm.IntegerTypeKind {
		if fromType.IntTypeWidth() < toType.IntTypeWidth() {
			return g.builder.CreateSExt(value, toType, name), nil
		}
		if fromType.IntTypeWidth() > toType.IntTypeWidth() {
			return g.builder.CreateTrunc(value, toType, name), nil
		}
		return value, nil
	}
	return llvm.Value{}, fmt.Errorf("cannot convert %s to %s", from.Name(), to.Name())
}

func (g *Generator) commonNumericType(left, right types.Type) types.Type {
	if isFloatType(left) || isFloatType(right) {
		return &typesys.FloatType{}
	}
	if g.llvmType(left).IntTypeWidth() > g.llvmType(right).IntTypeWidth() {
		return left
	}
	return right
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

func (g *Generator) newBlock(prefix string) llvm.BasicBlock {
	g.blockCount++
	return g.context.AddBasicBlock(g.currentFunction(), fmt.Sprintf("%s_%d", prefix, g.blockCount))
}

func (g *Generator) newValueName(prefix string) string {
	g.valueCount++
	return fmt.Sprintf("%s_%d", prefix, g.valueCount)
}

func (g *Generator) currentFunction() llvm.Value {
	return g.builder.GetInsertBlock().Parent()
}

func (g *Generator) currentBlockTerminated() bool {
	block := g.builder.GetInsertBlock()
	if block.IsNil() {
		return false
	}
	last := block.LastInstruction()
	if last.IsNil() {
		return false
	}
	switch last.InstructionOpcode() {
	case llvm.Ret, llvm.Br, llvm.Unreachable:
		return true
	default:
		return false
	}
}
