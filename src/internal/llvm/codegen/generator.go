package codegen

import (
	"fmt"
	"strconv"
	"strings"

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
	llvmName    string
	isConst     bool
	symbolType  types.Type // The actual type of this symbol
	staticValue any        // Can be *int or *float64
}

type Generator struct {
	root              *ast.NodeProgram
	output            *strings.Builder
	functions         strings.Builder
	funcCount         int
	tmpCount          int
	fileName          string
	typeChecker       *typesys.TypeChecker
	target            *target.Target
	scope             *Scope
	globalScope       *Scope
	currentFuncReturn types.Type
	loopEndLabel      string
	loopNextLabel     string
}

func NewGenerator(root *ast.NodeProgram, filename string) *Generator {
	return &Generator{
		root:        root,
		fileName:    filename,
		output:      &strings.Builder{},
		typeChecker: typesys.NewTypeChecker(filename),
		scope:       NewScope(nil),
	}
}

func (g *Generator) Generate(mod llvm.Module, target target.Target, builder llvm.Builder, ctx llvm.Context) error {
	g.globalScope = g.scope
	g.target = &target

	fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{}, false)
	fn := llvm.AddFunction(mod, target.EntryPoint(), fnType)

	entry := ctx.AddBasicBlock(fn, "entry")
	builder.SetInsertPointAtEnd(entry)

	for _, statement := range g.root.Statements {
		if err := g.generateStatement(mod, builder, ctx, &statement); err != nil {
			return err
		}
	}

	// Exit func
	i32 := ctx.Int32Type()
	i64 := ctx.Int64Type()

	fnTypeExit := llvm.FunctionType(ctx.VoidType(), []llvm.Type{i32}, false)
	fnExit := llvm.AddFunction(mod, "exit", fnTypeExit)

	entryExit := ctx.AddBasicBlock(fnExit, "entry")
	builder.SetInsertPointAtEnd(entryExit)

	code := fnExit.Param(0)

	// %code64 = sext i32 %code to i64
	code64 := builder.CreateSExt(code, i64, "code64")

	// the inline-asm callee: void(i64)
	asmFnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{i64}, false)

	// this is where target.SysExit() plugs in — the asm string + constraints
	// it currently produces as a raw string in your text backend
	sysExitAsm := llvm.InlineAsm(
		asmFnType,
		target.SysExit(), // e.g. "syscall" body text
		"",               // e.g. "{rdi}" or whatever operand constraints you use
		true,             // hasSideEffects
		false,            // isAlignStack
		0,                // dialect (0 = ATT, 1 = Intel)
		false,            // canThrow
	)

	// call void asm "...", "..."(i64 %code64)
	builder.CreateCall(asmFnType, sysExitAsm, []llvm.Value{code64}, "")

	// unreachable
	builder.CreateUnreachable()

	if strings.HasSuffix(target.Triple(), "-pc-windows-msvc") {
		i32 := ctx.Int32Type()

		declTypeExitProcess := llvm.FunctionType(ctx.VoidType(), []llvm.Type{i32}, false)
		llvm.AddFunction(mod, "ExitProcess", declTypeExitProcess)
	}

	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateExit(mod llvm.Module, builder llvm.Builder, ctx llvm.Context, exit *ast.NodeExit) error {
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
		builder.CreateCall(ctx.VoidType(), "exit", []llvm.Value{ctx.Int32Type()}, "exit")
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

func (g *Generator) tmpVar() string {
	g.tmpCount++
	return fmt.Sprintf("%%tmp%d", g.tmpCount)
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
func isFloatType(t types.Type) bool {
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

func (g *Generator) newLabel(prefix string) string {
	g.tmpCount++
	return fmt.Sprintf("%s_%d", prefix, g.tmpCount)
}

// llvmStorageType returns the LLVM type to use for alloca/load/store of a
// value of type t. Function values are stored as opaque pointers; everything
// else uses its natural LLVMType().
func llvmStorageType(t types.Type) string {
	if _, ok := t.(*typesys.FunctionType); ok {
		return "ptr"
	}
	return t.LLVMType()
}
