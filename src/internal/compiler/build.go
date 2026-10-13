package compiler

import (
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
	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
	"github.com/mohae/randchars"
)

type BuildOptions struct {
	Input              []string
	Output             string
	SaveIntermediaries bool
	Target             string
	Opt                int
}

func Build(opts BuildOptions) error {
	if err := prepareOutputDir(opts.Output); err != nil {
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

	linker.Link(objects, opts.Output, opts.SaveIntermediaries, target.TargetTriples[opts.Target])

	return nil
}

func validateTarget(host, target string) error {
	// Windows can only be targeted from Windows
	if strings.HasPrefix(target, "windows/") &&
		host != "windows" {
		return fmt.Errorf("cross-compiling to Windows is not currently supported")
	}
	return nil
}

func prepareOutputDir(output string) error {
	outDir := filepath.Dir(output)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("Failed to create output directory %s: %v\n", outDir, err)
		}
	}
	return nil
}

func validateInputs(inputs []string) error {
	for _, input := range inputs {
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return fmt.Errorf("Failed to read input file %s: %v\n", input, err)
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

	results := make(chan result, len(opts.Input))
	var wg sync.WaitGroup

	for i := range opts.Input {
		wg.Add(1)
		go func(input string, index int) {
			defer wg.Done()

			obj, err := compileFile(input, opts.Output, opts.Target, opts.SaveIntermediaries, index, opts.Opt)

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

	errCounter := 0

	// Collect results and handle errors
	for r := range results {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, r.err)
			errCounter++
		} else {
			objects[r.index] = r.outputFile
		}
	}

	if errCounter > 0 {
		return nil, fmt.Errorf("%d error(s) occurred during compilation", errCounter)
	}

	return objects, nil
}

func compileFile(input, outputPrefix, targetTriple string, saveIntermediaries bool, idx, opt int) (string, error) {
	content, err := os.ReadFile(input)
	if err != nil {
		return "", fmt.Errorf("Failed to read input file %s: %v\n", input, err)
	}

	lexer := lexer.NewLexer(string(content), input)
	output := make(chan types.LexerResult, 4096)
	go lexer.Tokenize(output)

	parser := parser.NewParser(filepath.Base(input))
	program, err := parser.Parse(output)
	if err != nil {
		return "", err
	}

	tc := typesys.NewTypeChecker(input)
	if err := tc.CheckProgram(&program); err != nil {
		return "", err
	}

	outputFile := outputPrefix + "_" + string(randchars.LowerAlpha(8))

	generator := codegen.NewGenerator(&program, input)
	targetTriple, ok := target.TargetTriples[targetTriple]
	if !ok {
		return "", fmt.Errorf("unsupported target: %s", targetTriple)
	}
	ir, err := generator.Generate(targetTriple)

	if err != nil {
		return "", err
	}

	if err := os.WriteFile(outputFile+".ll", []byte(ir), 0644); err != nil {
		return "", fmt.Errorf("failed to write LLVM IR to file for %s: %w", input, err)
	}

	if opt > 0 {
		if err := optimizer.Optimize(outputFile+".ll", outputFile+".ll", opt); err != nil {
			return "", err
		}
	}

	if err := assembler.Assemble(outputFile+".ll", outputFile+".o", outputPrefix, saveIntermediaries, idx); err != nil {
		return "", fmt.Errorf("assemble failed for %s: %w", input, err)
	}

	return outputFile + ".o", nil
}
