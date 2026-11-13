package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/llvm/assembler"
	"github.com/Grizak/Wick/src/internal/llvm/codegen"
	"github.com/Grizak/Wick/src/internal/llvm/linker"
	optimizer "github.com/Grizak/Wick/src/internal/llvm/opt"
	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/semantic/analysis"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
	"tinygo.org/x/go-llvm"
)

type BuildOptions struct {
	Input              []string
	Output             string
	SaveIntermediaries bool
	Target             string
	Opt                int
	Ldflags            []string
}

func Build(opts BuildOptions) error {
	if err := validateConfig(opts); err != nil {
		return err
	}

	if err := prepareOutputDir(opts.Output); err != nil {
		return err
	}

	if err := prepareLLVM(); err != nil {
		return err
	}

	if err := validateTarget(runtime.GOOS, opts.Target); err != nil {
		return err
	}

	// Make sure that the input files exists
	if err := validateInputs(opts.Input); err != nil {
		return err
	}

	toolchain.Init()

	object, libs, err := compileInputs(opts)
	if err != nil {
		return err
	}

	linker.Link(object, opts.Output, opts.SaveIntermediaries, target.TargetTriples[opts.Target], append(opts.Ldflags, libs...))

	return nil
}

func validateConfig(opts BuildOptions) error {
	if opts.Opt > 3 || opts.Opt < 0 {
		return fmt.Errorf("invalid optimization value %d: must be between 0 and 3", opts.Opt)
	}
	return nil
}

func validateTarget(host, targetName string) error {
	if _, ok := target.TargetTriples[targetName]; !ok {
		return fmt.Errorf("unsupported target %q", targetName)
	}

	return nil
}

func prepareLLVM() error {
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllAsmParsers()
	llvm.InitializeAllAsmPrinters()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllTargets()
	return nil
}

func prepareOutputDir(output string) error {
	outDir := filepath.Dir(output)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create output directory %q: %w", outDir, err)
	}
	return nil
}

func validateInputs(inputs []string) error {
	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("stat input file %q: %w", input, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("input %q is not a regular file", input)
		}
	}
	return nil
}

func compileInputs(opts BuildOptions) (string, []string, error) {
	objects := make([]string, len(opts.Input))

	type result struct {
		index      int
		outputFile string
		err        error
	}

	targetTriple := target.TargetTriples[opts.Target]

	optimizer.Initialize()
	defer optimizer.Deinitialize()

	results := make(chan result, len(opts.Input))

	context := llvm.NewContext()
	defer context.Dispose()

	llvmTarget, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		return "", nil, err
	}
	targetMachine := llvmTarget.CreateTargetMachine(
		targetTriple, "generic", "",
		llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault,
	)
	defer targetMachine.Dispose()

	output := make(chan types.LexerResult, 4096)
	go func() {
		for _, input := range opts.Input {
			content, err := os.ReadFile(input)
			if err != nil {
				output <- types.LexerResult{Err: &types.CompileError{Msg: fmt.Sprintf("err: %w", fmt.Errorf("Failed to read file %s: %w", input, err))}}
			}
			lexer := lexer.NewLexer(string(content), input)
			lexer.Tokenize(output)
		}
	}()

	parser := parser.NewParser()
	program, err := parser.Parse(output, filepath.Dir(opts.Input[0]))
	if err != nil {
		return "", nil, err
	}

	sa := analysis.NewAnalyzer()
	if err := sa.CheckProgram(&program); err != nil {
		return "", nil, err
	}

	generator := codegen.NewGenerator(&program)

	mod := context.NewModule("main.wk")
	defer mod.Dispose()

	target, err := target.NewTarget(targetTriple)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create target: %w", err)
	}

	mod.SetDataLayout(target.DataLayout())
	mod.SetTarget(targetTriple)

	builder := context.NewBuilder()
	defer builder.Dispose()

	err = generator.Generate(mod, target, builder, context)

	if err != nil {
		return "", nil, err
	}

	if opts.Opt > 0 {
		if err := optimizer.Optimize(&mod, opts.Opt, targetMachine); err != nil {
			return "", nil, err
		}
	}

	objFile, err := assembler.Assemble(&mod, opts.Output, opts.SaveIntermediaries, targetMachine)

	if err != nil {
		return "", nil, fmt.Errorf("assemble failed: %w", err)
	}

	// Collect results and handle errors
	var errs = make([]error, len(opts.Input))

	for r := range results {
		if r.err != nil {
			errs[r.index] = fmt.Errorf("%s: %w", opts.Input[r.index], r.err)
		} else {
			objects[r.index] = r.outputFile
		}
	}

	if err := errors.Join(errs...); err != nil {
		return "", nil, fmt.Errorf("compilation failed:\n%w", err)
	}

	return objFile, parser.Libs, nil
}
