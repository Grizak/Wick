package codegen

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
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
	funcCount         int
	blockCount        int
	valueCount        int
}

func NewGenerator(root *ast.NodeProgram, filename string) *Generator {
	return &Generator{
		root:        root,
		fileName:    filename,
		typeChecker: typesys.NewTypeChecker(filename),
		scope:       NewScope(nil),
	}
}

func (g *Generator) Generate(mod llvm.Module, target target.Target, builder llvm.Builder, ctx llvm.Context) error {
	g.module = mod
	g.target = target
	g.builder = builder
	g.context = ctx
	g.globalScope = g.scope

	voidType := ctx.VoidType()
	entryType := llvm.FunctionType(voidType, nil, false)
	entry := llvm.AddFunction(mod, target.EntryPoint(), entryType)
	entryBlock := ctx.AddBasicBlock(entry, "entry")

	exitFn, err := g.createExitFunction()
	if err != nil {
		return err
	}
	exitType := &typesys.FunctionType{
		ParamTypes: []types.Type{&typesys.Int64Type{}},
		ReturnType: &typesys.VoidType{},
		Defaults:   []*ast.NodeExpression{nil},
	}
	if err := g.scope.declare("exit", Symbol{
		llvmValue:  exitFn,
		isConst:    true,
		symbolType: exitType,
	}); err != nil {
		return err
	}
	g.typeChecker.Environment().Define("exit", exitType)

	builder.SetInsertPointAtEnd(entryBlock)
	for _, statement := range g.root.Statements {
		if g.currentBlockTerminated() {
			break
		}
		if err := g.generateStatement(&statement); err != nil {
			return err
		}
	}
	if !g.currentBlockTerminated() {
		builder.CreateRetVoid()
	}

	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return err
	}
	return nil
}

func (g *Generator) createExitFunction() (llvm.Value, error) {
	i32 := g.context.Int32Type()
	fnType := llvm.FunctionType(g.context.VoidType(), []llvm.Type{i32}, false)
	exitFn := llvm.AddFunction(g.module, "exit", fnType)
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
		File: g.fileName,
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
