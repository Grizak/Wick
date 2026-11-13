package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/types"
)

type Analyzer struct {
	programs       []*ast.NodeProgram
	programsByPath map[string]*ast.NodeProgram
	visiting       map[string]bool
	loaded         map[string]bool
	stack          []string
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) CheckProgram(prog *ast.NodeProgram) error {
	a.programs = nil
	a.programsByPath = make(map[string]*ast.NodeProgram)
	a.visiting = make(map[string]bool)
	a.loaded = make(map[string]bool)
	a.stack = nil
	if err := a.validateTopLevelImports(prog); err != nil {
		return err
	}
	rootPath, err := filepath.Abs(prog.Dir)
	if err != nil {
		return fmt.Errorf("resolve source path %q: %w", prog.Dir, err)
	}
	rootPath = filepath.Clean(rootPath)
	a.programsByPath[rootPath] = prog
	a.visiting[rootPath] = true
	a.stack = []string{rootPath}
	if err := a.resolveImports(prog, rootPath); err != nil {
		return err
	}
	a.stack = nil
	delete(a.visiting, rootPath)
	a.loaded[rootPath] = true
	a.programs = append(a.programs, prog)

	for _, module := range a.programs {
		tc := typesys.NewTypeChecker()
		if err := a.defineImportedModules(tc, module); err != nil {
			return err
		}
		if err := tc.CheckProgram(module); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) resolveImports(prog *ast.NodeProgram, currentPath string) error {
	namespaces := make(map[string]string)
	for i := range prog.Statements {
		importNode := prog.Statements[i].Import
		if importNode == nil {
			continue
		}

		resolvedPath := importNode.Module
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(filepath.Dir(currentPath), resolvedPath)
		}
		resolvedPath, err := filepath.Abs(resolvedPath)
		if err != nil {
			return a.errorAt(prog.Dir, fmt.Sprintf("cannot resolve module path %q: %v", importNode.Module, err), importNode.Pos)
		}
		resolvedPath = filepath.Clean(resolvedPath)

		namespace := moduleNamespace(resolvedPath)
		if !validIdentifier(namespace) {
			return a.errorAt(prog.Dir, fmt.Sprintf("module filename %q does not produce a valid namespace", filepath.Base(resolvedPath)), importNode.Pos)
		}
		if previousPath, exists := namespaces[namespace]; exists && previousPath != resolvedPath {
			return a.errorAt(prog.Dir, fmt.Sprintf("multiple modules use namespace %q", namespace), importNode.Pos)
		}
		namespaces[namespace] = resolvedPath

		if a.visiting[resolvedPath] {
			cycle := append(append([]string(nil), a.stack...), resolvedPath)
			return a.errorAt(prog.Dir, "cyclic module import: "+strings.Join(cycle, " -> "), importNode.Pos)
		}

		module, exists := a.programsByPath[resolvedPath]
		if !exists {
			module, err = a.parseModule(resolvedPath, prog.Dir, importNode)
			if err != nil {
				return err
			}
			a.programsByPath[resolvedPath] = module
		}
		importNode.Namespace = namespace
		importNode.ResolvedPath = resolvedPath
		importNode.Program = module

		if a.loaded[resolvedPath] {
			continue
		}
		a.visiting[resolvedPath] = true
		a.stack = append(a.stack, resolvedPath)
		if err := a.resolveImports(module, resolvedPath); err != nil {
			return err
		}
		a.stack = a.stack[:len(a.stack)-1]
		delete(a.visiting, resolvedPath)
		a.loaded[resolvedPath] = true
		a.programs = append(a.programs, module)
	}
	return nil
}

func (a *Analyzer) parseModule(path, importer string, importNode *ast.NodeImport) (*ast.NodeProgram, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, a.errorAt(importer, fmt.Sprintf("failed to read module %q: %v", importNode.Module, err), importNode.Pos)
	}

	lex := lexer.NewLexer(string(content), path)
	tokens := make(chan types.LexerResult, 4096)
	go lex.Tokenize(tokens)
	module, err := parser.NewParser().Parse(tokens, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	if err := a.validateTopLevelImports(&module); err != nil {
		return nil, err
	}

	for _, stmt := range module.Statements {
		if stmt.Import != nil || (stmt.VarDecl != nil && stmt.VarDecl.Expr.FuncDecl != nil) {
			continue
		}
		return nil, a.errorAt(path, "imported modules may only contain function declarations and imports", stmtPosition(stmt))
	}
	return &module, nil
}

