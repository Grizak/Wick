package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

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

	objects, err := compileInputs(opts)
	if err != nil {
		return err
	}

	linker.Link(objects, opts.Output, opts.SaveIntermediaries, target.TargetTriples[opts.Target], opts.Ldflags)

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

	// Windows can only be targeted from Windows.
	if strings.HasPrefix(targetName, "windows/") && host != "windows" {
		return fmt.Errorf("cross-compiling to Windows is not currently supported")
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

func compileInputs(opts BuildOptions) ([]string, error) {
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
	var wg sync.WaitGroup

	for i := range opts.Input {
		wg.Add(1)
		go func(input string, index int) {
			defer wg.Done()

			obj, err := compileFile(input, opts.Output, targetTriple, opts.SaveIntermediaries, index, opts.Opt)

			if err != nil {
				results <- result{err: err, index: index}
				return
			}

			results <- result{outputFile: obj, index: index}
		}(opts.Input[i], i)
	}

	// Close results once all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

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
		return nil, fmt.Errorf("compilation failed:\n%w", err)
	}

	return objects, nil
}

func compileFile(input, outputPrefix, targetTriple string, saveIntermediaries bool, idx, opt int) (string, error) {
	context := llvm.NewContext()
	defer context.Dispose()

	llvmTarget, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		return "", err
	}
	targetMachine := llvmTarget.CreateTargetMachine(
		targetTriple, "generic", "",
		llvm.CodeGenLevelDefault, llvm.RelocDefault, llvm.CodeModelDefault,
	)
	defer targetMachine.Dispose()

	content, err := os.ReadFile(input)
	if err != nil {
		return "", fmt.Errorf("Failed to read input file %s: %v\n", input, err)
	}

	lexer := lexer.NewLexer(string(content), input)
	output := make(chan types.LexerResult, 4096)
	go lexer.Tokenize(output)

	parser := parser.NewParser(input)
	program, err := parser.Parse(output)
	if err != nil {
		return "", err
	}

	sa := analysis.NewAnalyzer(input)
	if err := sa.CheckProgram(&program); err != nil {
		return "", err
	}

	generator := codegen.NewGenerator(&program, input)

	mod := context.NewModule(input)
	defer mod.Dispose()

	target, err := target.NewTarget(targetTriple)
	if err != nil {
		return "", fmt.Errorf("failed to create target for %s: %w", input, err)
	}

	mod.SetDataLayout(target.DataLayout())
	mod.SetTarget(targetTriple)

	builder := context.NewBuilder()
	defer builder.Dispose()

	err = generator.Generate(mod, target, builder, context)

	if err != nil {
		return "", err
	}

	if opt > 0 {
		if err := optimizer.Optimize(&mod, opt, targetMachine); err != nil {
			return "", err
		}
	}

	objFile, err := assembler.Assemble(&mod, outputPrefix, saveIntermediaries, idx, targetMachine)

	if err != nil {
		return "", fmt.Errorf("assemble failed for %s: %w", input, err)
	}

	return objFile, nil
}
