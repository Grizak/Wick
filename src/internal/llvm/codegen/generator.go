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

func (g *Generator) Generate(triple string) (string, error) {
	g.globalScope = g.scope
	target, err := target.NewTarget(triple)
	if err != nil {
		return "", err
	}
	g.target = &target
	// Write some metadata about the file (based on target)
	g.writeLine(fmt.Sprintf(`target triple = "%s"`, triple))
	g.writeLine(fmt.Sprintf(`target datalayout = "%s"`, target.DataLayout()))

	g.writeLine(fmt.Sprintf(`source_filename = "%s"`, g.fileName))

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

	if strings.HasSuffix(triple, "-pc-windows-msvc") {
		g.writeLine(`declare void @ExitProcess(i32)`)
	}

	return g.output.String() + "\n" + g.functions.String(), nil
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