func (a *Analyzer) validateTopLevelImports(program *ast.NodeProgram) error {
	return a.validateStatements(program.Statements, true, program.Dir)
}

func (a *Analyzer) validateStatements(statements []ast.NodeStatement, topLevel bool, filename string) error {
	for _, stmt := range statements {
		if stmt.Import != nil && !topLevel {
			return a.errorAt(filename, "imports are only allowed at the top level", stmt.Import.Pos)
		}
		switch {
		case stmt.VarDecl != nil:
			if err := a.validateExpressionImports(stmt.VarDecl.Expr, filename); err != nil {
				return err
			}
		case stmt.VarAssign != nil:
			if err := a.validateExpressionImports(stmt.VarAssign.Expr, filename); err != nil {
				return err
			}
		case stmt.Block != nil:
			if err := a.validateStatements(stmt.Block.Statements, false, filename); err != nil {
				return err
			}
		case stmt.If != nil:
			if err := a.validateExpressionImports(stmt.If.Condition, filename); err != nil {
				return err
			}
			if err := a.validateStatements(stmt.If.Then.Statements, false, filename); err != nil {
				return err
			}
			if stmt.If.Else != nil {
				if err := a.validateStatements(stmt.If.Else.Statements, false, filename); err != nil {
					return err
				}
			}
		case stmt.For != nil:
			if stmt.For.Init != nil {
				if err := a.validateStatements([]ast.NodeStatement{*stmt.For.Init}, false, filename); err != nil {
					return err
				}
			}
			if stmt.For.Condition != nil {
				if err := a.validateExpressionImports(*stmt.For.Condition, filename); err != nil {
					return err
				}
			}
			if stmt.For.Post != nil {
				if err := a.validateStatements([]ast.NodeStatement{*stmt.For.Post}, false, filename); err != nil {
					return err
				}
			}
			if err := a.validateStatements(stmt.For.Body.Statements, false, filename); err != nil {
				return err
			}
		case stmt.Exit != nil:
			if err := a.validateExpressionImports(stmt.Exit.Expr, filename); err != nil {
				return err
			}
		case stmt.Return != nil && stmt.Return.Expr != nil:
			if err := a.validateExpressionImports(*stmt.Return.Expr, filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Analyzer) validateExpressionImports(expr ast.NodeExpression, filename string) error {
	if expr.FuncDecl != nil {
		return a.validateStatements(expr.FuncDecl.Body.Statements, false, filename)
	}
	if expr.BinExpr != nil {
		if err := a.validateExpressionImports(expr.BinExpr.Left, filename); err != nil {
			return err
		}
		return a.validateExpressionImports(expr.BinExpr.Right, filename)
	}
	if expr.FuncCall != nil {
		for _, arg := range expr.FuncCall.Args {
			if err := a.validateExpressionImports(arg, filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Analyzer) defineImportedModules(tc *typesys.TypeChecker, prog *ast.NodeProgram) error {
	for _, stmt := range prog.Statements {
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
			functionType, err := tc.InferType(&expr)
			if err != nil {
				return err
			}
			if _, exists := functions[decl.Name]; exists {
				return a.errorAt(stmt.Import.Program.Dir, fmt.Sprintf("duplicate function declaration: %s", decl.Name), decl.Pos)
			}
			functions[decl.Name] = functionType
		}
		tc.DefineModule(stmt.Import.Namespace, functions)
	}
	return nil
}

func moduleNamespace(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func validIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return false
	}
	return true
}

func stmtPosition(stmt ast.NodeStatement) types.Position {
	switch {
	case stmt.VarDecl != nil:
		return stmt.VarDecl.Pos
	case stmt.Import != nil:
		return stmt.Import.Pos
	case stmt.Exit != nil:
		return stmt.Exit.Pos
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
	default:
		return types.Position{}
	}
}

func (a *Analyzer) errorAt(file, msg string, pos types.Position) *types.CompileError {
	return &types.CompileError{
		File: file,
		Pos:  &pos,
		Msg:  msg,
	}
}
